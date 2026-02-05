package bootstrap

import (
	"context"
	"errors"
	"fmt"

	"github.com/lennardclaproth/my-finances-tracker/internal/app/vendors"
	"github.com/lennardclaproth/my-finances-tracker/internal/domain"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
)

func Vendors(ctx context.Context, vc vendors.VendorCreator, logger logging.Logger) {
	h := vendors.NewCreateHandler(vc)
	for _, v := range domain.SupportedVendors {
		err := h.Handle(string(v))
		if err == nil {
			continue
		}
		if errors.Is(err, domain.ErrVendorAlreadyExists) {
			logger.Info(ctx, fmt.Sprintf("Vendor %s already exists, skipping creation.", v))
			continue
		}

		err = fmt.Errorf("failed to create vendor %s: %w", v, err)
		panic(err)
	}
}
