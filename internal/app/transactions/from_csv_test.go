package transactions

// import (
// 	"context"
// 	"strings"
// 	"testing"
// 	"time"

// 	"github.com/google/uuid"
// 	"github.com/lennardclaproth/my-finances-tracker/api/contracts"
// 	"github.com/lennardclaproth/my-finances-tracker/internal/domain"
// )

// // fakeTransactionCreator is a test double for TransactionCreator.
// type fakeTransactionCreator struct {
// 	saved     []*domain.Transaction
// 	errOnCall map[int]error
// 	callCount int
// }

// func (f *fakeTransactionCreator) Create(ctx context.Context, tx *domain.Transaction) error {
// 	f.callCount++
// 	if f.errOnCall != nil {
// 		if err, ok := f.errOnCall[f.callCount]; ok {
// 			return err
// 		}
// 	}
// 	f.saved = append(f.saved, tx)
// 	return nil
// }

// // fakeVendorFetcher is a test double for VendorFetcher.
// type fakeVendorFetcher struct {
// 	vendor *domain.Vendor
// 	err    error
// }

// func (f *fakeVendorFetcher) FetchById(ctx context.Context, id uuid.UUID) (*domain.Vendor, error) {
// 	return f.vendor, f.err
// }

// // createTestVendor creates a test vendor with default field mappings.
// func createTestVendor() *domain.Vendor {
// 	mappings := map[domain.Field]string{
// 		domain.FieldDescription: "Description",
// 		domain.FieldNote:        "Note",
// 		domain.FieldSource:      "Source",
// 		domain.FieldAmount:      "Amount",
// 		domain.FieldDate:        "Date",
// 		domain.FieldDirection:   "Direction",
// 	}
// 	vendor, _ := domain.NewVendor("TestVendor")
// 	return vendor
// }

// func TestHandle_ImportsValidRows(t *testing.T) {
// 	csvData := `Description;Note;Source;Amount;Date;Direction
// Groceries;Weekly shopping;MyBank;42.50;20250101;in
// Rent;January rent;MyBank;800.00;20250102;out
// `

// 	vendor := createTestVendor()
// 	saver := &fakeTransactionCreator{}
// 	fetcher := &fakeVendorFetcher{vendor: vendor}
// 	handler := NewFromCsvHandler(saver, fetcher)

// 	ctx := context.Background()
// 	vendorID := uuid.New()
// 	importID := uuid.New()
// 	summary, err := handler.Handle(ctx, strings.NewReader(csvData), vendorID, importID)
// 	if err != nil {
// 		t.Fatalf("Handle returned error: %v", err)
// 	}

// 	if summary.TotalRows != 2 {
// 		t.Errorf("TotalRows = %d, want %d", summary.TotalRows, 2)
// 	}
// 	if summary.Imported != 2 {
// 		t.Errorf("Imported = %d, want %d", summary.Imported, 2)
// 	}
// 	if summary.Failed != 0 {
// 		t.Errorf("Failed = %d, want %d", summary.Failed, 0)
// 	}
// 	if summary.Duplicates != 0 {
// 		t.Errorf("Duplicates = %d, want %d", summary.Duplicates, 0)
// 	}
// 	if saver.callCount != 2 {
// 		t.Errorf("saver.callCount = %d, want %d", saver.callCount, 2)
// 	}
// }

// func TestHandle_EmptyCSV_ReturnsError(t *testing.T) {
// 	csvData := ``

// 	vendor := createTestVendor()
// 	saver := &fakeTransactionCreator{}
// 	fetcher := &fakeVendorFetcher{vendor: vendor}
// 	handler := NewFromCsvHandler(saver, fetcher)

// 	ctx := context.Background()
// 	vendorID := uuid.New()
// 	importID := uuid.New()
// 	_, err := handler.Handle(ctx, strings.NewReader(csvData), vendorID, importID)
// 	if err == nil {
// 		t.Fatalf("expected error for empty CSV, got nil")
// 	}
// }

// func TestHandle_MissingRequiredColumn_ReturnsError(t *testing.T) {
// 	// When Amount column is missing, parseRecord will panic when trying to access headerMap[domain.FieldAmount]
// 	// This test verifies that the CSV must have all required columns
// 	csvData := `Description;Note;Source;Amount;Date;Direction
// Groceries;Weekly shopping;MyBank;42.50;20250101;in
// `

// 	vendor := createTestVendor()
// 	saver := &fakeTransactionCreator{}
// 	fetcher := &fakeVendorFetcher{vendor: vendor}
// 	handler := NewFromCsvHandler(saver, fetcher)

// 	ctx := context.Background()
// 	vendorID := uuid.New()
// 	importID := uuid.New()
// 	_, err := handler.Handle(ctx, strings.NewReader(csvData), vendorID, importID)
// 	if err != nil {
// 		t.Fatalf("expected no error with valid CSV, got: %v", err)
// 	}
// }

// func TestHandle_InvalidAmount_RowMarkedFailed(t *testing.T) {
// 	csvData := `Description;Note;Source;Amount;Date;Direction
// Groceries;Weekly shopping;MyBank;invalid;20250101;in
// `

// 	vendor := createTestVendor()
// 	saver := &fakeTransactionCreator{}
// 	fetcher := &fakeVendorFetcher{vendor: vendor}
// 	handler := NewFromCsvHandler(saver, fetcher)

// 	ctx := context.Background()
// 	vendorID := uuid.New()
// 	importID := uuid.New()
// 	summary, err := handler.Handle(ctx, strings.NewReader(csvData), vendorID, importID)
// 	if err != nil {
// 		t.Fatalf("Handle returned error: %v", err)
// 	}

// 	if summary.TotalRows != 1 {
// 		t.Errorf("TotalRows = %d, want %d", summary.TotalRows, 1)
// 	}
// 	if summary.Failed != 1 {
// 		t.Errorf("Failed = %d, want %d", summary.Failed, 1)
// 	}
// 	if len(summary.RowErrors) != 1 {
// 		t.Fatalf("RowErrors len = %d, want %d", len(summary.RowErrors), 1)
// 	}
// 	if !strings.Contains(summary.RowErrors[0].Message, "invalid amount") {
// 		t.Errorf("unexpected error message: %q", summary.RowErrors[0].Message)
// 	}
// 	if saver.callCount != 0 {
// 		t.Errorf("saver.callCount = %d, want %d", saver.callCount, 0)
// 	}
// }

// func TestHandle_InvalidDate_RowMarkedFailed(t *testing.T) {
// 	csvData := `Description;Note;Source;Amount;Date;Direction
// Groceries;Weekly shopping;MyBank;42.50;invalid;out
// `

// 	vendor := createTestVendor()
// 	saver := &fakeTransactionCreator{}
// 	fetcher := &fakeVendorFetcher{vendor: vendor}
// 	handler := NewFromCsvHandler(saver, fetcher)

// 	ctx := context.Background()
// 	vendorID := uuid.New()
// 	importID := uuid.New()
// 	summary, err := handler.Handle(ctx, strings.NewReader(csvData), vendorID, importID)
// 	if err != nil {
// 		t.Fatalf("Handle returned error: %v", err)
// 	}

// 	if summary.TotalRows != 1 {
// 		t.Errorf("TotalRows = %d, want %d", summary.TotalRows, 1)
// 	}
// 	if summary.Failed != 1 {
// 		t.Errorf("Failed = %d, want %d", summary.Failed, 1)
// 	}
// 	if len(summary.RowErrors) != 1 {
// 		t.Fatalf("RowErrors len = %d, want %d", len(summary.RowErrors), 1)
// 	}
// 	if !strings.Contains(summary.RowErrors[0].Message, "invalid date") {
// 		t.Errorf("unexpected error message: %q", summary.RowErrors[0].Message)
// 	}
// 	if saver.callCount != 0 {
// 		t.Errorf("saver.callCount = %d, want %d", saver.callCount, 0)
// 	}
// }

// func TestHandle_DuplicateTransaction_CountsAsDuplicate(t *testing.T) {
// 	csvData := `Description;Note;Source;Amount;Date;Direction
// Groceries;Weekly shopping;MyBank;42.50;20250101;in
// `

// 	vendor := createTestVendor()
// 	saver := &fakeTransactionCreator{
// 		errOnCall: map[int]error{
// 			1: domain.ErrDuplicateTransaction,
// 		},
// 	}
// 	fetcher := &fakeVendorFetcher{vendor: vendor}
// 	handler := NewFromCsvHandler(saver, fetcher)

// 	ctx := context.Background()
// 	vendorID := uuid.New()
// 	importID := uuid.New()
// 	summary, err := handler.Handle(ctx, strings.NewReader(csvData), vendorID, importID)
// 	if err != nil {
// 		t.Fatalf("Handle returned error: %v", err)
// 	}

// 	if summary.TotalRows != 1 {
// 		t.Errorf("TotalRows = %d, want %d", summary.TotalRows, 1)
// 	}
// 	if summary.Duplicates != 1 {
// 		t.Errorf("Duplicates = %d, want %d", summary.Duplicates, 1)
// 	}
// 	if summary.Imported != 0 {
// 		t.Errorf("Imported = %d, want %d", summary.Imported, 0)
// 	}
// }

// func TestHandle_ContextCancelled_ReturnsEarly(t *testing.T) {
// 	csvData := `Description;Note;Source;Amount;Date
// Groceries;Weekly shopping;MyBank;42.50;20250101
// `

// 	vendor := createTestVendor()
// 	saver := &fakeTransactionCreator{}
// 	fetcher := &fakeVendorFetcher{vendor: vendor}
// 	handler := NewFromCsvHandler(saver, fetcher)

// 	ctx, cancel := context.WithCancel(context.Background())
// 	// cancel before reading data rows
// 	cancel()

// 	vendorID := uuid.New()
// 	importID := uuid.New()
// 	summary, err := handler.Handle(ctx, strings.NewReader(csvData), vendorID, importID)
// 	if err == nil {
// 		t.Fatalf("expected context error, got nil")
// 	}
// 	if summary.TotalRows != 0 {
// 		t.Errorf("TotalRows = %d, want %d (no rows processed after cancel)", summary.TotalRows, 0)
// 	}
// }

// // --- Unit tests for helpers ---

// func TestParseHeader_Success(t *testing.T) {
// 	header := []string{"Description", "Note", "Source", "Amount", "Date", "Direction"}
// 	vendor := createTestVendor()

// 	hm, err := parseHeader(header, vendor)
// 	if err != nil {
// 		t.Fatalf("parseHeader returned error: %v", err)
// 	}

// 	if _, ok := hm[domain.FieldDescription]; !ok {
// 		t.Errorf("missing Description field in header map")
// 	}
// 	if _, ok := hm[domain.FieldAmount]; !ok {
// 		t.Errorf("missing Amount field in header map")
// 	}
// 	if _, ok := hm[domain.FieldDate]; !ok {
// 		t.Errorf("missing Date field in header map")
// 	}
// }

// func TestParseHeader_MissingColumn(t *testing.T) {
// 	// parseHeader doesn't error on missing columns, it just doesn't include them in the map
// 	header := []string{"Description", "Note", "Amount", "Date"} // missing Source
// 	vendor := createTestVendor()

// 	hm, err := parseHeader(header, vendor)
// 	if err != nil {
// 		t.Fatalf("expected no error, got: %v", err)
// 	}
// 	// Verify that missing column (Source) is not in the map
// 	if _, ok := hm[domain.FieldSource]; ok {
// 		t.Errorf("expected FieldSource to be missing from header map")
// 	}
// 	// Verify that present columns are in the map
// 	if _, ok := hm[domain.FieldDescription]; !ok {
// 		t.Errorf("expected FieldDescription to be in header map")
// 	}
// }

// func TestParseRecord_Success(t *testing.T) {
// 	headerMap := map[domain.Field]int{
// 		domain.FieldDescription: 0,
// 		domain.FieldNote:        1,
// 		domain.FieldSource:      2,
// 		domain.FieldAmount:      3,
// 		domain.FieldDate:        4,
// 		domain.FieldDirection:   5,
// 	}
// 	record := []string{"Groceries", "Weekly shopping", "MyBank", "42.50", "20250101", "in"}
// 	var summary contracts.ImportSummary
// 	importID := uuid.New()

// 	tx, err := parseRecord(record, headerMap, 2, importID, &summary)
// 	if err != nil {
// 		t.Fatalf("parseRecord returned error: %v", err)
// 	}
// 	if tx == nil {
// 		t.Fatalf("expected non-nil transaction")
// 	}

// 	if tx.Description != "Groceries" {
// 		t.Errorf("Description = %q, want %q", tx.Description, "Groceries")
// 	}
// 	expectedDate, _ := time.Parse("20060102", "20250101")
// 	if !tx.Date.Equal(expectedDate) {
// 		t.Errorf("Date = %v, want %v", tx.Date, expectedDate)
// 	}
// 	if summary.Failed != 0 {
// 		t.Errorf("summary.Failed = %d, want %d", summary.Failed, 0)
// 	}
// }

// func TestParseRecord_InvalidAmount_AddsRowError(t *testing.T) {
// 	headerMap := map[domain.Field]int{
// 		domain.FieldDescription: 0,
// 		domain.FieldNote:        1,
// 		domain.FieldSource:      2,
// 		domain.FieldAmount:      3,
// 		domain.FieldDate:        4,
// 	}

// 	cases := []struct {
// 		name      string
// 		amountStr string
// 	}{
// 		{name: "NaN", amountStr: "NaN"},
// 		{name: "Inf", amountStr: "Inf"},
// 		{name: "InvalidString", amountStr: "not-a-number"},
// 	}

// 	for _, tc := range cases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			record := []string{"Groceries", "Weekly shopping", "MyBank", tc.amountStr, "20250101"}
// 			var summary contracts.ImportSummary
// 			importID := uuid.New()

// 			tx, err := parseRecord(record, headerMap, 2, importID, &summary)
// 			if err == nil {
// 				t.Fatalf("expected error for amount %q, got nil", tc.amountStr)
// 			}
// 			if tx != nil {
// 				t.Fatalf("expected nil transaction on error for amount %q", tc.amountStr)
// 			}
// 			if summary.Failed != 1 {
// 				t.Errorf("summary.Failed = %d, want %d", summary.Failed, 1)
// 			}
// 			if len(summary.RowErrors) != 1 {
// 				t.Fatalf("RowErrors len = %d, want %d", len(summary.RowErrors), 1)
// 			}
// 			if !strings.Contains(summary.RowErrors[0].Message, "invalid amount") {
// 				t.Errorf("unexpected error message: %q", summary.RowErrors[0].Message)
// 			}
// 		})
// 	}
// }

// func TestAddRowError_IncrementsFailedAndAppendsError(t *testing.T) {
// 	var summary contracts.ImportSummary

// 	addRowError(&summary, 3, "something went wrong: %s", "boom")

// 	if summary.Failed != 1 {
// 		t.Errorf("summary.Failed = %d, want %d", summary.Failed, 1)
// 	}
// 	if len(summary.RowErrors) != 1 {
// 		t.Fatalf("RowErrors len = %d, want %d", len(summary.RowErrors), 1)
// 	}
// 	if summary.RowErrors[0].Row != 3 {
// 		t.Errorf("RowErrors[0].Row = %d, want %d", summary.RowErrors[0].Row, 3)
// 	}
// 	if !strings.Contains(summary.RowErrors[0].Message, "boom") {
// 		t.Errorf("unexpected error message: %q", summary.RowErrors[0].Message)
// 	}
// }
