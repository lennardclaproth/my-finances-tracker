package parsers

import (
	"testing"

	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
)

func TestCreateEODParser_BrandNewDay(t *testing.T) {
	t.Parallel()

	parser, err := CreateEODParser(marketdata.SourceBrandNewDay)
	if err != nil {
		t.Fatalf("expected parser, got error: %v", err)
	}
	if parser == nil {
		t.Fatalf("expected non-nil parser")
	}
}

func TestCreateEODParser_UnknownSource(t *testing.T) {
	t.Parallel()

	_, err := CreateEODParser(marketdata.SourceAlphaVantage)
	if err == nil {
		t.Fatalf("expected unsupported source error")
	}
}
