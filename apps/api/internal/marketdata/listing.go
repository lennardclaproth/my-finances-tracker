package marketdata

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

// Listing is the canonical market instrument record.
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
	// AccumulatedStart is the first EOD entry date that has been accumulated for this listing.
	AccumulatedStart *time.Time `db:"accumulated_start"`
	// AccumulatedEnd is the last date for which EOD accumulation finished.
	AccumulatedEnd   *time.Time `db:"accumulated_end"`
	ShouldAccumulate bool       `db:"should_accumulate"`
	Syncing          bool       `db:"syncing"`
	Source           Source     `db:"source"`
	ProviderID       *uuid.UUID `db:"provider"`
	Provider         Provider
	Currency         *money.Currency `db:"currency"`
}

// MissingField labels listing metadata fields that are currently absent.
type MissingField string

// Source identifies the upstream source system for listing data.
type Source string

// ListingOption mutates optional listing fields during construction.
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
	// ErrListingSymbolEmpty indicates missing listing symbol.
	ErrListingSymbolEmpty = fmt.Errorf("listing symbol cannot be empty")
	// ErrListingNameEmpty indicates missing listing name.
	ErrListingNameEmpty = fmt.Errorf("listing name cannot be empty")
	// ErrListingSourceEmpty indicates missing listing source.
	ErrListingSourceEmpty = fmt.Errorf("listing source cannot be empty")
	// ErrListingAlreadyExists indicates duplicate symbol/source insert.
	ErrListingAlreadyExists = fmt.Errorf("listing already exists")
	// ErrListingNotFound indicates listing lookup misses.
	ErrListingNotFound = fmt.Errorf("listing not found")
	// ErrNoListingFieldsToUpdate indicates an empty patch request.
	ErrNoListingFieldsToUpdate = fmt.Errorf("no listing fields to update")
	// ErrInvalidListingCurrency indicates unsupported listing currency value.
	ErrInvalidListingCurrency = fmt.Errorf("invalid listing currency")
	// ErrSyncInProgress indicates a listing sync lock is already held.

)

const (
	SourceAlphaVantage Source = "alpha_vantage"
	SourceMarketStack  Source = "market_stack"
	SourceBrandNewDay  Source = "brandnewday"
)

// IsManualIngestion returns true if the source is a manual ingestion provider,
// which means that data for listings with this source will not be automatically synced
// and must be ingested manually by the user.
func (s Source) IsManualIngestion() bool {
	return s == SourceBrandNewDay
}

// NewListing constructs a validated listing with optional metadata.
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

// ListingWithISIN sets listing ISIN.
func ListingWithISIN(isin string) ListingOption {
	return func(l *Listing) {
		l.ISIN = &isin
	}
}

// ListingWithExchange sets listing exchange.
func ListingWithExchange(exchange string) ListingOption {
	return func(l *Listing) {
		l.Exchange = &exchange
	}
}

// ListingWithCurrency sets listing currency.
func ListingWithCurrency(currency money.Currency) ListingOption {
	return func(l *Listing) {
		l.Currency = &currency
	}
}

// ListingWithDescription sets listing description.
func ListingWithDescription(desc string) ListingOption {
	return func(l *Listing) {
		l.Description = &desc
	}
}

// ListingWithRegion sets listing region.
func ListingWithRegion(region string) ListingOption {
	return func(l *Listing) {
		l.Region = &region
	}
}

// ListingWithType sets listing type.
func ListingWithType(typ string) ListingOption {
	return func(l *Listing) {
		l.Type = &typ
	}
}

// ListingWithTicker sets listing ticker.
func ListingWithTicker(ticker string) ListingOption {
	return func(l *Listing) {
		l.Ticker = &ticker
	}
}

// MissingFields reports optional metadata fields that are currently missing.
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
