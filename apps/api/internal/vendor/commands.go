package vendor

import "context"

// Commands creates supported vendors.
type Commands struct {
	creator VendorCreator
}

// NewCommands returns vendor write-side use cases.
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
