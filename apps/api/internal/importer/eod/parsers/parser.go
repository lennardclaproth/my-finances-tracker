package parsers

import (
	"io"
	"time"
)

type EODRow struct {
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
	Rows      []EODRow
	RowErrors []RowError
	TotalRows int
}

type EODParser interface {
	ParseAll(rc io.ReadCloser) (ParseResult, error)
}
