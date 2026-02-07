package importer

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
)

var (
	ErrNoImportsPending = fmt.Errorf("no imports pending")
)

type ImportStatus string

const (
	ImportStatusPending    ImportStatus = "pending"
	ImportStatusInProgress ImportStatus = "in_progress"
	ImportStatusCompleted  ImportStatus = "completed"
	ImportStatusFailed     ImportStatus = "failed"
)

type Import struct {
	ID         uuid.UUID    `db:"id"`
	CreatedAt  time.Time    `db:"created_at"`
	UpdatedAt  time.Time    `db:"updated_at"`
	VendorID   uuid.UUID    `db:"vendor_id"`
	Path       string       `db:"path"`
	Status     ImportStatus `db:"status"`
	StatusMsg  string       `db:"status_msg"`
	Duplicates int          `db:"duplicates"`
	TotalRows  int          `db:"total_rows"`
	Imported   int          `db:"imported"`
	Failed     int          `db:"failed"`
}

// Shared interfaces

type ImportCreator interface {
	Create(ctx context.Context, imp *Import) error
}

func NewImport(v vendor.Vendor, path string) *Import {
	return &Import{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		VendorID:  v.ID,
		Path:      path,
		Status:    ImportStatusPending,
		StatusMsg: "",
		TotalRows: 0,
		Imported:  0,
		Failed:    0,
	}
}

func (imp *Import) MarkInProgress() {
	imp.Status = ImportStatusInProgress
	imp.UpdatedAt = time.Now().UTC()
}

func (imp *Import) MarkCompleted(duplicates, totalRows, imported, failed int) {
	imp.Status = ImportStatusCompleted
	imp.UpdatedAt = time.Now().UTC()
	imp.Duplicates = duplicates
	imp.TotalRows = totalRows
	imp.Imported = imported
	imp.Failed = failed
}

func (imp *Import) MarkFailed(statusMsg string) {
	imp.Status = ImportStatusFailed
	imp.UpdatedAt = time.Now().UTC()
	imp.StatusMsg = statusMsg
}
