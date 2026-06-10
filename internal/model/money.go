package model

import (
	"fmt"
	"strconv"
	"strings"
)

// Money represents an amount in euro-cents to avoid float64 drift.
type Money int64

func (m Money) Euros() float64 { return float64(m) / 100.0 }

func (m Money) String() string { return fmt.Sprintf("%.2f", m.Euros()) }

func MoneyFromString(s string) (Money, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", "")
	if s == "" {
		return 0, nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid money value %q: %w", s, err)
	}
	return moneyFromFloat(f), nil
}

func moneyFromFloat(f float64) Money {
	if f < 0 {
		return Money(int64(f*100 - 0.5))
	}
	return Money(int64(f*100 + 0.5))
}

func Abs(m Money) Money {
	if m < 0 {
		return -m
	}
	return m
}
