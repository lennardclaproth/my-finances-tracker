package assets

import (
	"strings"
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

const (
	// PortfolioClassName is the fixed name for the portfolio-linked class.
	PortfolioClassName = "Portfolio"
	// PortfolioAssetName is the fixed asset name for the portfolio-linked class.
	PortfolioAssetName = "Portfolio Worth"
)

// Class groups related assets (for example, property or savings).
type Class struct {
	ID        uuid.UUID   `db:"id"`
	AccountID uuid.UUID   `db:"account_id"`
	Assets    []Asset     `db:"-"`
	Mutations []Mutation  `db:"-"`
	Name      string      `db:"name"`
	Source    ClassSource `db:"source"`
	Archived  bool        `db:"archived"`
	CreatedAt time.Time   `db:"created_at"`
	UpdatedAt time.Time   `db:"updated_at"`
}

func NewClass(accID uuid.UUID, srcI *ClassSource, nRaw string) (*Class, error) {
	// If srcI is not set we default ClassSourceManual
	src := ClassSourceManual
	if srcI != nil {
		src = *srcI
	}
	// Guard against illegal classnames
	name := strings.TrimSpace(nRaw)
	if name == "" {
		return nil, ErrClassNameEmpty
	}
	if strings.EqualFold(name, PortfolioClassName) {
		return nil, ErrClassReserved
	}
	// Set now time.
	now := time.Now()
	return &Class{
		ID:        uuid.New(),
		AccountID: accID,
		Name:      name,
		Source:    src,
		Archived:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (c *Class) Update(name *string, archived *bool) error {
	// TODO: fix potential issue here
	if archived != nil {
		c.Archived = *archived
	}

	if name != nil {
		n := strings.TrimSpace(*name)
		if n == "" {
			return ErrClassNameEmpty
		}
		if strings.EqualFold(n, PortfolioClassName) {
			return ErrClassReserved
		}
		c.Name = *name
	}
	return nil
}
