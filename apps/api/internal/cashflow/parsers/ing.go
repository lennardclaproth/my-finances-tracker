package parsers

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"iter"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
)

// IngParser parses ING CSV files and constructs Transactions.
type IngParser struct {
	headerToColumn map[string]int
	accountType    *cashflow.AccountType
}

func NewIngParser() *IngParser {
	acct := cashflow.AccountTypeChecking
	return &IngParser{
		headerToColumn: make(map[string]int),
		accountType:    &acct,
	}
}

func (p *IngParser) ParseAll(rc io.ReadCloser) (iter.Seq2[int, cashflow.TransactionData], error) {
	raw, err := io.ReadAll(rc)
	closeErr := rc.Close()
	if err != nil {
		if closeErr != nil {
			return nil, errors.Join(err, closeErr)
		}
		return nil, err
	}
	if closeErr != nil {
		return nil, closeErr
	}
	csvReader := csv.NewReader(bytes.NewReader(raw))
	csvReader.Comma = ';'
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true
	// Read header
	header, err := csvReader.Read()
	if err != nil {
		// Handle error (e.g., log and return an empty sequence)
		return nil, err
	}
	if err := p.parseHeader(header); err != nil {
		// Handle error (e.g., log and return an empty sequence)
		return nil, err
	}
	// Return an iterator (Seq) that yields TransactionData items
	seq := func(yield func(int, cashflow.TransactionData) bool) {
		rowNumber := 1 // first data row after header
		for {
			record, err := csvReader.Read()
			if err == io.EOF {
				return
			}
			if err != nil {
				// Decide policy: skip bad CSV row reads
				rowNumber++
				continue
			}
			td, err := p.ParseRow(record, rowNumber, uuid.Nil)
			if err != nil {
				// Decide policy: skip rows that fail to parse
				rowNumber++
				continue
			}
			// Respect early-stop from consumer
			if !yield(rowNumber, td) {
				return
			}
			rowNumber++
		}
	}

	return seq, nil
}

// ParseHeader initializes the header map from a CSV header row.
func (p *IngParser) parseHeader(headers []string) error {
	p.headerToColumn = make(map[string]int, len(headers))
	for i, h := range headers {
		trimmed := strings.TrimSpace(h)
		p.headerToColumn[trimmed] = i
	}
	return nil
}

// ParseRow parses a single ING CSV row into a TransactionData.
func (p *IngParser) ParseRow(record []string, rowNumber int, importId uuid.UUID) (cashflow.TransactionData, error) {
	if len(p.headerToColumn) == 0 {
		return cashflow.TransactionData{}, fmt.Errorf("header map not initialized")
	}
	// Extract fields
	dateStr := record[p.headerToColumn["Date"]]
	desc := record[p.headerToColumn["Name / Description"]]
	note := record[p.headerToColumn["Notifications"]]
	source := "ING"
	amountStr := record[p.headerToColumn["Amount (EUR)"]]
	directionRaw := record[p.headerToColumn["Debit/credit"]]

	// Parse amount (replace comma with dot)
	amountStr = strings.ReplaceAll(amountStr, ",", ".")
	var amount float64
	if _, err := fmt.Sscanf(amountStr, "%f", &amount); err != nil {
		return cashflow.TransactionData{}, fmt.Errorf("invalid amount: %w", err)
	}

	// Parse direction
	var direction cashflow.CashFlowDirection
	if strings.EqualFold(directionRaw, "Debit") {
		direction = cashflow.CashOut
	} else if strings.EqualFold(directionRaw, "Credit") {
		direction = cashflow.CashIn
	} else {
		return cashflow.TransactionData{}, fmt.Errorf("invalid direction: %s", directionRaw)
	}

	// Parse date
	parsedDate, err := time.Parse("20060102", dateStr)
	if err != nil {
		return cashflow.TransactionData{}, fmt.Errorf("invalid date: %w", err)
	}

	return cashflow.TransactionData{
		Description: desc,
		Note:        note,
		Source:      source,
		Direction:   direction,
		Amount:      amount,
		Date:        parsedDate,
		AccountType: p.accountType, // default to checking for ING
	}, nil
}
