package contracts

import (
	"time"

	"github.com/google/uuid"
)

// ImportSummary represents the result of a CSV import operation, containing
// counts of processed, imported, duplicate, and failed rows along with detailed errors.
type ImportSummary struct {
	// TotalRows is the total number of data rows in the CSV file (excluding header)
	TotalRows int `json:"totalRows" example:"100"`
	// Imported is the number of rows successfully imported into the database
	Imported int `json:"imported" example:"98"`
	// Duplicates is the number of rows skipped due to duplicate checksums
	Duplicates int `json:"duplicates" example:"1"`
	// Failed is the number of rows that failed to import
	Failed int `json:"failed" example:"1"`
	// RowErrors contains detailed error information for each failed or problematic row
	RowErrors []RowError `json:"rowErrors"`
}

type Transaction struct {
	ID          uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Description string    `json:"description" example:"Grocery shopping"`
	Note        string    `json:"note" example:"Bought fruits and vegetables"`
	Source      string    `json:"source" example:"MyBank"`
	AmountCents int64     `json:"amountCents" example:"4250"`
	Date        time.Time `json:"date" example:"2025-01-15T00:00:00Z"`
	Tag         string    `json:"tag" example:"Food"`
}
