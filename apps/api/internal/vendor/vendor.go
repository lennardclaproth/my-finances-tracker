package vendor

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// VendorID is the symbolic vendor identifier used in supported-vendor definitions.
type VendorID string

// VendorType classifies a vendor as bank or brokerage.
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

// SupportedVendors defines the allowed vendor/type combinations.
var SupportedVendors = map[VendorID]VendorType{
	VendorING:    VendorTypeBank,
	VendorDEGIRO: VendorTypeBrokerage,
	VendorN26:    VendorTypeBank,
	VendorBND:    VendorTypeBrokerage,
}

// Vendor represents an import-capable financial data provider.
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
	// ErrUnsupportedVendor indicates the requested vendor is not in SupportedVendors.
	ErrUnsupportedVendor = fmt.Errorf("unsupported vendor")
	// ErrUnsupportedVendorType indicates vendor/type combination mismatch.
	ErrUnsupportedVendorType = fmt.Errorf("unsupported vendor type for the given vendor")
	// ErrVendorAlreadyExists indicates a uniqueness conflict on vendor creation.
	ErrVendorAlreadyExists = fmt.Errorf("vendor already exists with the given name")
	// ErrVendorNotFound indicates the requested vendor does not exist.
	ErrVendorNotFound = fmt.Errorf("vendor not found")
)

// VendorCreator persists vendors.
type VendorCreator interface {
	Create(ctx context.Context, vendor *Vendor) error
}

// VendorFetcher retrieves vendors by ID or name.
type VendorFetcher interface {
	FetchById(ctx context.Context, id uuid.UUID) (*Vendor, error)
	FetchByName(ctx context.Context, name VendorID) (*Vendor, error)
}

// ActiveVendorLister returns active vendors.
type ActiveVendorLister interface {
	ListActive(ctx context.Context) ([]*Vendor, error)
}

// NewVendor validates the vendor identity/type pair and returns an active vendor.
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
