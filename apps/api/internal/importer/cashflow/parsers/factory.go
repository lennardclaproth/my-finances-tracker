package parsers

import (
	"fmt"

	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
)

func CreateCsvParser(ID vendor.VendorID) (cashflow.CsvParser, error) {
	switch ID {
	case vendor.VendorING:
		return NewIngParser(), nil
	case vendor.VendorDEGIRO:
		return NewDegiroParser(), nil
	case vendor.VendorN26:
		return NewN26Parser(), nil
	default:
		return nil, fmt.Errorf("unsupported vendor ID: %s", ID)
	}
}
