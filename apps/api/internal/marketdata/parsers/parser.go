package parsers

import (
	"io"
	"time"

	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
)

type DailyRow struct {
	RowNumber int
	Date      time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    int64
}

type ParseResult struct {
	Rows      []DailyRow
	RowErrors []marketdata.DailyUploadRowError
	TotalRows int
}

type DailyParser interface {
	ParseAll(rc io.ReadCloser) (ParseResult, error)
}
