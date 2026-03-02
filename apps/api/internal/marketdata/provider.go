package marketdata

import (
	"fmt"
)

type Provider struct {
	Name      ProviderName `db:"name"`
	BaseURI   string       `db:"base_uri"`
	ApiKey    string       `db:"api_key"`
	Remaining int          `db:"remaining"`
	Used      int          `db:"used"`
	Total     int          `db:"total"`
	ResetsAt  string       `db:"resets_at"`
}

var (
	ErrProviderNameEmpty    = fmt.Errorf("provider name cannot be empty")
	ErrProviderBaseURIEmpty = fmt.Errorf("provider base URI cannot be empty")
	ErrProviderAPIKeyEmpty  = fmt.Errorf("provider API key cannot be empty")
	ErrProviderNotFound     = fmt.Errorf("provider not found")
)

type ProviderName string

const (
	ProviderMarketStack  ProviderName = "marketstack"
	ProviderAlphaVantage ProviderName = "alphavantage"
)

func NewProvider(name ProviderName, baseURI string) (*Provider, error) {
	if name == "" {
		return nil, ErrProviderNameEmpty
	}
	if baseURI == "" {
		return nil, ErrProviderBaseURIEmpty
	}
	return &Provider{
		Name:    name,
		BaseURI: baseURI,
	}, nil
}

func NewProviderWithAPIKey(name ProviderName, baseURI, apiKey string) (*Provider, error) {
	p, err := NewProvider(name, baseURI)
	if err != nil {
		return nil, err
	}
	if apiKey == "" {
		return nil, ErrProviderAPIKeyEmpty
	}
	p.ApiKey = apiKey
	return p, nil
}
