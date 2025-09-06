package test

import (
	"context"
	"testing"
	"time"

	"github.com/MdSadiqMd/Exchange-Rate-Service/internal/domain"
	service "github.com/MdSadiqMd/Exchange-Rate-Service/internal/services"
	"github.com/MdSadiqMd/Exchange-Rate-Service/pkg/cache"
	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockExchangeRateAPI struct {
	mock.Mock
}

func (m *MockExchangeRateAPI) Convert(ctx context.Context, req domain.ExchangeRate) (domain.ExchangeRateResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(domain.ExchangeRateResponse), args.Error(1)
}

func (m *MockExchangeRateAPI) GetRate(ctx context.Context, from, to string, date time.Time) (domain.Money, error) {
	args := m.Called(ctx, from, to, date)
	return args.Get(0).(domain.Money), args.Error(1)
}

func TestMoneyPrecision(t *testing.T) {
	tests := []struct {
		name        string
		value       float64
		scale       int
		expectedStr string
	}{
		{
			name:        "Basic USD amount",
			value:       123.45,
			scale:       2,
			expectedStr: "123.45",
		},
		{
			name:        "High precision forex rate",
			value:       1.234567,
			scale:       6,
			expectedStr: "1.234567",
		},
		{
			name:        "Very small amount",
			value:       0.000001,
			scale:       6,
			expectedStr: "0.000001",
		},
		{
			name:        "Large amount",
			value:       1234567.89,
			scale:       2,
			expectedStr: "1234567.89",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			money := domain.NewMoney(tt.value, tt.scale)
			assert.Equal(t, tt.expectedStr, money.String())
			backToFloat := money.ToFloat()
			assert.InDelta(t, tt.value, backToFloat, 0.0000001)
		})
	}
}

func TestMoneyArithmetic(t *testing.T) {
	t.Run("Addition", func(t *testing.T) {
		m1 := domain.NewMoney(100.50, 2)
		m2 := domain.NewMoney(25.75, 2)
		result := m1.Add(m2)

		expected := domain.NewMoney(126.25, 2)
		assert.Equal(t, expected.Amount, result.Amount)
		assert.Equal(t, expected.Scale, result.Scale)
	})

	t.Run("Subtraction", func(t *testing.T) {
		m1 := domain.NewMoney(100.50, 2)
		m2 := domain.NewMoney(25.75, 2)
		result := m1.Subtract(m2)

		expected := domain.NewMoney(74.75, 2)
		assert.Equal(t, expected.Amount, result.Amount)
	})

	t.Run("Multiplication - Currency Conversion", func(t *testing.T) {
		amount := domain.NewMoney(100.0, 2)
		rate := domain.NewMoney(1.234567, 6)
		result := amount.Multiply(rate)
		assert.InDelta(t, 123.4567, result.ToFloat(), 0.000001)
	})

	t.Run("Division", func(t *testing.T) {
		dividend := domain.NewMoney(100.0, 2)
		divisor := domain.NewMoney(4.0, 2)
		result := dividend.Divide(divisor)
		assert.InDelta(t, 25.0, result.ToFloat(), 0.000001)
	})
}

func TestFloatingPointPrecisionIssues(t *testing.T) {
	t.Run("Classic 0.1 + 0.2 problem", func(t *testing.T) {
		var a, b float64 = 0.1, 0.2
		floatResult := a + b
		assert.NotEqual(t, 0.3, floatResult)

		m1 := domain.NewMoney(0.1, 6)
		m2 := domain.NewMoney(0.2, 6)
		moneyResult := m1.Add(m2)

		expected := domain.NewMoney(0.3, 6)
		assert.Equal(t, expected.Amount, moneyResult.Amount)
		assert.InDelta(t, 0.3, moneyResult.ToFloat(), 0.000001)
	})

	t.Run("Large number precision", func(t *testing.T) {
		largeAmount := domain.NewMoney(9999999999.99, 2)
		smallAmount := domain.NewMoney(0.01, 2)

		result := largeAmount.Add(smallAmount)
		expected := domain.NewMoney(10000000000.00, 2)

		assert.Equal(t, expected.Amount, result.Amount)
	})
}

func TestCurrencyConversionPrecision(t *testing.T) {
	mockAPI := &MockExchangeRateAPI{}
	c := cache.NewMemoryCache(1 * time.Hour)
	logger := log.NewNopLogger()

	svc := service.NewConversionService(logger, mockAPI, c)

	t.Run("High precision conversion", func(t *testing.T) {
		rate := domain.NewMoney(1.234567, 6)
		mockResponse := domain.ExchangeRateResponse{
			Success:   true,
			Rate:      rate,
			Amount:    rate,
			Timestamp: time.Now(),
		}

		mockAPI.On("Convert", mock.Anything, mock.Anything).Return(mockResponse, nil)

		req := &domain.ConversionRequest{
			From:   "EUR",
			To:     "USD",
			Amount: domain.NewMoney(1000.0, 2),
			Date:   time.Now().UTC(),
		}

		resp, err := svc.ConvertCurrency(context.Background(), req)

		assert.NoError(t, err)
		assert.True(t, resp.Success)

		expectedAmount := domain.NewMoney(1234.567, 6)
		assert.InDelta(t, expectedAmount.ToFloat(), resp.Result.ToFloat(), 0.000001)
	})

	t.Run("Precision maintained in cache", func(t *testing.T) {
		rate := domain.NewMoney(1.234567, 6)
		mockResponse := domain.ExchangeRateResponse{
			Success:   true,
			Rate:      rate,
			Amount:    rate,
			Timestamp: time.Now(),
		}

		mockAPI.On("Convert", mock.Anything, mock.Anything).Return(mockResponse, nil).Once()

		req := &domain.ConversionRequest{
			From:   "USD",
			To:     "EUR",
			Amount: domain.NewMoney(100.0, 2),
			Date:   time.Now().UTC(),
		}

		resp1, err1 := svc.ConvertCurrency(context.Background(), req)
		assert.NoError(t, err1)

		resp2, err2 := svc.ConvertCurrency(context.Background(), req)
		assert.NoError(t, err2)

		assert.Equal(t, resp1.Result.Amount, resp2.Result.Amount)
		assert.Equal(t, resp1.Result.Scale, resp2.Result.Scale)
	})
}
