package marketdata

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

type Daily struct {
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
	ErrDailySymbolEmpty    = fmt.Errorf("daily symbol cannot be empty")
	ErrDailyListingIDEmpty = fmt.Errorf("daily listing id cannot be empty")
)

func NewDaily(symbol string, date time.Time, open, close, high, low float64, volume int64) (Daily, error) {
	if symbol == "" {
		return Daily{}, ErrDailySymbolEmpty
	}
	openCents, err := money.NewPrice(open)
	if err != nil {
		return Daily{}, fmt.Errorf("NewDaily failed, invalid open price: %w", err)
	}
	closeCents, err := money.NewPrice(close)
	if err != nil {
		return Daily{}, fmt.Errorf("NewDaily failed, invalid close price: %w", err)
	}
	highCents, err := money.NewPrice(high)
	if err != nil {
		return Daily{}, fmt.Errorf("NewDaily failed, invalid high price: %w", err)
	}
	lowCents, err := money.NewPrice(low)
	if err != nil {
		return Daily{}, fmt.Errorf("NewDaily failed, invalid low price: %w", err)
	}
	return Daily{
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
