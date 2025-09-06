package service

import (
	"context"
	"fmt"
	"time"

	"github.com/MdSadiqMd/Exchange-Rate-Service/internal/domain"
	"github.com/MdSadiqMd/Exchange-Rate-Service/pkg/cache"
	"github.com/MdSadiqMd/Exchange-Rate-Service/pkg/external"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type ConversionService interface {
	ConvertCurrency(ctx context.Context, req *domain.ConversionRequest) (*domain.ConversionResponse, error)
	GetExchangeRate(ctx context.Context, from, to string, date time.Time) (domain.Money, error)
	GetPrecisionRate(ctx context.Context, from, to string) (domain.Money, error)
}

type conversionService struct {
	logger    log.Logger
	api       external.ExchangeRateAPI
	cache     cache.Cache
	rateCache *domain.RateCache
}

func NewConversionService(logger log.Logger, api external.ExchangeRateAPI, c cache.Cache) ConversionService {
	return &conversionService{
		logger: logger,
		api:    api,
		cache:  c,
		rateCache: &domain.RateCache{
			BaseRates:   make(map[string]domain.Money),
			Adjustments: make(map[string]domain.Money),
		},
	}
}

func (s *conversionService) ConvertCurrency(ctx context.Context, req *domain.ConversionRequest) (*domain.ConversionResponse, error) {
	level.Info(s.logger).Log("msg", "converting currency", "from", req.From, "to", req.To, "amount", req.Amount.String())
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if req.Date.IsZero() {
		req.Date = time.Now().UTC()
	}

	key := fmt.Sprintf("%s:%s:%d:%d:%s",
		req.From,
		req.To,
		req.Amount.Amount,
		req.Amount.Scale,
		req.Date.Format("2006-01-02"))

	cached, ok := s.cache.Get(key)
	if ok {
		level.Info(s.logger).Log("msg", "cache hit", "key", key)
		resp, ok := cached.(*domain.ConversionResponse)
		if ok {
			return resp, nil
		}
	}

	rate, err := s.GetPrecisionRate(ctx, req.From, req.To)
	if err != nil {
		level.Error(s.logger).Log("msg", "failed to get precision rate", "error", err)
		rate, err = s.GetExchangeRate(ctx, req.From, req.To, req.Date)
		if err != nil {
			level.Error(s.logger).Log("msg", "conversion failed", "error", err)
			return nil, err
		}
	}

	result := req.Amount.Multiply(rate)

	finalResp := &domain.ConversionResponse{
		Success: true,
		Result:  result,
		Rate:    rate,
	}

	s.cache.Set(key, finalResp)
	level.Info(s.logger).Log(
		"msg", "conversion completed",
		"from", req.From,
		"to", req.To,
		"amount", req.Amount.String(),
		"rate", rate.String(),
		"result", result.String(),
	)
	return finalResp, nil
}

func (s *conversionService) GetExchangeRate(ctx context.Context, from, to string, date time.Time) (domain.Money, error) {
	rateReq := domain.ExchangeRate{
		From: from,
		To:   to,
		Rate: domain.NewMoney(1.0, domain.DefaultScale),
		Date: date,
	}

	resp, err := s.api.Convert(ctx, rateReq)
	if err != nil {
		return domain.Money{}, fmt.Errorf("failed to fetch exchange rate: %w", err)
	}
	return resp.Rate, nil
}

func (s *conversionService) GetPrecisionRate(ctx context.Context, from, to string) (domain.Money, error) {
	rate := s.rateCache.GetPrecisionRate(from, to)
	if !rate.IsZero() && !s.rateCache.IsStale(5*time.Minute) {
		return rate, nil
	}

	if from != "USD" && to != "USD" {
		crossRate := s.rateCache.CrossRate(from, to, "USD")
		if !crossRate.IsZero() {
			return crossRate, nil
		}
	}
	return s.GetExchangeRate(ctx, from, to, time.Now().UTC())
}

func (s *conversionService) UpdateRateCache(ctx context.Context) error {
	level.Info(s.logger).Log("msg", "updating rate cache")

	base := "USD"
	newRates := make(map[string]domain.Money)

	for _, target := range domain.SupportedCurrencies {
		if target.Code == base {
			continue
		}

		rate, err := s.GetExchangeRate(ctx, base, target.Code, time.Now().UTC())
		if err != nil {
			level.Error(s.logger).Log("msg", "failed to fetch rate", "pair", base+"->"+target.Code, "error", err)
			continue
		}

		key := base + ":" + target.Code
		newRates[key] = rate
		s.cache.Set("rate_"+key, rate)
	}

	for fromCurrency := range newRates {
		for toCurrency := range newRates {
			if fromCurrency == toCurrency {
				continue
			}

			from := fromCurrency[4:]
			to := toCurrency[4:]

			crossRate := s.rateCache.CrossRate(from, to, base)
			if !crossRate.IsZero() {
				crossKey := from + ":" + to
				newRates[crossKey] = crossRate
				s.cache.Set("rate_"+crossKey, crossRate)
			}
		}
	}

	s.rateCache.BaseRates = newRates
	s.rateCache.LastFetch = time.Now()

	level.Info(s.logger).Log("msg", "rate cache updated", "rates_count", len(newRates))
	return nil
}

func (s *conversionService) InterpolateRate(from, to string, timestamp time.Time) domain.Money {
	currentRate := s.rateCache.GetPrecisionRate(from, to)
	timeFactor := float64(timestamp.Hour()) / 24.0
	adjustment := currentRate.MultiplyByFloat(0.0001 * timeFactor)
	return currentRate.Add(adjustment)
}
