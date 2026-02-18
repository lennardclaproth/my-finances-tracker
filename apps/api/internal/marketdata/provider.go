package marketdata

import (
	"fmt"
)

type Quota struct {
	Remaining int    `db:"remaining"`
	Used      int    `db:"used"`
	Total     int    `db:"total"`
	ResetsAt  string `db:"resets_at"`
}

type Provider struct {
	Name    ProviderName `db:"name"`
	BaseURI string       `db:"base_uri"`
	ApiKey  string       `db:"api_key"`
}

var (
	ErrProviderNameEmpty   = fmt.Errorf("provider name cannot be empty")
	ErrProviderAPIKeyEmpty = fmt.Errorf("provider API key cannot be empty")
)

type ProviderName string

const (
	ProviderMarketStack ProviderName = "marketstack"
)

func NewProvider(name ProviderName, baseURI string) (*Provider, error) {
	if name == "" {
		return nil, fmt.Errorf("provider name cannot be empty")
	}
	if baseURI == "" {
		return nil, fmt.Errorf("provider base URI cannot be empty")
	}
	return &Provider{
		Name:    name,
		BaseURI: baseURI,
	}, nil
}

func NewProviderWithAPIKey(name ProviderName, baseURI, apiKey string) (*Provider, error) {
	if name == "" {
		return nil, fmt.Errorf("provider name cannot be empty")
	}
	if baseURI == "" {
		return nil, fmt.Errorf("provider base URI cannot be empty")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("provider API key cannot be empty")
	}
	return &Provider{
		Name:    name,
		BaseURI: baseURI,
		ApiKey:  apiKey,
	}, nil
}
