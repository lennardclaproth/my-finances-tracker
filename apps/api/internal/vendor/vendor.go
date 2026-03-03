package vendor

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type VendorID string
type VendorType string

const (
	VendorTypeBrokerage VendorType = "brokerage"
	VendorTypeBank      VendorType = "bank"
)

const (
	VendorING    VendorID = "ING"
	VendorDEGIRO VendorID = "DEGIRO"
	VendorN26    VendorID = "N26"
	VendorBND    VendorID = "BrandNewDay"
)

var SupportedVendors = map[VendorID]VendorType{
	VendorING:    VendorTypeBank,
	VendorDEGIRO: VendorTypeBrokerage,
	VendorN26:    VendorTypeBank,
	VendorBND:    VendorTypeBrokerage,
}

type Vendor struct {
	ID             uuid.UUID  `db:"id"`
	Name           VendorID   `db:"name"`
	Active         bool       `db:"active"`
	ImportDisabled bool       `db:"import_disabled"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"`
	Type           VendorType `db:"type"`
}

var (
	ErrUnsupportedVendor     = fmt.Errorf("unsupported vendor")
	ErrUnsupportedVendorType = fmt.Errorf("unsupported vendor type for the given vendor")
	ErrVendorAlreadyExists   = fmt.Errorf("vendor already exists with the given name")
	ErrVendorNotFound        = fmt.Errorf("vendor not found")
)

// Shared interfaces used by multiple use cases

type VendorCreator interface {
	Create(ctx context.Context, vendor *Vendor) error
}

type VendorFetcher interface {
	FetchById(ctx context.Context, id uuid.UUID) (*Vendor, error)
	FetchByName(ctx context.Context, name VendorID) (*Vendor, error)
}

type ActiveVendorLister interface {
	ListActive(ctx context.Context) ([]*Vendor, error)
}

func NewVendor(name VendorID, vendorType VendorType) (*Vendor, error) {
	supportedType, supported := SupportedVendors[name]
	if !supported {
		return nil, ErrUnsupportedVendor
	}
	if vendorType != supportedType {
		return nil, ErrUnsupportedVendorType
	}
	v := &Vendor{
		ID:             uuid.New(),
		Name:           name,
		Active:         true,
		ImportDisabled: false,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
		Type:           vendorType,
	}
	return v, nil
}
