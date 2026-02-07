package vendor

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
)

type VendorID string

const (
	VendorING VendorID = "ING"
)

var SupportedVendors = []VendorID{
	VendorING,
}

type Vendor struct {
	ID        uuid.UUID `db:"id"`
	Name      VendorID  `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

var (
	ErrUnsupportedVendor   = fmt.Errorf("unsupported vendor")
	ErrVendorAlreadyExists = fmt.Errorf("vendor already exists with the given name")
	ErrVendorNotFound      = fmt.Errorf("vendor not found")
)

// Shared interfaces used by multiple use cases

type VendorCreator interface {
	Create(ctx context.Context, vendor *Vendor) error
}

type VendorFetcher interface {
	FetchById(ctx context.Context, id uuid.UUID) (*Vendor, error)
	FetchByName(ctx context.Context, name VendorID) (*Vendor, error)
}

func NewVendor(name VendorID) (*Vendor, error) {
	if !slices.Contains(SupportedVendors, name) {
		return nil, ErrUnsupportedVendor
	}
	v := &Vendor{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	return v, nil
}
