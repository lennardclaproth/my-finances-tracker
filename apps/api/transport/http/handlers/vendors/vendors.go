package vendors

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
	httpx "github.com/lennardclaproth/my-finances-tracker/transport/http"
)

// VendorResponse represents one vendor record.
type VendorResponse struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	Active         bool      `json:"active"`
	ImportDisabled bool      `json:"import_disabled"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// List lists active vendors for import selection.
//
// @Summary List vendors
// @Description Lists active vendors for imports.
// @Tags vendors
// @Accept json
// @Produce json
// @Success 200 {array} VendorResponse
// @Failure 500 {object} map[string]string
// @Router /vendors [get]
func List(
	log logging.Logger,
	queries *vendor.Queries,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vendors, err := queries.ListActive(r.Context())
		if err != nil {
			log.Error(r.Context(), "list vendors: failed to list active vendors", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to list vendors"})
			return
		}

		response := make([]VendorResponse, 0, len(vendors))
		for _, v := range vendors {
			if v == nil {
				continue
			}
			response = append(response, VendorResponse{
				ID:             v.ID,
				Name:           string(v.Name),
				Type:           string(v.Type),
				Active:         v.Active,
				ImportDisabled: v.ImportDisabled,
				CreatedAt:      v.CreatedAt,
				UpdatedAt:      v.UpdatedAt,
			})
		}

		_ = httpx.JSONEncode(w, http.StatusOK, response)
	})
}
