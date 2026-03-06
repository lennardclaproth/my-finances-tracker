package parsers

import (
	"io"
	"strings"
	"testing"
)

func TestBrandNewDayParser_ParseAll_TabSeparatedHappyPath(t *testing.T) {
	t.Parallel()

	input := "Date\tNAV\tAsk\tBid\tDividend\n" +
		"28/02/2026\t38,286191\t38,286191\t38,286191\t-\n" +
		"27/02/2026\t38,286065\t38,286065\t38,286065\t-\n"

	parser := NewBrandNewDayParser()
	res, err := parser.ParseAll(io.NopCloser(strings.NewReader(input)))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if res.TotalRows != 2 {
		t.Fatalf("expected total rows 2, got %d", res.TotalRows)
	}
	if len(res.RowErrors) != 0 {
		t.Fatalf("expected no row errors, got %+v", res.RowErrors)
	}
	if len(res.Rows) != 2 {
		t.Fatalf("expected 2 parsed rows, got %d", len(res.Rows))
	}
	if res.Rows[0].Close != 38.286191 {
		t.Fatalf("expected first NAV 38.286191, got %f", res.Rows[0].Close)
	}
	if res.Rows[0].Open != res.Rows[0].Close || res.Rows[0].High != res.Rows[0].Close || res.Rows[0].Low != res.Rows[0].Close {
		t.Fatalf("expected OHLC equal to NAV")
	}
}

func TestBrandNewDayParser_ParseAll_SemicolonSeparated(t *testing.T) {
	t.Parallel()

	input := "Date;NAV;Ask;Bid;Dividend\n" +
		"26/02/2026;38,365916;38,365916;38,365916;-\n"

	parser := NewBrandNewDayParser()
	res, err := parser.ParseAll(io.NopCloser(strings.NewReader(input)))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if len(res.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(res.Rows))
	}
}

func TestBrandNewDayParser_ParseAll_InvalidDateAndNavAreSkipped(t *testing.T) {
	t.Parallel()

	input := "Date\tNAV\tAsk\tBid\tDividend\n" +
		"2026-02-28\t38,286191\t38,286191\t38,286191\t-\n" +
		"27/02/2026\t-\t38,286065\t38,286065\t-\n" +
		"26/02/2026\t38,365916\t38,365916\t38,365916\t-\n"

	parser := NewBrandNewDayParser()
	res, err := parser.ParseAll(io.NopCloser(strings.NewReader(input)))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if res.TotalRows != 3 {
		t.Fatalf("expected total rows 3, got %d", res.TotalRows)
	}
	if len(res.RowErrors) != 2 {
		t.Fatalf("expected 2 row errors, got %+v", res.RowErrors)
	}
	if len(res.Rows) != 1 {
		t.Fatalf("expected only one valid row, got %d", len(res.Rows))
	}
}

func TestBrandNewDayParser_ParseAll_DuplicateDateInFile(t *testing.T) {
	t.Parallel()

	input := "Date\tNAV\tAsk\tBid\tDividend\n" +
		"28/02/2026\t38,286191\t38,286191\t38,286191\t-\n" +
		"28/02/2026\t38,300000\t38,300000\t38,300000\t-\n"

	parser := NewBrandNewDayParser()
	res, err := parser.ParseAll(io.NopCloser(strings.NewReader(input)))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if len(res.Rows) != 1 {
		t.Fatalf("expected first duplicate date to be kept once, got %d rows", len(res.Rows))
	}
	if len(res.RowErrors) != 1 {
		t.Fatalf("expected one row error for duplicate date, got %+v", res.RowErrors)
	}
}

func TestBrandNewDayParser_ParseAll_MissingHeaders(t *testing.T) {
	t.Parallel()

	input := "Date\tNAV\n28/02/2026\t38,286191\n"

	parser := NewBrandNewDayParser()
	_, err := parser.ParseAll(io.NopCloser(strings.NewReader(input)))
	if err == nil {
		t.Fatalf("expected missing header error")
	}
}
