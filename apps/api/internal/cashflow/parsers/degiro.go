package parsers

import (
	"encoding/csv"
	"fmt"
	"io"
	"iter"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
)

// DegiroParser parses DEGIRO CSV files and constructs TransactionData entries.
type DegiroParser struct {
	headerToColumn map[string]int
	accountType    *cashflow.AccountType
}

func NewDegiroParser() *DegiroParser {
	acct := cashflow.AccountTypeBrokerage
	return &DegiroParser{
		headerToColumn: make(map[string]int),
		accountType:    &acct,
	}
}

func (p *DegiroParser) ParseAll(rc io.ReadCloser) (iter.Seq2[int, cashflow.TransactionData], error) {
	csvReader := csv.NewReader(rc)
	csvReader.Comma = ','
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true

	header, err := csvReader.Read()
	if err != nil {
		return nil, err
	}
	if err := p.parseHeader(header); err != nil {
		return nil, err
	}

	seq := func(yield func(int, cashflow.TransactionData) bool) {
		defer rc.Close()
		rowNumber := 1 // first data row after header
		for {
			record, err := csvReader.Read()
			if err == io.EOF {
				return
			}
			if err != nil {
				rowNumber++
				continue
			}
			td, err := p.ParseRow(record, rowNumber, uuid.Nil)
			if err != nil {
				rowNumber++
				continue
			}
			if !yield(rowNumber, td) {
				return
			}
			rowNumber++
		}
	}

	return seq, nil
}

func (p *DegiroParser) parseHeader(headers []string) error {
	p.headerToColumn = make(map[string]int, len(headers))
	for i, h := range headers {
		p.headerToColumn[strings.TrimSpace(h)] = i
	}
	required := []string{"Date", "Value date", "Product", "ISIN", "Description", "Change", "Order Id"}
	for _, key := range required {
		if _, ok := p.headerToColumn[key]; !ok {
			return fmt.Errorf("missing required header: %s", key)
		}
	}
	return nil
}

func (p *DegiroParser) ParseRow(record []string, rowNumber int, importID uuid.UUID) (cashflow.TransactionData, error) {
	if len(p.headerToColumn) == 0 {
		return cashflow.TransactionData{}, fmt.Errorf("header map not initialized")
	}

	valueDateRaw := p.value(record, "Value date")
	if valueDateRaw == "" {
		valueDateRaw = p.value(record, "Date")
	}
	parsedDate, err := time.Parse("02-01-2006", valueDateRaw)
	if err != nil {
		return cashflow.TransactionData{}, fmt.Errorf("invalid date: %w", err)
	}

	amount, err := parseDegiroAmount(p.value(record, "Change"))
	if err != nil {
		// Some DEGIRO exports put change currency in "Change" and the numeric amount in the next unnamed column.
		if fallback, fbErr := p.amountFromAdjacentColumn(record); fbErr == nil {
			amount = fallback
		} else {
			return cashflow.TransactionData{}, fmt.Errorf("invalid amount: %w", err)
		}
	}

	direction := cashflow.CashIn
	if amount < 0 {
		direction = cashflow.CashOut
	}

	product := p.value(record, "Product")
	isin := p.value(record, "ISIN")
	orderID := p.value(record, "Order Id")
	desc := p.value(record, "Description")
	if desc == "" {
		desc = product
	}

	var noteParts []string
	if product != "" {
		noteParts = append(noteParts, product)
	}
	if isin != "" {
		noteParts = append(noteParts, "ISIN:"+isin)
	}
	if orderID != "" {
		noteParts = append(noteParts, "OrderID:"+orderID)
	}

	return cashflow.TransactionData{
		Description: desc,
		Note:        strings.Join(noteParts, " | "),
		Source:      "DEGIRO",
		Direction:   direction,
		Amount:      amount,
		Date:        parsedDate,
		AccountType: p.accountType,
	}, nil
}

func (p *DegiroParser) value(record []string, key string) string {
	idx := p.headerToColumn[key]
	if idx < 0 || idx >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[idx])
}

func parseDegiroAmount(raw string) (float64, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0, fmt.Errorf("empty amount")
	}
	// DEGIRO uses EU number format, e.g. "1.234,56".
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", ".")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func (p *DegiroParser) amountFromAdjacentColumn(record []string) (float64, error) {
	idx := p.headerToColumn["Change"]
	if idx+1 >= len(record) {
		return 0, fmt.Errorf("no adjacent amount column")
	}
	return parseDegiroAmount(record[idx+1])
}
