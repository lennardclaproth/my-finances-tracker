package marketdata

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

type commandStore interface {
	Get(ctx context.Context, lsID uuid.UUID) (*Listing, error)
	Create(ctx context.Context, listing *Listing) error
	Update(ctx context.Context, listing *Listing) error
}

type Commands struct {
	cs commandStore
	s  *Syncer
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
