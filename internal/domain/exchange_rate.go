package domain

import "time"

type ExchangeRate struct {
	From string    `json:"from"`
	To   string    `json:"to"`
	Rate Money     `json:"rate"`
	Date time.Time `json:"date"`
}

type ExchangeRateResponse struct {
	Success   bool      `json:"success"`
	Rate      Money     `json:"rate"`
	Amount    Money     `json:"amount"`
	Timestamp time.Time `json:"timestamp"`
}

type RateCache struct {
	BaseRates   map[string]Money `json:"base_rates"`  // 1-hour cache for base rates
	Adjustments map[string]Money `json:"adjustments"` // 5-minute cache for rate adjustments
	LastFetch   time.Time        `json:"last_fetch"`
	LastUpdate  time.Time        `json:"last_update"`
}

func (c *RateCache) GetPrecisionRate(from, to string) Money {
	key := from + ":" + to
	base := c.BaseRates[key]
	adjustment := c.Adjustments[key]
	if base.IsZero() {
		return Money{}
	}

	if !adjustment.IsZero() {
		return base.Add(adjustment)
	}

	return base
}

func (c *RateCache) IsStale(maxAge time.Duration) bool {
	return time.Since(c.LastFetch) > maxAge
}

func (c *RateCache) CrossRate(from, to, base string) Money {
	fromRate := c.GetPrecisionRate(base, from)
	toRate := c.GetPrecisionRate(base, to)

	if fromRate.IsZero() || toRate.IsZero() {
		return Money{}
	}

	return toRate.Divide(fromRate)
}

func (e *ExchangeRate) ConvertToMoney() ExchangeRate {
	return ExchangeRate{
		From: e.From,
		To:   e.To,
		Rate: NewMoney(float64(e.Rate.Amount), DefaultScale),
		Date: e.Date,
	}
}
