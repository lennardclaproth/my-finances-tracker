package money

import (
	"fmt"
	"math"
	"strconv"
	"strings"
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

// ParsePrice parses a non-negative decimal price with up to six fractional
// digits into its fixed-scale representation.
//
// Accepted examples include "12", "12.34", and "0.000001".
// Values with more than six fractional digits, exponent notation, locale
// separators, negative values, and values outside the Price range are rejected.
func ParsePrice(raw string) (Price, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0, invalidPrice(raw)
	}

	if strings.HasPrefix(s, "-") {
		return 0, invalidPrice(raw)
	}

	if strings.HasPrefix(s, "+") {
		s = s[1:]
		if s == "" {
			return 0, invalidPrice(raw)
		}
	}

	parts := strings.Split(s, ".")
	if len(parts) > 2 {
		return 0, invalidPrice(raw)
	}

	wholePart := parts[0]
	if wholePart == "" {
		return 0, invalidPrice(raw)
	}

	whole, err := strconv.ParseInt(wholePart, 10, 64)
	if err != nil || whole < 0 {
		return 0, invalidPrice(raw)
	}

	fractionPart := ""
	if len(parts) == 2 {
		fractionPart = parts[1]
		if fractionPart == "" || len(fractionPart) > 6 {
			return 0, invalidPrice(raw)
		}
	}

	fraction := int64(0)
	if fractionPart != "" {
		paddedFraction := fractionPart + strings.Repeat("0", 6-len(fractionPart))

		fraction, err = strconv.ParseInt(paddedFraction, 10, 64)
		if err != nil {
			return 0, invalidPrice(raw)
		}
	}

	if whole > math.MaxInt64/Scale {
		return 0, invalidPrice(raw)
	}

	scaledWhole := whole * Scale
	if scaledWhole > math.MaxInt64-fraction {
		return 0, invalidPrice(raw)
	}

	return Price(scaledWhole + fraction), nil
}

func invalidPrice(raw string) error {
	return fmt.Errorf("%w: %q", ErrInvalidPrice, raw)
}
