package eodimport

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	// ErrManualUploadListingNotFound indicates the listing does not exist.
	ErrManualUploadListingNotFound = fmt.Errorf("manual upload listing not found")
	// ErrManualUploadProviderUnavailable indicates no provider can serve the listing source.
	ErrManualUploadProviderUnavailable = fmt.Errorf("manual upload provider unavailable")
	// ErrManualUploadProviderNotManual indicates the provider does not support manual ingestion.
	ErrManualUploadProviderNotManual = fmt.Errorf("manual upload provider is not manual")
	// ErrManualUploadParserUnavailable indicates no parser exists for the listing source.
	ErrManualUploadParserUnavailable = fmt.Errorf("manual upload parser unavailable")
)

type manualUploadListingStore interface {
	FetchByID(ctx context.Context, id uuid.UUID) (*Listing, error)
}

type manualUploadProviderStore interface {
	GetByName(ctx context.Context, name ProviderName) (*Provider, error)
}

type manualUploadParserValidator func(source Source) error

// ManualUploadPolicy validates whether a listing can accept manual daily uploads.
type ManualUploadPolicy struct {
	listings    manualUploadListingStore
	providers   manualUploadProviderStore
	parserValid manualUploadParserValidator
}

// NewManualUploadPolicy constructs a manual upload validation policy.
func NewManualUploadPolicy(
	listings manualUploadListingStore,
	providers manualUploadProviderStore,
	parserValid manualUploadParserValidator,
) *ManualUploadPolicy {
	return &ManualUploadPolicy{
		listings:    listings,
		providers:   providers,
		parserValid: parserValid,
	}
}

// ValidateListing verifies provider/parser support and returns the target listing.
func (p *ManualUploadPolicy) ValidateListing(ctx context.Context, listingID uuid.UUID) (*Listing, error) {
	if p == nil || p.listings == nil || p.providers == nil {
		return nil, ErrManualUploadProviderUnavailable
	}
	listing, err := p.listings.FetchByID(ctx, listingID)
	if err != nil {
		return nil, err
	}
	if listing == nil {
		return nil, ErrManualUploadListingNotFound
	}

	providerName, err := ProviderNameFromSource(listing.Source)
	if err != nil {
		return nil, ErrManualUploadProviderUnavailable
	}
	provider, err := p.providers.GetByName(ctx, providerName)
	if err != nil {
		if errors.Is(err, ErrProviderNotFound) {
			return nil, ErrManualUploadProviderUnavailable
		}
		return nil, err
	}
	if provider == nil {
		return nil, ErrManualUploadProviderUnavailable
	}
	if !provider.IsManualIngestion() {
		return nil, ErrManualUploadProviderNotManual
	}
	if p.parserValid == nil || p.parserValid(listing.Source) != nil {
		return nil, ErrManualUploadParserUnavailable
	}
	return listing, nil
}
