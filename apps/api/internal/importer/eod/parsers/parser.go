package parsers

import (
	"io"
	"time"
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

type RowError struct {
	RowNumber int
	Reason    string
}

type ParseResult struct {
	Rows      []DailyRow
	RowErrors []RowError
	TotalRows int
}

type DailyParser interface {
	ParseAll(rc io.ReadCloser) (ParseResult, error)
}
