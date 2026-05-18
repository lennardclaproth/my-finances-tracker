package parsers

import (
	"fmt"

	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
)

func CreateDailyParser(source marketdata.Source) (DailyParser, error) {
	switch source {
	case marketdata.SourceBrandNewDay:
		return NewBrandNewDayParser(), nil
	default:
		return nil, fmt.Errorf("unsupported listing source parser: %s", source)
	}
}
