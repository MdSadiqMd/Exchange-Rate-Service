package domain

import (
	"fmt"
	"time"
)

type ConversionRequest struct {
	From   string    `json:"from"`
	To     string    `json:"to"`
	Amount float64   `json:"amount"`
	Date   time.Time `json:"date,omitempty"`
}

type ConversionResponse struct {
	Success bool            `json:"success"`
	Query   ConversionQuery `json:"query"`
	Info    ConversionInfo  `json:"info"`
	Result  float64         `json:"result"`
}

type ConversionQuery struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
}

type ConversionInfo struct {
	Rate      float64 `json:"rate"`
	Timestamp int64   `json:"timestamp"`
}

func (r *ConversionRequest) Validate() error {
	if r.From == "" {
		return fmt.Errorf("from currency is required")
	}
	if r.To == "" {
		return fmt.Errorf("to currency is required")
	}
	if r.From == r.To {
		return fmt.Errorf("from and to currencies cannot be the same")
	}
	if !IsValidCurrency(r.From) {
		return fmt.Errorf("unsupported currency: %s", r.From)
	}
	if !IsValidCurrency(r.To) {
		return fmt.Errorf("unsupported currency: %s", r.To)
	}
	if r.Amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	if !r.Date.IsZero() {
		ninetyDaysAgo := time.Now().AddDate(0, 0, -90)
		if r.Date.Before(ninetyDaysAgo) {
			return fmt.Errorf("date is too old (max 90 days)")
		}
		if r.Date.After(time.Now()) {
			return fmt.Errorf("date cannot be in the future")
		}
	}
	return nil
}
