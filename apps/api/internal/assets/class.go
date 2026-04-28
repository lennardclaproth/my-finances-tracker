package assets

import (
	"time"

	"github.com/google/uuid"
)

// ClassSource identifies how an asset class is managed.
type ClassSource string

const (
	// ClassSourceManual represents user-managed classes.
	ClassSourceManual ClassSource = "MANUAL"
	// ClassSourcePortfolio represents the read-only portfolio-linked class.
	ClassSourcePortfolio ClassSource = "PORTFOLIO"
)

// Class groups related assets (for example, property or savings).
type Class struct {
	ID        uuid.UUID   `db:"id"`
	AccountID uuid.UUID   `db:"account_id"`
	Assets    []Asset     `db:"-"`
	Name      string      `db:"name"`
	Source    ClassSource `db:"source"`
	Archived  bool        `db:"archived"`
	CreatedAt time.Time   `db:"created_at"`
	UpdatedAt time.Time   `db:"updated_at"`
}
