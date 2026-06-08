package marketdata

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

// EOD is one OHLCV end-of-day datapoint for a listing/date.
type EOD struct {
	ID        uuid.UUID   `db:"id"`
	ListingID uuid.UUID   `db:"listing_id"`
	Symbol    string      `db:"symbol"` // Kept for response readability
	Date      time.Time   `db:"date"`
	Open      money.Price `db:"open"`
	Close     money.Price `db:"close"`
	High      money.Price `db:"high"`
	Low       money.Price `db:"low"`
	Volume    int64       `db:"volume"`
	CreatedAt time.Time   `db:"created_at"`
	UpdatedAt time.Time   `db:"updated_at"`
}

var (
	// ErrEODSymbolEmpty indicates missing symbol.
	ErrEODSymbolEmpty = fmt.Errorf("eod symbol cannot be empty")
	// ErrEODListingIDEmpty indicates missing listing identifier.
	ErrEODListingIDEmpty = fmt.Errorf("eod listing id cannot be empty")
)

// NewEOD constructs an EOD datapoint from decimal prices.
func NewEOD(symbol string, date time.Time, open, close, high, low float64, volume int64) (EOD, error) {
	if symbol == "" {
		return EOD{}, ErrEODSymbolEmpty
	}
	openCents, err := money.NewPrice(open)
	if err != nil {
		return EOD{}, fmt.Errorf("NewEOD failed, invalid open price: %w", err)
	}
	closeCents, err := money.NewPrice(close)
	if err != nil {
		return EOD{}, fmt.Errorf("NewEOD failed, invalid close price: %w", err)
	}
	highCents, err := money.NewPrice(high)
	if err != nil {
		return EOD{}, fmt.Errorf("NewEOD failed, invalid high price: %w", err)
	}
	lowCents, err := money.NewPrice(low)
	if err != nil {
		return EOD{}, fmt.Errorf("NewEOD failed, invalid low price: %w", err)
	}
	return EOD{
		ID:        uuid.New(),
		Symbol:    symbol,
		Date:      date,
		Open:      openCents,
		Close:     closeCents,
		High:      highCents,
		Low:       lowCents,
		Volume:    volume,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}
