package external

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/MdSadiqMd/Exchange-Rate-Service/internal/domain"
)

type ExchangeRateAPI interface {
	Convert(ctx context.Context, req domain.ExchangeRate) (domain.ExchangeRateResponse, error)
	GetRate(ctx context.Context, from, to string, date time.Time) (domain.Money, error)
}

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Convert(ctx context.Context, params domain.ExchangeRate) (domain.ExchangeRateResponse, error) {
	date := params.Date
	dateStr := ""
	if !date.IsZero() {
		dateStr = date.Format("2006-01-02")
	}

	q := url.Values{}
	q.Set("access_key", c.apiKey)
	q.Set("from", params.From)
	q.Set("to", params.To)
	q.Set("amount", params.Rate.String())
	if dateStr != "" {
		q.Set("date", dateStr)
	}

	fullURL := fmt.Sprintf("%s/convert?%s", c.baseURL, q.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return domain.ExchangeRateResponse{Success: false}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return domain.ExchangeRateResponse{Success: false}, err
	}
	defer resp.Body.Close()

	var apiResp struct {
		Success bool    `json:"success"`
		Result  float64 `json:"result"`
		Info    struct {
			Rate      float64 `json:"rate"`
			Timestamp int64   `json:"timestamp"`
		} `json:"info,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return domain.ExchangeRateResponse{Success: false}, err
	}
	if !apiResp.Success {
		return domain.ExchangeRateResponse{Success: false}, fmt.Errorf("conversion failed")
	}

	resultMoney := domain.NewMoney(apiResp.Result, domain.DefaultScale)
	var rateMoney domain.Money
	if apiResp.Info.Rate > 0 {
		rateMoney = domain.NewMoney(apiResp.Info.Rate, domain.DefaultScale)
	} else {
		rateMoney = resultMoney
	}

	timestamp := time.Now()
	if apiResp.Info.Timestamp > 0 {
		timestamp = time.Unix(apiResp.Info.Timestamp, 0)
	}

	return domain.ExchangeRateResponse{
		Success:   true,
		Rate:      rateMoney,
		Amount:    resultMoney,
		Timestamp: timestamp,
	}, nil
}

func (c *Client) GetRate(ctx context.Context, from, to string, date time.Time) (domain.Money, error) {
	rateReq := domain.ExchangeRate{
		From: from,
		To:   to,
		Rate: domain.NewMoney(1.0, domain.DefaultScale),
		Date: date,
	}

	resp, err := c.Convert(ctx, rateReq)
	if err != nil {
		return domain.Money{}, err
	}
	return resp.Rate, nil
}

func (c *Client) BatchConvert(ctx context.Context, requests []domain.ExchangeRate) ([]domain.ExchangeRateResponse, error) {
	results := make([]domain.ExchangeRateResponse, len(requests))

	for i, req := range requests {
		resp, err := c.Convert(ctx, req)
		if err != nil {
			results[i] = domain.ExchangeRateResponse{Success: false}
			continue
		}
		results[i] = resp
	}
	return results, nil
}

func (c *Client) GetHistoricalRates(ctx context.Context, from, to string, startDate, endDate time.Time) ([]domain.ExchangeRateResponse, error) {
	var results []domain.ExchangeRateResponse
	current := startDate
	for !current.After(endDate) {
		rate, err := c.GetRate(ctx, from, to, current)
		if err != nil {
			current = current.AddDate(0, 0, 1)
			continue
		}

		results = append(results, domain.ExchangeRateResponse{
			Success:   true,
			Rate:      rate,
			Amount:    rate,
			Timestamp: current,
		})
		current = current.AddDate(0, 0, 1)
	}
	return results, nil
}

func (c *Client) ValidateConnection(ctx context.Context) error {
	_, err := c.GetRate(ctx, "USD", "EUR", time.Now())
	return err
}
