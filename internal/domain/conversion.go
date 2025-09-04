package domain

import (
	"fmt"
	"time"
)

type ConversionRequest struct {
	From   string    `json:"from"`
	To     string    `json:"to"`
	Amount Money     `json:"amount"`
	Date   time.Time `json:"date,omitempty"`
}

type ConversionResponse struct {
	Success bool  `json:"success"`
	Result  Money `json:"result"`
	Rate    Money `json:"rate"`
}

func (r *ConversionRequest) Validate() error {
	if r.From == "" {
		return fmt.Errorf("from currency is required")
	}
	if r.To == "" {
		return fmt.Errorf("to currency is required")
	}
	if r.Amount.IsZero() || r.Amount.IsNegative() {
		return fmt.Errorf("amount must be positive")
	}
	if err := r.Amount.Validate(); err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}
	if r.Date.After(time.Now()) {
		return fmt.Errorf("date cannot be in the future")
	}
	if !r.Date.IsZero() && r.Date.Before(time.Now().AddDate(0, 0, -90)) {
		return fmt.Errorf("date is too old (max 90 days)")
	}
	return nil
}
