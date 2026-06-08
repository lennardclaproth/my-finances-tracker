package importer

import (
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
)

// ImportType identifies the explicit CSV import workflow selected by the user.
type ImportType string

const (
	// ImportTypeCashflow imports bank/payment CSV rows into cashflow transactions.
	ImportTypeCashflow ImportType = "cashflow"
	// ImportTypePortfolio imports brokerage CSV rows into portfolio transactions.
	ImportTypePortfolio ImportType = "portfolio"
	// ImportTypeEOD imports end-of-day market data CSV rows into EOD records.
	ImportTypeEOD ImportType = "eod"
)

// ImportStatus represents the lifecycle status of an import.
type ImportStatus string

const (
	// ImportStatusPending marks imports waiting to be processed.
	ImportStatusPending ImportStatus = "pending"
	// ImportStatusProcessing marks imports claimed by a worker.
	ImportStatusProcessing ImportStatus = "processing"
	// ImportStatusInProgress is the legacy status used by existing storage wiring.
	ImportStatusInProgress ImportStatus = "in_progress"
	// ImportStatusCompleted marks imports that finished processing.
	ImportStatusCompleted ImportStatus = "completed"
	// ImportStatusFailed marks imports that terminated with a fatal error.
	ImportStatusFailed ImportStatus = "failed"
)

// Import is the durable record for one uploaded import file and processing outcome.
type Import struct {
	ID         uuid.UUID    `db:"id"`
	Type       ImportType   `db:"type"`
	CreatedAt  time.Time    `db:"created_at"`
	UpdatedAt  time.Time    `db:"updated_at"`
	VendorID   uuid.UUID    `db:"vendor_id"`
	AccountID  *uuid.UUID   `db:"account_id"`
	ListingID  *uuid.UUID   `db:"listing_id"`
	Path       string       `db:"path"`
	Source     string       `db:"source"`
	Status     ImportStatus `db:"status"`
	StatusMsg  string       `db:"status_msg"`
	Duplicates int          `db:"duplicates"`
	TotalRows  int          `db:"total_rows"`
	Imported   int          `db:"imported"`
	Failed     int          `db:"failed"`
}

// ImportFileWriter persists an uploaded CSV payload and returns its stored path.
type ImportFileWriter interface {
	WriteCsv(r io.Reader) (string, error)
}

// FileRemover deletes a stored file, used to clean up after a rejected import.
type FileRemover interface {
	Remove(path string) error
}

// ImportOption customizes a new import record.
type ImportOption func(*Import)

// WithImportVendorID sets the source vendor for vendor-backed imports.
func WithImportVendorID(id uuid.UUID) ImportOption {
	return func(imp *Import) {
		imp.VendorID = id
	}
}

// WithImportAccountID sets the account scope for account-backed imports.
func WithImportAccountID(id uuid.UUID) ImportOption {
	return func(imp *Import) {
		if id == uuid.Nil {
			return
		}
		accountID := id
		imp.AccountID = &accountID
	}
}

// WithImportListingID sets the listing scope for EOD imports.
func WithImportListingID(id uuid.UUID) ImportOption {
	return func(imp *Import) {
		if id == uuid.Nil {
			return
		}
		listingID := id
		imp.ListingID = &listingID
	}
}

// WithImportSource stores the selected source/vendor label for parser resolution.
func WithImportSource(source string) ImportOption {
	return func(imp *Import) {
		imp.Source = source
	}
}

// NewTypedImport creates a pending import record for an explicit import type.
func NewTypedImport(importType ImportType, path string, options ...ImportOption) (*Import, error) {
	if path == "" {
		return nil, ErrImportPathRequired
	}
	switch importType {
	case ImportTypeCashflow, ImportTypePortfolio, ImportTypeEOD:
	case "":
		return nil, ErrImportTypeRequired
	default:
		return nil, ErrUnsupportedImportType
	}

	now := time.Now().UTC()
	imp := &Import{
		ID:        uuid.New(),
		Type:      importType,
		CreatedAt: now,
		UpdatedAt: now,
		Path:      path,
		Status:    ImportStatusPending,
		StatusMsg: "",
		TotalRows: 0,
		Imported:  0,
		Failed:    0,
	}
	for _, option := range options {
		option(imp)
	}
	return imp, nil
}

// NewImport creates a pending cashflow import record for a vendor and file path.
func NewImport(v vendor.Vendor, path string, accountID ...uuid.UUID) *Import {
	options := []ImportOption{
		WithImportVendorID(v.ID),
		WithImportSource(string(v.Name)),
	}
	if len(accountID) > 0 && accountID[0] != uuid.Nil {
		options = append(options, WithImportAccountID(accountID[0]))
	}
	imp, err := NewTypedImport(ImportTypeCashflow, path, options...)
	if err != nil {
		return nil
	}
	return imp
}

// MarkProcessing transitions the import to processing state.
func (imp *Import) MarkProcessing() {
	imp.Status = ImportStatusProcessing
	imp.UpdatedAt = time.Now().UTC()
}

// MarkCompleted transitions the import to completed state and stores result counters.
func (imp *Import) MarkCompleted(duplicates, totalRows, imported, failed int) {
	imp.Status = ImportStatusCompleted
	imp.UpdatedAt = time.Now().UTC()
	imp.Duplicates = duplicates
	imp.TotalRows = totalRows
	imp.Imported = imported
	imp.Failed = failed
}

// MarkFailed transitions the import to failed state and stores the failure message.
func (imp *Import) MarkFailed(statusMsg string) {
	imp.Status = ImportStatusFailed
	imp.UpdatedAt = time.Now().UTC()
	imp.StatusMsg = statusMsg
}

// RequireAccountID returns the import's account ID, or ErrAccountIDRequired when it is absent.
func (imp *Import) RequireAccountID() (uuid.UUID, error) {
	if imp == nil || imp.AccountID == nil || *imp.AccountID == uuid.Nil {
		return uuid.Nil, ErrAccountIDRequired
	}
	return *imp.AccountID, nil
}

// RequireListingID returns the import's listing ID, or ErrListingIDRequired when it is absent.
func (imp *Import) RequireListingID() (uuid.UUID, error) {
	if imp == nil || imp.ListingID == nil || *imp.ListingID == uuid.Nil {
		return uuid.Nil, ErrListingIDRequired
	}
	return *imp.ListingID, nil
}
