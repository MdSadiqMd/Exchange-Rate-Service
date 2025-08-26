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
	Success bool    `json:"success"`
	Result  float64 `json:"result"`
}

func (r *ConversionRequest) Validate() error {
	if r.From == "" {
		return fmt.Errorf("from currency is required")
	}
	if r.To == "" {
		return fmt.Errorf("to currency is required")
	}
	if r.Amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	if r.Date.After(time.Now()) {
		return fmt.Errorf("date cannot be in the future")
	}
	if !r.Date.IsZero() && r.Date.Before(time.Now().AddDate(0, 0, -90)) {
		return fmt.Errorf("date is too old (max 90 days)")
	}
	return nil
}
