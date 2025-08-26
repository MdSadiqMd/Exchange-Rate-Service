package domain

import "time"

type ExchangeRate struct {
	From   string    `json:"from"`
	To     string    `json:"to"`
	Amount float64   `json:"amount"`
	Date   time.Time `json:"date,omitempty"`
}

type ExchangeRateResponse struct {
	Success bool    `json:"success"`
	Amount  float64 `json:"amount"`
}
