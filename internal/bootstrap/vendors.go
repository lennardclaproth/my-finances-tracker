package bootstrap

import (
	"context"
	"errors"
	"fmt"

	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
)

func Vendors(ctx context.Context, vc vendor.VendorCreator, logger logging.Logger) {
	h := vendor.NewCreateHandler(vc)
	for _, v := range vendor.SupportedVendors {
		err := h.Handle(ctx, string(v))
		if err == nil {
			continue
		}
		if errors.Is(err, vendor.ErrVendorAlreadyExists) {
			logger.Info(ctx, fmt.Sprintf("Vendor %s already exists, skipping creation.", v))
			continue
		}

		err = fmt.Errorf("failed to create vendor %s: %w", v, err)
		panic(err)
	}
}
