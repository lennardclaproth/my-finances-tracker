package marketdata

import "testing"

func TestNewAPIProviderWithAPIKey_SetsAPIIngestionMode(t *testing.T) {
	p, err := NewAPIProviderWithAPIKey(ProviderMarketStack, "https://api.marketstack.com/v2", "key-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.IngestionMode != ProviderIngestionModeAPI {
		t.Fatalf("expected API ingestion mode, got %s", p.IngestionMode)
	}
	if p.BaseURI == nil || *p.BaseURI == "" {
		t.Fatalf("expected base_uri to be set")
	}
	if p.ApiKey == nil || *p.ApiKey == "" {
		t.Fatalf("expected api_key to be set")
	}
	if p.ResetsAt == nil {
		t.Fatalf("expected resets_at to be initialized")
	}
}

func TestNewAPIProviderWithAPIKey_RequiresAPIKey(t *testing.T) {
	_, err := NewAPIProviderWithAPIKey(ProviderMarketStack, "https://api.marketstack.com/v2", "")
	if err == nil {
		t.Fatalf("expected api key validation error")
	}
}

func TestNewManualProvider_SetsNullableConnectionFields(t *testing.T) {
	p, err := NewManualProvider(ProviderBrandNewDay)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.IngestionMode != ProviderIngestionModeManual {
		t.Fatalf("expected MANUAL ingestion mode, got %s", p.IngestionMode)
	}
	if p.BaseURI != nil || p.ApiKey != nil || p.ResetsAt != nil {
		t.Fatalf("expected manual provider connection fields to be nil")
	}
}

func TestProviderNameFromSource(t *testing.T) {
	tests := []struct {
		name   string
		source Source
		want   ProviderName
	}{
		{name: "marketstack", source: SourceMarketStack, want: ProviderMarketStack},
		{name: "alphavantage", source: SourceAlphaVantage, want: ProviderAlphaVantage},
		{name: "brandnewday", source: SourceBrandNewDay, want: ProviderBrandNewDay},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := ProviderNameFromSource(tc.source)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if got != tc.want {
				t.Fatalf("expected %s, got %s", tc.want, got)
			}
		})
	}
}
