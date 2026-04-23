package marketdata

import (
	"fmt"

	"github.com/google/uuid"
)

// ProviderIngestionMode defines how provider data is ingested.
type ProviderIngestionMode string

const (
	ProviderIngestionModeAPI    ProviderIngestionMode = "API"
	ProviderIngestionModeManual ProviderIngestionMode = "MANUAL"
)

// Provider stores provider connection metadata and token state.
type Provider struct {
	ID            uuid.UUID             `db:"id"`
	Name          ProviderName          `db:"name"`
	IngestionMode ProviderIngestionMode `db:"ingestion_mode"`
	BaseURI       *string               `db:"base_uri"`
	ApiKey        *string               `db:"api_key"`
	Remaining     int                   `db:"remaining"`
	Used          int                   `db:"used"`
	Total         int                   `db:"total"`
	ResetsAt      *string               `db:"resets_at"`
}

var (
	// ErrProviderNameEmpty indicates missing provider name.
	ErrProviderNameEmpty = fmt.Errorf("provider name cannot be empty")
	// ErrProviderBaseURIEmpty indicates missing provider base URI for API providers.
	ErrProviderBaseURIEmpty = fmt.Errorf("provider base URI cannot be empty")
	// ErrProviderAPIKeyEmpty indicates missing provider API key for API providers.
	ErrProviderAPIKeyEmpty = fmt.Errorf("provider API key cannot be empty")
	// ErrProviderIngestionModeEmpty indicates missing ingestion mode.
	ErrProviderIngestionModeEmpty = fmt.Errorf("provider ingestion mode cannot be empty")
	// ErrProviderNotFound indicates that the provider does not exist.
	ErrProviderNotFound = fmt.Errorf("provider not found")
	// ErrProviderSourceNotMapped indicates unknown listing source to provider mapping.
	ErrProviderSourceNotMapped = fmt.Errorf("provider source not mapped")
	// ErrProviderManualNotSupported indicates invalid manual-provider field population.
	ErrProviderManualNotSupported = fmt.Errorf("provider does not support manual ingestion")
	// ErrProviderAutomaticNotAllowed indicates invalid API-provider field population.
	ErrProviderAutomaticNotAllowed = fmt.Errorf("provider does not support automatic ingestion")
)

// ProviderName is the canonical provider identifier.
type ProviderName string

const (
	ProviderMarketStack  ProviderName = "marketstack"
	ProviderAlphaVantage ProviderName = "alphavantage"
	ProviderBrandNewDay  ProviderName = "brandnewday"
)

// NewAPIProviderWithAPIKey builds an API-ingestion provider.
func NewAPIProviderWithAPIKey(name ProviderName, baseURI, apiKey string) (*Provider, error) {
	if name == "" {
		return nil, ErrProviderNameEmpty
	}
	if baseURI == "" {
		return nil, ErrProviderBaseURIEmpty
	}
	if apiKey == "" {
		return nil, ErrProviderAPIKeyEmpty
	}
	emptyResetsAt := ""

	return &Provider{
		ID:            uuid.New(),
		Name:          name,
		IngestionMode: ProviderIngestionModeAPI,
		BaseURI:       &baseURI,
		ApiKey:        &apiKey,
		ResetsAt:      &emptyResetsAt,
	}, nil
}

// NewManualProvider builds a manual-ingestion provider.
func NewManualProvider(name ProviderName) (*Provider, error) {
	if name == "" {
		return nil, ErrProviderNameEmpty
	}

	return &Provider{
		ID:            uuid.New(),
		Name:          name,
		IngestionMode: ProviderIngestionModeManual,
		BaseURI:       nil,
		ApiKey:        nil,
		ResetsAt:      nil,
	}, nil
}

// ProviderNameFromSource maps a listing source to its provider name.
func ProviderNameFromSource(source Source) (ProviderName, error) {
	switch source {
	case SourceMarketStack:
		return ProviderMarketStack, nil
	case SourceAlphaVantage:
		return ProviderAlphaVantage, nil
	case SourceBrandNewDay:
		return ProviderBrandNewDay, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrProviderSourceNotMapped, source)
	}
}

// IsManualIngestion reports whether the provider uses manual ingestion.
func (p *Provider) IsManualIngestion() bool {
	return p != nil && p.IngestionMode == ProviderIngestionModeManual
}

// IsAPIIngestion reports whether the provider uses API ingestion.
func (p *Provider) IsAPIIngestion() bool {
	return p != nil && p.IngestionMode == ProviderIngestionModeAPI
}

// Validate checks provider fields for ingestion-mode compatibility.
func (p *Provider) Validate() error {
	if p == nil {
		return fmt.Errorf("provider cannot be nil")
	}
	if p.Name == "" {
		return ErrProviderNameEmpty
	}
	if p.IngestionMode == "" {
		return ErrProviderIngestionModeEmpty
	}
	switch p.IngestionMode {
	case ProviderIngestionModeManual:
		if p.BaseURI != nil || p.ApiKey != nil || p.ResetsAt != nil {
			return ErrProviderManualNotSupported
		}
		return nil
	case ProviderIngestionModeAPI:
		if p.BaseURI == nil || *p.BaseURI == "" {
			return ErrProviderBaseURIEmpty
		}
		if p.ApiKey == nil || *p.ApiKey == "" {
			return ErrProviderAPIKeyEmpty
		}
		if p.ResetsAt == nil {
			return ErrProviderAutomaticNotAllowed
		}
		return nil
	default:
		return ErrProviderIngestionModeEmpty
	}
}
