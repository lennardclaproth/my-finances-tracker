package parser

import (
	"fmt"
	"io"
	"iter"

	"github.com/lennardclaproth/my-finances-tracker/internal/transaction"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
)

type CsvParser interface {
	ParseAll(rc io.ReadCloser) (iter.Seq2[int, transaction.TransactionData], error)
}

func CreateCsvParser(ID vendor.VendorID) (CsvParser, error) {
	switch ID {
	case vendor.VendorING:
		return NewIngParser(), nil
	default:
		return nil, fmt.Errorf("unsupported vendor ID: %s", ID)
	}
}
