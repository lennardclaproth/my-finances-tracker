package vendor

import "context"

// CreateHandler creates supported vendors.
type Commands struct {
	creator VendorCreator
}

// NewCreateHandler returns a vendor create handler.
func NewCommands(creator VendorCreator) *Commands {
	return &Commands{creator: creator}
}

// Handle validates and creates a vendor for bootstrap and setup flows.
func (h *Commands) Handle(ctx context.Context, vID VendorID, vtype VendorType) error {
	v, err := NewVendor(vID, vtype)
	if err != nil {
		return err
	}
	return h.creator.Create(ctx, v)
}
