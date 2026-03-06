package marketdata

import (
	"fmt"

	"github.com/google/uuid"
)

type ProviderIngestionMode string

const (
	ProviderIngestionModeAPI    ProviderIngestionMode = "API"
	ProviderIngestionModeManual ProviderIngestionMode = "MANUAL"
)

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
	ErrProviderNameEmpty           = fmt.Errorf("provider name cannot be empty")
	ErrProviderBaseURIEmpty        = fmt.Errorf("provider base URI cannot be empty")
	ErrProviderAPIKeyEmpty         = fmt.Errorf("provider API key cannot be empty")
	ErrProviderIngestionModeEmpty  = fmt.Errorf("provider ingestion mode cannot be empty")
	ErrProviderNotFound            = fmt.Errorf("provider not found")
	ErrProviderSourceNotMapped     = fmt.Errorf("provider source not mapped")
	ErrProviderManualNotSupported  = fmt.Errorf("provider does not support manual ingestion")
	ErrProviderAutomaticNotAllowed = fmt.Errorf("provider does not support automatic ingestion")
)

type ProviderName string

const (
	ProviderMarketStack  ProviderName = "marketstack"
	ProviderAlphaVantage ProviderName = "alphavantage"
	ProviderBrandNewDay  ProviderName = "brandnewday"
)

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

func (p *Provider) IsManualIngestion() bool {
	return p != nil && p.IngestionMode == ProviderIngestionModeManual
}

func (p *Provider) IsAPIIngestion() bool {
	return p != nil && p.IngestionMode == ProviderIngestionModeAPI
}

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
