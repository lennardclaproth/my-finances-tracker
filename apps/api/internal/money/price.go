package money

import (
	"fmt"
	"math"
)

type Price int64

const Scale int64 = 1_000_000 // 6 decimal places

var (
	ErrInvalidPrice = fmt.Errorf("invalid price value")
)

func NewPrice(amount float64) (Price, error) {
	if math.IsNaN(amount) || math.IsInf(amount, 0) || amount < 0 {
		return 0, ErrInvalidPrice
	}

	return Price(math.Round(amount * float64(Scale))), nil
}

func (p Price) Float64() float64 {
	return float64(p) / float64(Scale)
}

func (p Price) String() string {
	return fmt.Sprintf("%.6f", p.Float64())
}

