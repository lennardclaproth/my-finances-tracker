package importer

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
)

var (
	// ErrNoImportsPending indicates no import rows matched the query.
	ErrNoImportsPending = fmt.Errorf("no imports pending")
	// ErrAccountIDRequired indicates imports require a non-nil account ID.
	ErrAccountIDRequired = fmt.Errorf("account_id is required for imports")
	// ErrImportAccountNotFound indicates the importer account projection does not contain the account.
	ErrImportAccountNotFound = fmt.Errorf("import account not found")
	// ErrImportAccountValidationNotReady indicates account projection dependencies are not configured.
	ErrImportAccountValidationNotReady = fmt.Errorf("account validation is not configured")
	// ErrVendorImportDisabled indicates the selected vendor is not import-enabled.
	ErrVendorImportDisabled = fmt.Errorf("vendor import disabled")
)

// ImportStatus represents the lifecycle status of an import.
type ImportStatus string

const (
	// ImportStatusPending marks imports waiting to be processed.
	ImportStatusPending ImportStatus = "pending"
	// ImportStatusInProgress marks imports claimed by a worker.
	ImportStatusInProgress ImportStatus = "in_progress"
	// ImportStatusCompleted marks imports that finished processing.
	ImportStatusCompleted ImportStatus = "completed"
	// ImportStatusFailed marks imports that terminated with a fatal error.
	ImportStatusFailed ImportStatus = "failed"
)

// Import is the durable record for one uploaded import file and processing outcome.
type Import struct {
	ID         uuid.UUID    `db:"id"`
	CreatedAt  time.Time    `db:"created_at"`
	UpdatedAt  time.Time    `db:"updated_at"`
	VendorID   uuid.UUID    `db:"vendor_id"`
	AccountID  *uuid.UUID   `db:"account_id"`
	Path       string       `db:"path"`
	Status     ImportStatus `db:"status"`
	StatusMsg  string       `db:"status_msg"`
	Duplicates int          `db:"duplicates"`
	TotalRows  int          `db:"total_rows"`
	Imported   int          `db:"imported"`
	Failed     int          `db:"failed"`
}

// ImportCreator persists new import records.
type ImportCreator interface {
	Create(ctx context.Context, imp *Import) error
}

// ImportEnqueuer schedules import IDs for asynchronous processing.
type ImportEnqueuer interface {
	Enqueue(ctx context.Context, importID uuid.UUID) error
}

// NewImport creates a pending import record for a vendor and file path.
func NewImport(v vendor.Vendor, path string, accountID ...uuid.UUID) *Import {
	var accID *uuid.UUID
	if len(accountID) > 0 && accountID[0] != uuid.Nil {
		id := accountID[0]
		accID = &id
	}

	return &Import{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		VendorID:  v.ID,
		AccountID: accID,
		Path:      path,
		Status:    ImportStatusPending,
		StatusMsg: "",
		TotalRows: 0,
		Imported:  0,
		Failed:    0,
	}
}

// MarkInProgress transitions the import to in-progress state.
func (imp *Import) MarkInProgress() {
	imp.Status = ImportStatusInProgress
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
