package storage

import (
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/domain"
)

const (
	TableTransactions = "transactions"
	TableVendors      = "vendors"
	TableImports      = "imports"
)

type TransactionRecord struct {
	ID          uuid.UUID                `db:"id"`
	Description string                   `db:"description"`
	Note        string                   `db:"note"`
	Source      string                   `db:"source"`
	AmountCents int64                    `db:"amount_cents"`
	Direction   domain.CashFlowDirection `db:"direction"`
	Date        time.Time                `db:"date"`
	Checksum    string                   `db:"checksum"`
	CreatedAt   time.Time                `db:"created_at"`
	UpdatedAt   time.Time                `db:"updated_at"`
	Tag         string                   `db:"tag"`
	RowNumber   int                      `db:"row_number"`
	Ignored     bool                     `db:"ignored"`
	ImportID    uuid.UUID                `db:"import_id"`
}

type ImportRecord struct {
	ID         uuid.UUID           `db:"id"`
	CreatedAt  time.Time           `db:"created_at"`
	UpdatedAt  time.Time           `db:"updated_at"`
	VendorID   uuid.UUID           `db:"vendor_id"`
	Path       string              `db:"path"`
	Status     domain.ImportStatus `db:"status"`
	StatusMsg  string              `db:"status_msg"`
	Duplicates int                 `db:"duplicates"`
	TotalRows  int                 `db:"total_rows"`
	Imported   int                 `db:"imported"`
	Failed     int                 `db:"failed"`
}

type VendorRecord struct {
	ID        uuid.UUID       `db:"id"`
	Name      domain.VendorID `db:"name"`
	CreatedAt time.Time       `db:"created_at"`
	UpdatedAt time.Time       `db:"updated_at"`
}
