package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"iter"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/transaction"
)

// IngParser parses ING CSV files and constructs Transactions.
type IngParser struct {
	headerToColumn map[string]int
}

func NewIngParser() *IngParser {
	return &IngParser{
		headerToColumn: make(map[string]int),
	}
}

func (p *IngParser) ParseAll(rc io.ReadCloser) (iter.Seq2[int, transaction.TransactionData], error) {
	csvReader := csv.NewReader(rc)
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
	seq := func(yield func(int, transaction.TransactionData) bool) {
		defer rc.Close()
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
func (p *IngParser) ParseRow(record []string, rowNumber int, importId uuid.UUID) (transaction.TransactionData, error) {
	if len(p.headerToColumn) == 0 {
		return transaction.TransactionData{}, fmt.Errorf("header map not initialized")
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
		return transaction.TransactionData{}, fmt.Errorf("invalid amount: %w", err)
	}

	// Parse direction
	var direction transaction.CashFlowDirection
	if strings.EqualFold(directionRaw, "Debit") {
		direction = transaction.CashOut
	} else if strings.EqualFold(directionRaw, "Credit") {
		direction = transaction.CashIn
	} else {
		return transaction.TransactionData{}, fmt.Errorf("invalid direction: %s", directionRaw)
	}

	// Parse date
	parsedDate, err := time.Parse("20060102", dateStr)
	if err != nil {
		return transaction.TransactionData{}, fmt.Errorf("invalid date: %w", err)
	}

	return transaction.TransactionData{
		Description: desc,
		Note:        note,
		Source:      source,
		Direction:   direction,
		Amount:      amount,
		Date:        parsedDate,
	}, nil
}
