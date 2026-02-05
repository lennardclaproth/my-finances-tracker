package domain

import (
	"time"

	"github.com/google/uuid"
)

type ImportStatus string

const (
	ImportStatusPending    ImportStatus = "pending"
	ImportStatusInProgress ImportStatus = "in_progress"
	ImportStatusCompleted  ImportStatus = "completed"
	ImportStatusFailed     ImportStatus = "failed"
)

type Import struct {
	ID         uuid.UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Vendor     Vendor
	Path       string
	Status     ImportStatus
	StatusMsg  string
	Duplicates int
	TotalRows  int
	Imported   int
	Failed     int
}

func NewImport(vendor Vendor, path string) *Import {
	return &Import{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Vendor:    vendor,
		Path:      path,
		Status:    ImportStatusPending,
		StatusMsg: "",
		TotalRows: 0,
		Imported:  0,
		Failed:    0,
	}
}
