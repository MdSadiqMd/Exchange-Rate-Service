package test

import (
	"context"
	"testing"
	"time"

	"github.com/MdSadiqMd/Exchange-Rate-Service/internal/domain"
	service "github.com/MdSadiqMd/Exchange-Rate-Service/internal/services"
	"github.com/MdSadiqMd/Exchange-Rate-Service/pkg/cache"
	"github.com/go-kit/log"
)

type mockAPI struct {
	callCount int
}

func (m *mockAPI) Convert(ctx context.Context, req domain.ExchangeRate) (domain.ExchangeRateResponse, error) {
	m.callCount++
	return domain.ExchangeRateResponse{
		Success: true,
		Amount:  req.Amount * 0.9,
	}, nil
}

func TestConversionService_ConvertCurrency(t *testing.T) {
	logger := log.NewNopLogger()
	c := cache.NewMemoryCache(time.Hour)
	api := &mockAPI{}

	svc := service.NewConversionService(logger, api, c)
	req := &domain.ConversionRequest{
		From:   "USD",
		To:     "INR",
		Amount: 100.0,
	}

	resp, err := svc.ConvertCurrency(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}

	if resp.Result != 90.0 {
		t.Errorf("expected result 90.0, got %f", resp.Result)
	}

	if api.callCount != 1 {
		t.Errorf("expected 1 API call, got %d", api.callCount)
	}
}

func TestConversionService_UsesCache(t *testing.T) {
	logger := log.NewNopLogger()
	c := cache.NewMemoryCache(time.Hour)
	api := &mockAPI{}

	svc := service.NewConversionService(logger, api, c)
	req := &domain.ConversionRequest{
		From:   "USD",
		To:     "INR",
		Amount: 100.0,
		Date:   time.Now(),
	}

	_, err := svc.ConvertCurrency(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.ConvertCurrency(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if api.callCount != 1 {
		t.Errorf("expected API to be called once, got %d", api.callCount)
	}
}
