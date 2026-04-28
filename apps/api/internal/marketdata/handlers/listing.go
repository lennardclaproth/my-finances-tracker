package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

type listingCreator interface {
	Create(ctx context.Context, listing *marketdata.Listing) error
}

type listingSearcher interface {
	Search(ctx context.Context, q string, limit, offset *int) ([]*marketdata.Listing, int, error)
}

type listingLister interface {
	List(ctx context.Context, limit, offset *int) ([]*marketdata.Listing, error)
}

type listingUpdater interface {
	Update(ctx context.Context, listing *marketdata.Listing) error
}

type ListingHandler struct {
	lg  listingGetter
	lc  listingCreator
	lu  listingUpdater
	ll  listingLister
	ls  listingSearcher
	log logging.Logger
	sh  *SyncHandler
}

func NewListingHandler() *ListingHandler {
	return &ListingHandler{}
}

// CreateListing creates a new listing and persists it.
func (h *ListingHandler) CreateListing(
	ctx context.Context,
	symbol, name string,
	source marketdata.Source,
	options ...marketdata.ListingOption,
) (*marketdata.Listing, error) {
	// Check if listing with source and symbol already exists
	existing, err := h.lg.GetBySymbol(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("handlers: CreateListing failed to fetch listing: %w", err)
	}
	if existing != nil && existing.Source == source {
		return nil, fmt.Errorf("handlers: CreateListing failed: %w", marketdata.ErrListingAlreadyExists)
	}
	// Create new listing and persist
	listing, err := marketdata.NewListing(symbol, name, source, options...)
	if err != nil {
		return nil, fmt.Errorf("handlers: CreateListing failed to create listing: %w", err)
	}
	if err := h.lc.Create(ctx, listing); err != nil {
		if errors.Is(err, marketdata.ErrListingAlreadyExists) {
			return nil, fmt.Errorf("handlers: CreateListing failed: %w", marketdata.ErrListingAlreadyExists)
		}
		return nil, fmt.Errorf("handlers: CreateListing failed to persist listing: %w", err)
	}

	if listing.Source.IsManualIngestion() {
		return listing, nil
	}

	h.sh.SyncEOD(ctx, listing.ID, nil, nil)
	return listing, nil
}

// UpdateListingFields updates only the provided listing fields.
func (h *ListingHandler) UpdateListingFields(
	ctx context.Context,
	id uuid.UUID,
	description, exchange, region, currency, isin, ticker, typ *string,
) (*marketdata.Listing, error) {
	if description == nil && exchange == nil && region == nil && currency == nil && isin == nil && ticker == nil && typ == nil {
		return nil, ErrNoListingFieldsToUpdate
	}
	listing, err := h.lg.GetIncludingProvider(ctx, id)
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

	if err := h.lu.Update(ctx, listing); err != nil {
		return nil, fmt.Errorf("UpdateListingFields failed to persist listing: %w", err)
	}
	return listing, nil
}

// ListListings returns all listings in deterministic order for UI presentation.
func (h *ListingHandler) ListListings(ctx context.Context) ([]*marketdata.Listing, error) {
	listings, err := h.ll.List(ctx, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("ListListings failed to fetch listings: %w", err)
	}
	return listings, nil
}

func (h *ListingHandler) SearchListings(ctx context.Context, q string, limit, offset int) ([]*marketdata.Listing, int, error) {
	listings, total, err := h.ls.Search(ctx, q, &limit, &offset)
	if err != nil {
		return nil, 0, fmt.Errorf("SearchListings failed to fetch listings: %w", err)
	}
	return listings, total, nil
}
