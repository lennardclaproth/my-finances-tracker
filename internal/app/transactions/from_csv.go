package transactions

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/api/contracts"
	"github.com/lennardclaproth/my-finances-tracker/internal/domain"
)

type FromCsvHandler struct {
	creator       TransactionCreator
	vendorFetcher VendorFetcher
	rowParser     CsvParser
}

func NewFromCsvHandler(creator TransactionCreator, vendorFetcher VendorFetcher, rowParser CsvParser) *FromCsvHandler {
	return &FromCsvHandler{creator: creator, vendorFetcher: vendorFetcher, rowParser: rowParser}
}

// Handle processes the CSV data from the provided reader and returns an import summary.
// The csv should have a header row that consists of the following columns:
// Description, Note, Source, Amount, Date
func (h *FromCsvHandler) Handle(ctx context.Context, r io.Reader, vendorId, importId uuid.UUID) (contracts.ImportSummary, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	reader.Comma = ';' // Expect semicolon-separated values
	var summary contracts.ImportSummary
	// Read headerRow row from csv and make sure that the required columns are present
	// in the csv headerRow.
	headerRow, err := reader.Read()
	if err == io.EOF {
		return summary, fmt.Errorf("csv is empty, expected header row")
	}
	if err != nil {
		return summary, fmt.Errorf("reading header row: %w", err)
	}
	err = h.rowParser.ParseHeader(headerRow)
	if err != nil {
		return summary, fmt.Errorf("parsing header row: %w", err)
	}
	// Read each subsequent row and process it into a Transaction.
	// Rows should be correctly handled, these checks include:
	// - missing columns
	// - empty rows
	// - invalid amount format
	// - invalid date format
	rowNumber := 1 // Start at 1 to account for header row
	for {
		// Check for context cancellation
		if err := ctx.Err(); err != nil {
			return summary, err
		}
		// Read next row
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		rowNumber++
		// Handle read error
		if err != nil {
			addRowError(&summary, rowNumber, "reading row: %v", err)
			continue
		}
		// Skip empty lines
		if len(record) == 1 && record[0] == "" {
			continue
		}
		// update total rows count considering we can conclude that
		// this row is not an empty line.
		summary.TotalRows++
		if len(record) < len(headerRow) {
			addRowError(&summary, rowNumber, "row has %d columns, expected at least %d", len(record), len(headerRow))
			continue
		}
		txd, err := h.rowParser.ParseRow(record, rowNumber, importId)
		if err != nil {

			continue
		}
		tx, err := domain.NewTransaction(
			txd.Description,
			txd.Note,
			txd.Source,
			txd.Direction,
			txd.Amount,
			txd.Date,
			rowNumber,
			importId,
		)
		// Save transaction
		err = h.creator.Create(ctx, tx)
		if err != nil {
			if err == domain.ErrDuplicateTransaction {
				summary.Duplicates++
			}
			addRowError(&summary, rowNumber, "saving transaction: %v", err)
			continue
		}
		summary.Imported++
	}
	return summary, nil
}

// // parseHeader parses the CSV header and returns a map of fields with column names to their indices.
// func parseHeader(header []string, vendor *domain.Vendor) (map[domain.Field]int, error) {
// 	// Create a map of header columns to their indices, we use the Field of the
// 	// vendor mapping to ensure we have the correct columns.
	
// 	headerIndex := make(map[domain.Field]int, len(domain.SupportedVendorFields))
// 	err := vendor.SetHeader(header)
// 	if err != nil {
// 		return nil, fmt.Errorf("setting vendor header: %w", err)
// 	}
// 	return headerIndex, nil
// }

// addRowError increments Failed and appends a formatted RowError.
func addRowError(summary *contracts.ImportSummary, row int, format string, args ...any) {
	summary.Failed++
	summary.RowErrors = append(summary.RowErrors, contracts.RowError{
		Row:     row,
		Message: fmt.Sprintf(format, args...),
	})
}
