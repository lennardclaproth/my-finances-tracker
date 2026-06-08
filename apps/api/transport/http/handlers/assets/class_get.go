package assets

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/assets"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	httpx "github.com/lennardclaproth/my-finances-tracker/transport/http"
)

// GetClassesRequest contains query filters for asset classes.
type GetClassesRequest struct {
	AccountID       uuid.UUID `query:"account_id"`
	IncludeArchived bool      `query:"include_archived"`
}

type ClassResponse struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	Source       string     `json:"source"`
	Archived     bool       `json:"archived"`
	CurrentWorth string     `json:"current_worth"`
	LastChangeAt *time.Time `json:"last_change_at,omitempty"`
	GrowthPct    *float64   `json:"growth_pct,omitempty"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// GetAssetClasses returns account-scoped asset classes.
//
// @Summary List asset classes
// @Description Returns asset classes with totals and growth metadata.
// @Tags assets
// @Accept json
// @Produce json
// @Param account_id query string true "Account ID"
// @Param include_archived query bool false "Include archived classes"
// @Success 200 {array} ClassResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assets/classes [get]
func GetClasses(log logging.Logger, queries assets.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.DecodeQuery[GetClassesRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}

			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{
				"error": "invalid query parameters",
			})
			return
		}

		classes, err := queries.ListClasses(r.Context(), req.AccountID, req.IncludeArchived)
		if err != nil {
			log.Error(r.Context(), "An error occurred while executing list classes", err)
			httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "could not list available classes"})
			return
		}
		res := make([]ClassResponse, 0, len(classes))
		for _, class := range classes {
			cRes := ClassResponse{
				ID:           class.ID,
				Name:         class.Name,
				Source:       string(class.Source),
				Archived:     class.Archived,
				CurrentWorth: class.CurrentWorth.String(),
				LastChangeAt: class.LastChangeAt,
				GrowthPct:    class.GrowthPct,
				UpdatedAt:    class.UpdatedAt,
			}
			res = append(res, cRes)
		}

		httpx.JSONEncode(w, http.StatusOK, res)
	})
}

type ClassDetailsResponse struct {
	Class     ClassResponse              `json:"class"`
	Assets    []AssetResponse            `json:"assets"`
	Growth    []ClassGrowthPointResponse `json:"growth"`
	Mutations []MutationResponse         `json:"mutations"`
}

type MutationResponse struct {
	ID              uuid.UUID `json:"id"`
	ItemID          uuid.UUID `json:"item_id"`
	ChangeType      string    `json:"change_type"`
	Direction       *string   `json:"direction,omitempty"`
	Amount          string    `json:"amount"`
	PreviousWorth   string    `json:"previous_worth"`
	NewWorth        string    `json:"new_worth"`
	ClassTotalWorth string    `json:"class_total_worth"`
	EffectiveDate   string    `json:"effective_date"`
	Note            *string   `json:"note,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type AssetResponse struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	CurrentWorth string    `json:"current_worth"`
	Archived     bool      `json:"archived"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ClassGrowthPointResponse struct {
	Date       string `json:"date"`
	TotalWorth string `json:"total_worth"`
}

// GetAssetClassDetails returns class details with items, growth, and mutations.
//
// @Summary Get asset class details
// @Description Returns class details for drawer rendering.
// @Tags assets
// @Accept json
// @Produce json
// @Param class_id path string true "Class ID"
// @Param account_id query string true "Account ID"
// @Success 200 {object} ClassDetailsResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assets/classes/{class_id} [get]
func GetClassDetails(log logging.Logger, queries assets.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		classID, err := uuid.Parse(r.PathValue("class_id"))
		if err != nil {
			httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"class_id": "class_id must be a valid UUID"})
			return
		}
		accID, err := uuid.Parse(r.URL.Query().Get("account_id"))
		if err != nil {
			httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"account_id": "account_id must be a valid UUID"})
			return
		}
		cd, err := queries.ClassDetails(r.Context(), classID, accID)
		if err != nil {
			log.Error(r.Context(), "An error occurred while trying to get the class details", err)
			httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "could not get class details"})
		}
		res := toClassDetailsResponse(cd)
		httpx.JSONEncode(w, http.StatusOK, res)
	})
}

func toClassDetailsResponse(details *assets.ClassDetails) ClassDetailsResponse {
	assets := make([]AssetResponse, 0, len(details.Assets))
	for _, asset := range details.Assets {
		assets = append(assets, AssetResponse{
			ID:           asset.ID,
			Name:         asset.Name,
			CurrentWorth: asset.CurrentWorth.String(),
			Archived:     asset.Archived,
			UpdatedAt:    asset.UpdatedAt,
		})
	}
	class := ClassResponse{
		ID:           details.Class.ID,
		Name:         details.Class.Name,
		Source:       string(details.Class.Source),
		Archived:     details.Class.Archived,
		CurrentWorth: details.Class.CurrentWorth.String(),
		LastChangeAt: details.Class.LastChangeAt,
		GrowthPct:    details.Class.GrowthPct,
		UpdatedAt:    details.Class.UpdatedAt,
	}
	mutations := make([]MutationResponse, 0, len(details.Mutations))
	for _, mutation := range details.Mutations {
		var direction *string
		if mutation.Direction != nil {
			value := string(*mutation.Direction)
			direction = &value
		}
		mutations = append(mutations, MutationResponse{
			ID:              mutation.ID,
			ItemID:          mutation.AssetID,
			ChangeType:      string(mutation.ChangeType),
			Direction:       direction,
			Amount:          mutation.Amount.String(),
			PreviousWorth:   mutation.PreviousWorth.String(),
			NewWorth:        mutation.NewWorth.String(),
			ClassTotalWorth: mutation.ClassTotalWorth.String(),
			EffectiveDate:   mutation.EffectiveDate.Format("2006-01-02"),
			Note:            mutation.Note,
			CreatedAt:       mutation.CreatedAt,
		})
	}
	growthPnts := make([]ClassGrowthPointResponse, 0, len(details.Growth))
	for _, point := range details.Growth {
		growthPnts = append(growthPnts, ClassGrowthPointResponse{
			Date:       point.Date.Format("2006-01-02"),
			TotalWorth: point.TotalWorth.String(),
		})
	}
	return ClassDetailsResponse{
		Assets:    assets,
		Mutations: mutations,
		Class:     class,
		Growth:    growthPnts,
	}
}
