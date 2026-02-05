package vendors

import "github.com/lennardclaproth/my-finances-tracker/internal/domain"

type VendorCreator interface {
	Create(vendor *domain.Vendor) error
}
