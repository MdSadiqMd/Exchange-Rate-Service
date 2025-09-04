package domain

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
)

type Money struct {
	Amount int64 `json:"amount"`
	Scale  int   `json:"scale"`
}

const (
	DefaultScale = 6
	CentScale    = 2
)

func NewMoney(value float64, scale int) Money {
	if scale < 0 {
		scale = DefaultScale
	}
	multiplier := int64(math.Pow10(scale))
	return Money{
		Amount: int64(math.Round(value * float64(multiplier))),
		Scale:  scale,
	}
}

func NewMoneyFromString(value string, scale int) (Money, error) {
	val, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return Money{}, fmt.Errorf("invalid money format: %s", value)
	}
	return NewMoney(val, scale), nil
}

func (m Money) ToFloat() float64 {
	if m.Scale == 0 {
		return float64(m.Amount)
	}
	divisor := math.Pow10(m.Scale)
	return float64(m.Amount) / divisor
}

func (m Money) String() string {
	if m.Scale == 0 {
		return strconv.FormatInt(m.Amount, 10)
	}

	value := m.ToFloat()
	return fmt.Sprintf("%."+strconv.Itoa(m.Scale)+"f", value)
}

// Add performs precise addition
func (m Money) Add(other Money) Money {
	normalized := m.normalizeScale(other)
	return Money{
		Amount: normalized.m1.Amount + normalized.m2.Amount,
		Scale:  normalized.scale,
	}
}

// Subtract performs precise subtraction
func (m Money) Subtract(other Money) Money {
	normalized := m.normalizeScale(other)
	return Money{
		Amount: normalized.m1.Amount - normalized.m2.Amount,
		Scale:  normalized.scale,
	}
}

// Multiply performs precise multiplication with another Money (for rates)
func (m Money) Multiply(rate Money) Money {
	result := m.Amount * rate.Amount
	newScale := m.Scale + rate.Scale

	// Normalize to target precision (DefaultScale)
	if newScale > DefaultScale {
		divisor := int64(math.Pow10(newScale - DefaultScale))
		result = result / divisor
		newScale = DefaultScale
	}

	return Money{Amount: result, Scale: newScale}
}

// MultiplyByFloat multiplies by a float (use sparingly)
func (m Money) MultiplyByFloat(multiplier float64) Money {
	return Money{
		Amount: int64(float64(m.Amount) * multiplier),
		Scale:  m.Scale,
	}
}

// Divide performs precise division
func (m Money) Divide(divisor Money) Money {
	if divisor.Amount == 0 {
		return Money{Amount: 0, Scale: m.Scale}
	}

	normalized := m.normalizeScale(divisor)
	// Add extra precision to avoid truncation
	numerator := normalized.m1.Amount * int64(math.Pow10(DefaultScale))
	result := numerator / normalized.m2.Amount

	return Money{Amount: result, Scale: DefaultScale}
}

func (m Money) IsZero() bool {
	return m.Amount == 0
}

func (m Money) IsPositive() bool {
	return m.Amount > 0
}

func (m Money) IsNegative() bool {
	return m.Amount < 0
}

func (m Money) ConvertToScale(targetScale int) Money {
	if m.Scale == targetScale {
		return m
	}

	if targetScale > m.Scale {
		multiplier := int64(math.Pow10(targetScale - m.Scale))
		return Money{Amount: m.Amount * multiplier, Scale: targetScale}
	}

	divisor := int64(math.Pow10(m.Scale - targetScale))
	return Money{Amount: m.Amount / divisor, Scale: targetScale}
}

type normalizedPair struct {
	m1, m2 Money
	scale  int
}

func (m Money) normalizeScale(other Money) normalizedPair {
	targetScale := m.Scale
	if other.Scale > targetScale {
		targetScale = other.Scale
	}

	return normalizedPair{
		m1:    m.ConvertToScale(targetScale),
		m2:    other.ConvertToScale(targetScale),
		scale: targetScale,
	}
}

func (m Money) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"amount": m.Amount,
		"scale":  m.Scale,
		"value":  m.String(),
	})
}

func (m *Money) UnmarshalJSON(data []byte) error {
	var temp struct {
		Amount *int64 `json:"amount"`
		Scale  *int   `json:"scale"`
		Value  string `json:"value"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if temp.Amount != nil && temp.Scale != nil {
		m.Amount = *temp.Amount
		m.Scale = *temp.Scale
		return nil
	}

	if temp.Value != "" {
		parsed, err := NewMoneyFromString(temp.Value, DefaultScale)
		if err != nil {
			return err
		}
		*m = parsed
		return nil
	}

	return fmt.Errorf("invalid money JSON format")
}

func (m Money) Validate() error {
	if m.Scale < 0 {
		return fmt.Errorf("scale cannot be negative")
	}
	if m.Scale > 18 {
		return fmt.Errorf("scale too large (max 18)")
	}
	return nil
}
