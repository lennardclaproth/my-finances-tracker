package vendor

import "context"

// CreateHandler creates supported vendors.
type CreateHandler struct {
	creator VendorCreator
}

// NewCreateHandler returns a vendor create handler.
func NewCreateHandler(creator VendorCreator) *CreateHandler {
	return &CreateHandler{creator: creator}
}

// Handle validates and creates a vendor for bootstrap and setup flows.
func (h *CreateHandler) Handle(ctx context.Context, vID VendorID, vtype VendorType) error {
	v, err := NewVendor(vID, vtype)
	if err != nil {
		return err
	}
	return h.creator.Create(ctx, v)
}
