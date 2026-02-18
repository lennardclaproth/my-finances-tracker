package vendor

import "context"

type CreateHandler struct {
	creator VendorCreator
}

func NewCreateHandler(creator VendorCreator) *CreateHandler {
	return &CreateHandler{creator: creator}
}

func (h *CreateHandler) Handle(ctx context.Context, vID VendorID, vtype VendorType) error {
	v, err := NewVendor(vID, vtype)
	if err != nil {
		return err
	}
	return h.creator.Create(ctx, v)
}
