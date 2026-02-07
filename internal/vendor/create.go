package vendor

import "context"

type CreateHandler struct {
	creator VendorCreator
}

func NewCreateHandler(creator VendorCreator) *CreateHandler {
	return &CreateHandler{creator: creator}
}

func (h *CreateHandler) Handle(ctx context.Context, name string) error {
	v, err := NewVendor(VendorID(name))
	if err != nil {
		return err
	}
	return h.creator.Create(ctx, v)
}
