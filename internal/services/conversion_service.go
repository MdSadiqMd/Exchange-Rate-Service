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
}

type conversionService struct {
	logger log.Logger
	api    external.ExchangeRateAPI
	cache  cache.Cache
}

func NewConversionService(logger log.Logger, api external.ExchangeRateAPI, c cache.Cache) ConversionService {
	return &conversionService{
		logger: logger,
		api:    api,
		cache:  c,
	}
}

func (s *conversionService) ConvertCurrency(ctx context.Context, req *domain.ConversionRequest) (*domain.ConversionResponse, error) {
	level.Info(s.logger).Log("msg", "converting currency", "from", req.From, "to", req.To, "amount", req.Amount)
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if req.Date.IsZero() {
		req.Date = time.Now().UTC()
	}

	key := fmt.Sprintf("%s:%s:%.2f:%s", req.From, req.To, req.Amount, req.Date.Format("2006-01-02"))
	cached, ok := s.cache.Get(key)
	if ok {
		level.Info(s.logger).Log("msg", "cache hit", "key", key)
		resp, ok := cached.(*domain.ConversionResponse)
		if ok {
			return resp, nil
		}
	}

	resp, err := s.api.Convert(ctx, domain.ExchangeRate{
		From:   req.From,
		To:     req.To,
		Amount: req.Amount,
		Date:   req.Date,
	})
	if err != nil {
		level.Error(s.logger).Log("msg", "conversion failed", "error", err)
		return nil, err
	}

	finalResp := &domain.ConversionResponse{
		Success: true,
		Result:  resp.Amount,
	}

	s.cache.Set(key, finalResp)
	return finalResp, nil
}
