package parsers

import (
	"fmt"

	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
)

func CreateCsvParser(id vendor.VendorID) (portfolio.CsvParser, error) {
	switch id {
	case vendor.VendorDEGIRO:
		return NewDegiroParser(), nil
	default:
		return nil, fmt.Errorf("unsupported vendor ID for portfolio parser: %s", id)
	}
}
