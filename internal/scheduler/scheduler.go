package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/MdSadiqMd/Exchange-Rate-Service/internal/domain"
	service "github.com/MdSadiqMd/Exchange-Rate-Service/internal/services"
	"github.com/MdSadiqMd/Exchange-Rate-Service/pkg/cache"
)

func StartRateUpdater(ctx context.Context, svc service.ConversionService, c cache.Cache) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	updateRates(ctx, svc, c)

	for {
		select {
		case <-ticker.C:
			updateRates(ctx, svc, c)
		case <-ctx.Done():
			log.Println("Rate updater stopped")
			return
		}
	}
}

func updateRates(ctx context.Context, svc service.ConversionService, c cache.Cache) {
	log.Println("Updating rates in cache...")
	base := "USD"

	rates := map[string]float64{}
	for _, target := range domain.SupportedCurrencies {
		if target.Code == base {
			continue
		}

		req := &domain.ConversionRequest{From: base, To: target.Code, Amount: 1}
		resp, err := svc.ConvertCurrency(ctx, req)
		if err != nil {
			log.Printf("failed to fetch rate %s -> %s: %v", base, target.Code, err)
			continue
		}

		rates[target.Code] = resp.Result
		c.Set(base+"_"+target.Code, resp.Result)
	}

	for from, rateFrom := range rates {
		for to, rateTo := range rates {
			if from == to {
				continue
			}
			cross := rateTo / rateFrom
			c.Set(from+"_"+to, cross)
		}
	}

	log.Println("Rates updated successfully")
}
