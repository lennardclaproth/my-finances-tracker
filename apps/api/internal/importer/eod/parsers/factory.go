package parsers

import (
	"fmt"

	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
)

func CreateEODParser(source marketdata.Source) (EODParser, error) {
	switch source {
	case marketdata.SourceBrandNewDay:
		return NewBrandNewDayParser(), nil
	default:
		return nil, fmt.Errorf("unsupported listing source parser: %s", source)
	}
}
