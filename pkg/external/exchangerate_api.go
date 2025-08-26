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
		dateStr = date.Format("2025-08-26")
	}

	q := url.Values{}
	q.Set("access_key", c.apiKey)
	q.Set("from", params.From)
	q.Set("to", params.To)
	q.Set("amount", fmt.Sprintf("%f", params.Amount))
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
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return domain.ExchangeRateResponse{Success: false}, err
	}
	if !apiResp.Success {
		return domain.ExchangeRateResponse{Success: false}, fmt.Errorf("conversion failed")
	}

	return domain.ExchangeRateResponse{
		Success: true,
		Amount:  apiResp.Result,
	}, nil
}
