package vendors

import "github.com/lennardclaproth/my-finances-tracker/internal/domain"

type CreateHandler struct {
	creator VendorCreator
}

func NewCreateHandler(creator VendorCreator) *CreateHandler {
	return &CreateHandler{creator: creator}
}

func (h *CreateHandler) Handle(name string) error {
	vendor, err := domain.NewVendor(domain.VendorID(name))
	if err != nil {
		return err
	}
	return h.creator.Create(vendor)
}
