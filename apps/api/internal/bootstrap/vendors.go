package bootstrap

import (
	"context"
	"errors"
	"fmt"

	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
)

// Vendors bootstraps all supported vendors and skips existing rows.
func Vendors(ctx context.Context, vc vendor.VendorCreator, logger logging.Logger) {
	h := vendor.NewCreateHandler(vc)
	for vname, vtype := range vendor.SupportedVendors {
		err := h.Handle(ctx, vname, vtype)
		if err == nil {
			continue
		}
		if errors.Is(err, vendor.ErrVendorAlreadyExists) {
			logger.Info(
				ctx,
				"vendor already exists, skipping bootstrap create",
				"vendor_name", string(vname),
				"vendor_type", string(vtype),
			)
			continue
		}

		err = fmt.Errorf("failed to create vendor %s: %w", vname, err)
		panic(err)
	}
}
