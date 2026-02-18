package marketdata

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

type Listing struct {
	ID          uuid.UUID `db:"id"`
	Symbol      string    `db:"symbol"`
	Name        string    `db:"name"`
	Type        *string   `db:"type"`
	Ticker      *string   `db:"ticker"`
	ISIN        *string   `db:"isin"`
	Description *string   `db:"description"`
	Exchange    *string   `db:"exchange"`
	Region      *string   `db:"region"`
	Active      bool      `db:"active"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	// AccumulatedStart is the first history entry date that has been accumulated for this listing. This signifies the available data we have.
	AccumulatedStart *time.Time `db:"accumulated_start"`
	// AccumulatedEnd is the end of the accumulation period. This is the last date for which data has been accumulated.
	AccumulatedEnd   *time.Time      `db:"accumulated_end"`
	ShouldAccumulate bool            `db:"should_accumulate"`
	Syncing          bool            `db:"syncing"`
	Source           Source          `db:"source"`
	Currency         *money.Currency `db:"currency"`
}

type MissingField string
type Source string
type ListingOption func(*Listing)

const (
	ListingISINMissing        MissingField = "isin_missing"
	ListingDescriptionMissing MissingField = "description_missing"
	ExchangeMissing           MissingField = "exchange_missing"
	RegionMissing             MissingField = "region_missing"
	TypeMissing               MissingField = "type_missing"
	CurrencyMissing           MissingField = "currency_missing"
	TickerMissing             MissingField = "ticker_missing"
)

var (
	ErrListingSymbolEmpty = fmt.Errorf("listing symbol cannot be empty")
	ErrListingNameEmpty   = fmt.Errorf("listing name cannot be empty")
	ErrListingSourceEmpty = fmt.Errorf("listing source cannot be empty")
	ErrSyncInProgress     = fmt.Errorf("sync is already in progress for this listing")
)

const (
	SourceAlphaVantage Source = "alpha_vantage"
)

func NewListing(symbol, name string, source Source, options ...ListingOption) (*Listing, error) {
	if symbol == "" {
		return nil, ErrListingSymbolEmpty
	}
	if name == "" {
		return nil, ErrListingNameEmpty
	}
	if source == "" {
		return nil, ErrListingSourceEmpty
	}

	listing := &Listing{
		ID:               uuid.New(),
		Symbol:           symbol,
		Name:             name,
		Active:           true,
		Source:           source,
		AccumulatedStart: nil,
		AccumulatedEnd:   nil,
		ShouldAccumulate: true,
	}

	for _, option := range options {
		option(listing)
	}
	return listing, nil
}

func ListingWithISIN(isin string) ListingOption {
	return func(l *Listing) {
		l.ISIN = &isin
	}
}

func ListingWithExchange(exchange string) ListingOption {
	return func(l *Listing) {
		l.Exchange = &exchange
	}
}

func ListingWithCurrency(currency money.Currency) ListingOption {
	return func(l *Listing) {
		l.Currency = &currency
	}
}

func ListingWithDescription(desc string) ListingOption {
	return func(l *Listing) {
		l.Description = &desc
	}
}

func ListingWithRegion(region string) ListingOption {
	return func(l *Listing) {
		l.Region = &region
	}
}

func ListingWithType(typ string) ListingOption {
	return func(l *Listing) {
		l.Type = &typ
	}
}

func ListingWithTicker(ticker string) ListingOption {
	return func(l *Listing) {
		l.Ticker = &ticker
	}
}

func (l *Listing) MissingFields() []MissingField {
	var missing []MissingField
	if l.ISIN == nil || *l.ISIN == "" {
		missing = append(missing, ListingISINMissing)
	}
	if l.Description == nil || *l.Description == "" {
		missing = append(missing, ListingDescriptionMissing)
	}
	if l.Exchange == nil || *l.Exchange == "" {
		missing = append(missing, ExchangeMissing)
	}
	if l.Region == nil || *l.Region == "" {
		missing = append(missing, RegionMissing)
	}
	if l.Type == nil || *l.Type == "" {
		missing = append(missing, TypeMissing)
	}
	if l.Currency == nil || *l.Currency == "" {
		missing = append(missing, CurrencyMissing)
	}
	if l.Ticker == nil || *l.Ticker == "" {
		missing = append(missing, TickerMissing)
	}
	return missing
}
