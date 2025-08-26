package domain

import "time"

type ExchangeRate struct {
	From      string    `json:"from"`
	To        string    `json:"to"`
	Rate      float64   `json:"rate"`
	Date      time.Time `json:"date"`
	Timestamp int64     `json:"timestamp"`
}

type ExchangeRateResponse struct {
	Success   bool               `json:"success"`
	Base      string             `json:"base"`
	Date      string             `json:"date"`
	Rates     map[string]float64 `json:"rates"`
	Timestamp int64              `json:"timestamp"`
}
