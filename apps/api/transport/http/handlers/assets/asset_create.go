package assets

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/assets"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
	httpx "github.com/lennardclaproth/my-finances-tracker/transport/http"
)

type CreateAssetRequest struct {
	AccountID     uuid.UUID `json:"account_id"`
	ClassID       uuid.UUID `json:"class_id"`
	Name          string    `json:"name"`
	InitialWorth  string    `json:"initial_worth"`
	EffectiveDate string    `json:"effective_date"`
	Note          *string   `json:"note,omitempty"`
}

type CreateAssetResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func (r CreateAssetRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)
	if strings.TrimSpace(r.Name) == "" {
		problems["name"] = "name is required"
	}

	if name := strings.TrimSpace(r.Name); len(name) > 255 {
		problems["name"] = "name cannot exceed 255 characters"
	}

	initWorth, err := money.ParsePrice(r.InitialWorth)
	if err != nil {
		problems["initial_worth"] = "initial_worth must be a valid decimal string"
	} else if initWorth < 0 {
		problems["initial_worth"] = "initial_worth cannot be negative"
	}

	_, err = time.Parse("2006-01-02", r.EffectiveDate)
	if err != nil {
		problems["effective_date"] = "effective_date must be in YYYY-MM-DD format"
	}

	return len(problems) == 0, problems
}

// CreateAsset creates a tracked asset in a manual class.
//
// @Summary Create asset
// @Description Adds a tracked asset and records its initial worth.
// @Tags assets
// @Accept json
// @Produce json
// @Param request body CreateAssetRequest true "Create asset payload"
// @Success 201 {object} CreateAssetResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assets [post]
func CreateAsset(log logging.Logger, commands assets.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.JSONDecode[CreateAssetRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}

			httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
			return
		}
		isValid, problems := req.isValid()
		if !isValid {
			httpx.JSONEncode(w, http.StatusBadRequest, problems)
			return
		}
		initWorth, _ := money.ParsePrice(req.InitialWorth)
		// if err != nil {
		// 	log.Error(r.Context(), "create asset: initial worth failed to parse, this error should not occur", err)
		// 	httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid initial worth format"})
		// 	return
		// }
		effectiveDate, _ := time.Parse("2006-01-02", req.EffectiveDate)
		// if err != nil {
		// 	log.Error(r.Context(), "create asset: effective date failed to parse, this error should not occur", err)
		// 	httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid effective date format"})
		// 	return
		// }
		asset, err := commands.CreateAsset(r.Context(), req.AccountID, req.ClassID, req.Name, initWorth, effectiveDate, req.Note)
		if err != nil {
			log.Error(r.Context(), "failed to create asset", err)
			httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to create asset"})
			return
		}
		httpx.JSONEncode(w, http.StatusCreated, CreateAssetResponse{
			ID:   asset.ID,
			Name: asset.Name,
		})
	})
}
