package vendors

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/app/transactions"
	"github.com/lennardclaproth/my-finances-tracker/internal/domain"
)

// IngParser parses ING CSV files and constructs Transactions.
type IngParser struct {
	headerToField  map[string]domain.Field
	headerToColumn map[string]int
	fieldToHeader  map[domain.Field]string
}

func NewIngParser() *IngParser {
	return &IngParser{
		headerToField: map[string]domain.Field{
			"Date":               domain.FieldDate,
			"Name / Description": domain.FieldDescription,
			"Notifications":      domain.FieldNote,
			"Amount (EUR)":       domain.FieldAmount,
			"Debit/credit":       domain.FieldDirection,
		},
		fieldToHeader: map[domain.Field]string{
			domain.FieldDate:        "Date",
			domain.FieldDescription: "Name / Description",
			domain.FieldNote:        "Notifications",
			domain.FieldAmount:      "Amount (EUR)",
			domain.FieldDirection:   "Debit/credit",
		},
		headerToColumn: make(map[string]int, len(domain.SupportedVendorFields)-1),
	}
}

// SetHeader implements VendorParser. Initializes the header map from a CSV header row.
func (v *IngParser) SetHeader(headers []string) {
	v.headerToColumn = make(map[string]int, len(domain.SupportedVendorFields)-1)
	for i, h := range headers {
		trimmed := strings.TrimSpace(h)
		v.headerToColumn[trimmed] = i
	}
}

// ParseRow implements VendorParser. Parses a single ING CSV row into a Transaction using the header map.
func (v *IngParser) ParseRow(record []string, rowNumber int, importId uuid.UUID) (transactions.TransactionData, error) {
	// We need the column index for fast mapping to the record
	if len(v.headerToColumn) == 0 {
		return transactions.TransactionData{}, fmt.Errorf("header map not initialized")
	}
	// Extract fields
	dateStr := record[v.headerToColumn["Date"]]
	desc := record[v.headerToColumn["Name / Description"]]
	note := record[v.headerToColumn["Notifications"]]
	source := "ING"
	amountStr := record[v.headerToColumn["Amount (EUR)"]]
	directionRaw := record[v.headerToColumn["Debit/credit"]]
	// Parse amount (replace comma with dot)
	amountStr = strings.ReplaceAll(amountStr, ",", ".")
	var amount float64
	if _, err := fmt.Sscanf(amountStr, "%f", &amount); err != nil {
		return transactions.TransactionData{}, fmt.Errorf("invalid amount: %w", err)
	}

	// Parse direction
	var direction domain.CashFlowDirection
	if strings.EqualFold(directionRaw, "Debit") {
		direction = domain.CashOut
	} else if strings.EqualFold(directionRaw, "Credit") {
		direction = domain.CashIn
	} else {
		return transactions.TransactionData{}, fmt.Errorf("invalid direction: %s", directionRaw)
	}

	// Parse date
	parsedDate, err := time.Parse("20060102", dateStr)
	if err != nil {
		return transactions.TransactionData{}, fmt.Errorf("invalid date: %w", err)
	}

	tx := transactions.TransactionData{
		Description: desc,
		Note:        note,
		Source:      source,
		Direction:   direction,
		Amount:      amount,
		Date:        parsedDate}
	return tx, nil
}
