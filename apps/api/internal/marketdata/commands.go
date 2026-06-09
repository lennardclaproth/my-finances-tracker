package marketdata

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

// CommandStore persists market-data listings, EOD datapoints, and providers.
type CommandStore interface {
	Get(ctx context.Context, lsID uuid.UUID) (*Listing, error)
	Create(ctx context.Context, listing *Listing) error
	Update(ctx context.Context, listing *Listing) error
	CreateEODs(ctx context.Context, eods []*EOD) (int, error)
	CreateProvider(ctx context.Context, provider *Provider) error
}

// Commands exposes market-data write-side use cases.
type Commands struct {
	cs CommandStore
	s  *Syncer
}

// NewCommands creates market-data write-side use cases.
func NewCommands(cs CommandStore, s *Syncer) *Commands {
	return &Commands{cs: cs, s: s}
}

// CreateProvider persists provider metadata through the market-data feature boundary.
func (c *Commands) CreateProvider(ctx context.Context, provider *Provider) error {
	if err := c.cs.CreateProvider(ctx, provider); err != nil {
		return fmt.Errorf("create provider: %w", err)
	}
	return nil
}

// CreateListing creates a new listing and persists it.
func (c *Commands) CreateListing(
	ctx context.Context,
	symbol, name string,
	source Source,
	options ...ListingOption,
) (*Listing, error) {
	listing, err := NewListing(symbol, name, source, options...)
	if err != nil {
		return nil, fmt.Errorf("create listing: failed to create listing: %w", err)
	}
	err = c.cs.Create(ctx, listing)
	if errors.Is(err, ErrListingAlreadyExists) {
		return nil, fmt.Errorf("create listing: %w", err)
	}

	if err != nil {
		return nil, fmt.Errorf("create listing: failed to persist listing: %w", err)
	}

	if listing.Source.IsManualIngestion() {
		return listing, nil
	}
	c.s.SyncEOD(ctx, listing.ID, nil, nil)
	return listing, nil
}

// UpdateListingFields updates only the provided listing fields.
func (c *Commands) UpdateListingFields(
	ctx context.Context,
	id uuid.UUID,
	description, exchange, region, currency, isin, ticker, typ *string,
) (*Listing, error) {
	if description == nil && exchange == nil && region == nil && currency == nil && isin == nil && ticker == nil && typ == nil {
		return nil, ErrNoListingFieldsToUpdate
	}
	listing, err := c.cs.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("UpdateListingFields failed to fetch listing: %w", err)
	}
	if listing == nil {
		return nil, ErrListingNotFound
	}

	if description != nil {
		listing.Description = description
	}
	if exchange != nil {
		listing.Exchange = exchange
	}
	if region != nil {
		listing.Region = region
	}
	if isin != nil {
		listing.ISIN = isin
	}
	if ticker != nil {
		listing.Ticker = ticker
	}
	if typ != nil {
		listing.Type = typ
	}
	if currency != nil {
		cur := money.Currency(*currency)
		if !cur.IsValid() {
			return nil, ErrInvalidListingCurrency
		}
		listing.Currency = &cur
	}

	if err := c.cs.Update(ctx, listing); err != nil {
		return nil, fmt.Errorf("update listing fields: failed to persist listing: %w", err)
	}
	return listing, nil
}

// EODInput is one parsed EOD datapoint awaiting persistence for a listing.
type EODInput struct {
	Date   time.Time
	Open   float64
	Close  float64
	High   float64
	Low    float64
	Volume int64
}

// CreateEODsResult reports the outcome of a batch EOD create.
type CreateEODsResult struct {
	// Imported is the number of datapoints newly inserted.
	Imported int
	// Duplicates is the number skipped because they already existed.
	Duplicates int
}

// CreateEODs builds EOD datapoints for a listing and persists them with a single bulk
// insert, skipping and counting datapoints that already exist.
func (c *Commands) CreateEODs(ctx context.Context, listingID uuid.UUID, symbol string, inputs []EODInput) (CreateEODsResult, error) {
	if len(inputs) == 0 {
		return CreateEODsResult{}, nil
	}

	eods := make([]*EOD, 0, len(inputs))
	for _, in := range inputs {
		eod, err := NewEOD(symbol, in.Date, in.Open, in.Close, in.High, in.Low, in.Volume)
		if err != nil {
			return CreateEODsResult{}, fmt.Errorf("create eods: %w", err)
		}
		eod.ListingID = listingID
		eods = append(eods, &eod)
	}

	inserted, err := c.cs.CreateEODs(ctx, eods)
	if err != nil {
		return CreateEODsResult{}, fmt.Errorf("create eods: %w", err)
	}

	return CreateEODsResult{
		Imported:   inserted,
		Duplicates: len(eods) - inserted,
	}, nil
}
