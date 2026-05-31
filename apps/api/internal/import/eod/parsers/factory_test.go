package parsers

import (
	"testing"

	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
)

func TestCreateDailyParser_BrandNewDay(t *testing.T) {
	t.Parallel()

	parser, err := CreateDailyParser(marketdata.SourceBrandNewDay)
	if err != nil {
		t.Fatalf("expected parser, got error: %v", err)
	}
	if parser == nil {
		t.Fatalf("expected non-nil parser")
	}
}

func TestCreateDailyParser_UnknownSource(t *testing.T) {
	t.Parallel()

	_, err := CreateDailyParser(marketdata.SourceAlphaVantage)
	if err == nil {
		t.Fatalf("expected unsupported source error")
	}
}
