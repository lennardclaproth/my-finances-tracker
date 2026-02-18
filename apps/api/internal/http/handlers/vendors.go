package handlers

import (
	"context"
	"net/http"

	"github.com/lennardclaproth/my-finances-tracker/api"
	httpx "github.com/lennardclaproth/my-finances-tracker/internal/http"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
)

// GetVendors lists active vendors for import selection.
func GetVendors(
	log logging.Logger,
	lister vendor.ActiveVendorLister,
) http.Handler {
	endpoint := func(ctx context.Context, _ struct{}) (status int, res []api.VendorResponse, err error) {
		vendors, err := lister.ListActive(ctx)
		if err != nil {
			return http.StatusInternalServerError, nil, err
		}

		response := make([]api.VendorResponse, 0, len(vendors))
		for _, v := range vendors {
			if v == nil {
				continue
			}
			response = append(response, api.VendorResponse{
				ID:        v.ID,
				Name:      string(v.Name),
				Type:      string(v.Type),
				Active:    v.Active,
				CreatedAt: v.CreatedAt,
				UpdatedAt: v.UpdatedAt,
			})
		}
		return http.StatusOK, response, nil
	}

	decodeFn := httpx.DecoderFunc[struct{}](func(r *http.Request) (struct{}, error) {
		return struct{}{}, nil
	})

	return httpx.Endpoint(decodeFn, log, endpoint)
}
