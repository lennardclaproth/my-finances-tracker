package domain

import (
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
)

type Vendor struct {
	ID        uuid.UUID
	Name      VendorID
	CreatedAt time.Time
	UpdatedAt time.Time
}

var (
	ErrVendorFieldNotFound    = fmt.Errorf("vendor field not found in mapping")
	ErrUnsupportedVendorField = fmt.Errorf("unsupported vendor field in mapping")
	ErrUnsupportedVendor      = fmt.Errorf("unsupported vendor")
	ErrVendorAlreadyExists    = fmt.Errorf("vendor already exists with the given name")
	ErrVendorNotFound         = fmt.Errorf("vendor not found")
)

type Field string
type VendorID string

const (
	VendorING VendorID = "ING"
)

const (
	FieldDescription Field = "description"
	FieldNote        Field = "note"
	FieldSource      Field = "source"
	FieldAmount      Field = "amount"
	FieldDate        Field = "date"
	FieldDirection   Field = "direction"
)

var SupportedVendors = []VendorID{
	VendorING,
}

var SupportedVendorFields = []Field{
	FieldDescription,
	FieldNote,
	FieldSource,
	FieldAmount,
	FieldDate,
	FieldDirection,
}

func NewVendor(name VendorID) (*Vendor, error) {
	if slices.Contains(SupportedVendors, name) == false {
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
