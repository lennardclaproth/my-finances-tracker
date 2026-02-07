package marketdata

import (
	"time"

	"github.com/google/uuid"
)

type History struct {
	ID     uuid.UUID `db:"id"`
	Symbol string    `db:"symbol"`
	Date   time.Time `db:"date"`
	Open   float64   `db:"open"`
	Close  float64   `db:"close"`
	High   float64   `db:"high"`
	Low    float64   `db:"low"`
	Volume int64     `db:"volume"`
}

func NewHistory(symbol string, date time.Time, open, close, high, low float64, volume int64) *History {
	return &History{
		ID:     uuid.New(),
		Symbol: symbol,
		Date:   date,
		Open:   open,
		Close:  close,
		High:   high,
		Low:    low,
		Volume: volume,
	}
}
