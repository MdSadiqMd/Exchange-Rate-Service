package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/MdSadiqMd/Exchange-Rate-Service/internal/domain"
	service "github.com/MdSadiqMd/Exchange-Rate-Service/internal/services"
	"github.com/MdSadiqMd/Exchange-Rate-Service/pkg/cache"
)

type Scheduler struct {
	conversionService service.ConversionService
	cache             cache.Cache
	baseUpdateTicker  *time.Ticker
	adjUpdateTicker   *time.Ticker
}

func NewScheduler(svc service.ConversionService, c cache.Cache) *Scheduler {
	return &Scheduler{
		conversionService: svc,
		cache:             c,
	}
}

func (s *Scheduler) StartRateUpdater(ctx context.Context) {
	s.baseUpdateTicker = time.NewTicker(1 * time.Hour)
	s.adjUpdateTicker = time.NewTicker(5 * time.Minute)

	s.updateBaseRates(ctx)
	s.updateAdjustmentRates(ctx)

	go func() {
		defer s.baseUpdateTicker.Stop()
		defer s.adjUpdateTicker.Stop()

		for {
			select {
			case <-s.baseUpdateTicker.C:
				s.updateBaseRates(ctx)
			case <-s.adjUpdateTicker.C:
				s.updateAdjustmentRates(ctx)
			case <-ctx.Done():
				log.Println("Rate updater stopped")
				return
			}
		}
	}()
}

func (s *Scheduler) updateBaseRates(ctx context.Context) {
	log.Println("Updating base rates in cache...")

	base := "USD"
	rates := make(map[string]domain.Money)

	for _, target := range domain.SupportedCurrencies {
		if target.Code == base {
			continue
		}
		rate, err := s.conversionService.GetExchangeRate(ctx, base, target.Code, time.Now().UTC())
		if err != nil {
			log.Printf("failed to fetch base rate %s -> %s: %v", base, target.Code, err)
			continue
		}

		key := base + ":" + target.Code
		rates[key] = rate

		s.cache.SetWithTTL("base_rate_"+key, rate, 1*time.Hour)
		log.Printf("Updated base rate %s: %s", key, rate.String())
	}

	s.updateCrossCurrencyRates(ctx, rates, base)
	log.Printf("Base rates updated successfully - %d rates cached", len(rates))
}

func (s *Scheduler) updateAdjustmentRates(ctx context.Context) {
	log.Println("Updating adjustment rates...")

	adjustmentCount := 0
	base := "USD"

	for _, target := range domain.SupportedCurrencies {
		if target.Code == base {
			continue
		}

		currentRate, err := s.conversionService.GetExchangeRate(ctx, base, target.Code, time.Now().UTC())
		if err != nil {
			log.Printf("failed to fetch current rate %s -> %s: %v", base, target.Code, err)
			continue
		}

		key := base + ":" + target.Code
		cachedBaseRate, ok := s.cache.Get("base_rate_" + key)
		if !ok {
			continue
		}

		baseRate, ok := cachedBaseRate.(domain.Money)
		if !ok {
			continue
		}

		adjustment := currentRate.Subtract(baseRate)

		// Only cache if adjustment is significant (> 0.01% threshold)
		threshold := baseRate.MultiplyByFloat(0.0001)
		if adjustment.Amount > threshold.Amount || adjustment.Amount < -threshold.Amount {
			s.cache.SetWithTTL("adj_rate_"+key, adjustment, 5*time.Minute)
			adjustmentCount++

			log.Printf("Updated adjustment for %s: %s (base: %s, current: %s)",
				key, adjustment.String(), baseRate.String(), currentRate.String())
		}
	}

	if adjustmentCount > 0 {
		log.Printf("Adjustment rates updated - %d adjustments cached", adjustmentCount)
	} else {
		log.Println("No significant rate adjustments found")
	}
}

func (s *Scheduler) updateCrossCurrencyRates(ctx context.Context, rates map[string]domain.Money, baseCurrency string) {
	log.Println("Calculating cross-currency rates...")

	crossRateCount := 0
	currencies := make([]string, 0, len(rates))

	for key := range rates {
		parts := key[4:]
		currencies = append(currencies, parts)
	}

	for i, fromCurrency := range currencies {
		for j, toCurrency := range currencies {
			if i == j {
				continue
			}

			fromKey := baseCurrency + ":" + fromCurrency
			toKey := baseCurrency + ":" + toCurrency

			fromRate, fromExists := rates[fromKey]
			toRate, toExists := rates[toKey]

			if !fromExists || !toExists {
				continue
			}

			crossRate := toRate.Divide(fromRate)

			crossKey := fromCurrency + ":" + toCurrency
			s.cache.SetWithTTL("cross_rate_"+crossKey, crossRate, 1*time.Hour)
			crossRateCount++
			log.Printf("Calculated cross rate %s: %s", crossKey, crossRate.String())
		}
	}

	log.Printf("Cross-currency rates updated - %d rates calculated", crossRateCount)
}

func (s *Scheduler) GetPrecisionRate(from, to string) (domain.Money, bool) {
	key := from + ":" + to
	baseRate, ok := s.cache.Get("base_rate_" + key)
	if !ok {
		crossRate, crossOk := s.cache.Get("cross_rate_" + key)
		if !crossOk {
			return domain.Money{}, false
		}
		baseRate = crossRate
	}

	baseMoney, ok := baseRate.(domain.Money)
	if !ok {
		return domain.Money{}, false
	}

	adjustment, ok := s.cache.Get("adj_rate_" + key)
	if !ok {
		return baseMoney, true
	}

	adjustmentMoney, ok := adjustment.(domain.Money)
	if !ok {
		return baseMoney, true
	}

	precisionRate := baseMoney.Add(adjustmentMoney)
	return precisionRate, true
}

func (s *Scheduler) ValidateRates(ctx context.Context) error {
	log.Println("Validating cached rates...")

	invalidCount := 0
	totalCount := 0

	for _, target := range domain.SupportedCurrencies {
		if target.Code == "USD" {
			continue
		}

		key := "USD:" + target.Code
		rate, ok := s.cache.Get("base_rate_" + key)
		if !ok {
			continue
		}

		rateMoney, ok := rate.(domain.Money)
		if !ok {
			invalidCount++
			totalCount++
			continue
		}
		if rateMoney.IsZero() || rateMoney.IsNegative() {
			log.Printf("Invalid rate detected for %s: %s", key, rateMoney.String())
			invalidCount++
		}
		totalCount++
	}

	log.Printf("Rate validation completed - %d/%d rates valid", totalCount-invalidCount, totalCount)
	if invalidCount > 0 {
		log.Printf("Warning: %d invalid rates detected", invalidCount)
	}
	return nil
}

func (s *Scheduler) Stop() {
	if s.baseUpdateTicker != nil {
		s.baseUpdateTicker.Stop()
	}
	if s.adjUpdateTicker != nil {
		s.adjUpdateTicker.Stop()
	}
	log.Println("Scheduler stopped")
}
