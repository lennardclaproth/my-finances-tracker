package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/api"
	"github.com/lennardclaproth/my-finances-tracker/internal/assets"
	httpx "github.com/lennardclaproth/my-finances-tracker/internal/http"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
)

type assetService interface {
	CreateClass(ctx context.Context, input assets.CreateClassInput) (*assets.Class, error)
	UpdateClass(ctx context.Context, input assets.UpdateClassInput) error
	DeleteClass(ctx context.Context, accountID, classID uuid.UUID) error
	ListClasses(ctx context.Context, accountID uuid.UUID, includeArchived bool) ([]assets.ClassSummary, error)
	ListSnapshots(ctx context.Context, input assets.ListSnapshotsInput) ([]assets.GrowthPoint, error)
	GetClassDetails(ctx context.Context, input assets.ListClassDetailsInput) (*assets.ClassDetails, error)
	CreateItem(ctx context.Context, input assets.CreateItemInput) (*assets.Item, error)
	SetItemWorth(ctx context.Context, input assets.SetItemWorthInput) error
	AdjustItemWorth(ctx context.Context, input assets.AdjustItemWorthInput) error
}

type assetClassPathRequest struct {
	AccountID uuid.UUID
	ClassID   uuid.UUID
}

func (r assetClassPathRequest) Valid(ctx context.Context) map[string]string {
	problems := map[string]string{}
	if r.AccountID == uuid.Nil {
		problems["account_id"] = "account_id is required"
	}
	if r.ClassID == uuid.Nil {
		problems["class_id"] = "class_id is required"
	}
	return problems
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
// @Success 200 {array} api.AssetClassResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assets/classes [get]
func GetAssetClasses(log logging.Logger, svc assetService) http.Handler {
	endpoint := func(ctx context.Context, req api.GetAssetClassesRequest) (status int, res any, err error) {
		rows, err := svc.ListClasses(ctx, req.AccountID, req.IncludeArchived)
		if err != nil {
			if handledStatus, payload, handled := mapAssetServiceError(err); handled {
				return handledStatus, payload, nil
			}
			return http.StatusInternalServerError, struct{}{}, err
		}
		response := make([]api.AssetClassResponse, 0, len(rows))
		for _, row := range rows {
			response = append(response, toAssetClassResponse(row))
		}
		return http.StatusOK, response, nil
	}
	decodeFn := httpx.DecoderFunc[api.GetAssetClassesRequest](func(r *http.Request) (api.GetAssetClassesRequest, error) {
		return httpx.DecodeQuery[api.GetAssetClassesRequest](r)
	})
	return httpx.Endpoint(decodeFn, log, endpoint)
}

// GetAssetSnapshots returns account-level total asset snapshots.
//
// @Summary List asset snapshots
// @Description Returns account-level daily total worth snapshot points.
// @Tags assets
// @Accept json
// @Produce json
// @Param account_id query string true "Account ID"
// @Param from query string false "Start date (YYYY-MM-DD)"
// @Param to query string false "End date (YYYY-MM-DD)"
// @Success 200 {array} api.AssetGrowthPointResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assets/snapshots [get]
func GetAssetSnapshots(log logging.Logger, svc assetService) http.Handler {
	endpoint := func(ctx context.Context, req api.GetAssetSnapshotsRequest) (status int, res any, err error) {
		from, to, err := parseAssetDateRange(req.From, req.To)
		if err != nil {
			return http.StatusBadRequest, assetDateRangeProblem(err), nil
		}
		points, err := svc.ListSnapshots(ctx, assets.ListSnapshotsInput{
			AccountID: req.AccountID,
			From:      from,
			To:        to,
		})
		if err != nil {
			if handledStatus, payload, handled := mapAssetServiceError(err); handled {
				return handledStatus, payload, nil
			}
			return http.StatusInternalServerError, struct{}{}, err
		}
		out := make([]api.AssetGrowthPointResponse, 0, len(points))
		for _, point := range points {
			out = append(out, api.AssetGrowthPointResponse{
				Date:       point.Date.Format("2006-01-02"),
				TotalWorth: formatDecimal(point.TotalWorth.Float64()),
			})
		}
		return http.StatusOK, out, nil
	}
	decodeFn := httpx.DecoderFunc[api.GetAssetSnapshotsRequest](func(r *http.Request) (api.GetAssetSnapshotsRequest, error) {
		return httpx.DecodeQuery[api.GetAssetSnapshotsRequest](r)
	})
	return httpx.Endpoint(decodeFn, log, endpoint)
}

// CreateAssetClass creates a manual asset class.
//
// @Summary Create asset class
// @Description Creates a manual asset class for the selected account.
// @Tags assets
// @Accept json
// @Produce json
// @Param request body api.CreateAssetClassRequest true "Create asset class payload"
// @Success 201 {object} api.AssetClassResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assets/classes [post]
func CreateAssetClass(log logging.Logger, svc assetService) http.Handler {
	endpoint := func(ctx context.Context, req api.CreateAssetClassRequest) (status int, res any, err error) {
		class, err := svc.CreateClass(ctx, assets.CreateClassInput{
			AccountID: req.AccountID,
			Name:      req.Name,
		})
		if err != nil {
			if handledStatus, payload, handled := mapAssetServiceError(err); handled {
				return handledStatus, payload, nil
			}
			return http.StatusInternalServerError, struct{}{}, err
		}
		return http.StatusCreated, toAssetClassResponse(assets.ClassSummary{
			ID:           class.ID,
			Name:         class.Name,
			Source:       class.Source,
			Archived:     class.Archived,
			CurrentWorth: 0,
			LastChangeAt: nil,
			GrowthPct:    nil,
			UpdatedAt:    class.UpdatedAt,
		}), nil
	}
	decodeFn := httpx.DecoderFunc[api.CreateAssetClassRequest](func(r *http.Request) (api.CreateAssetClassRequest, error) {
		return httpx.DecodeJSON[api.CreateAssetClassRequest](r)
	})
	return httpx.Endpoint(decodeFn, log, endpoint)
}

// UpdateAssetClass updates mutable class fields.
//
// @Summary Update asset class
// @Description Updates manual asset class name and archived state.
// @Tags assets
// @Accept json
// @Produce json
// @Param request body api.UpdateAssetClassRequest true "Update asset class payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assets/classes [patch]
func UpdateAssetClass(log logging.Logger, svc assetService) http.Handler {
	endpoint := func(ctx context.Context, req api.UpdateAssetClassRequest) (status int, res any, err error) {
		err = svc.UpdateClass(ctx, assets.UpdateClassInput{
			AccountID: req.AccountID,
			ClassID:   req.ID,
			Name:      req.Name,
			Archived:  req.Archived,
		})
		if err != nil {
			if handledStatus, payload, handled := mapAssetServiceError(err); handled {
				return handledStatus, payload, nil
			}
			return http.StatusInternalServerError, struct{}{}, err
		}
		return http.StatusOK, map[string]string{"status": "updated"}, nil
	}
	decodeFn := httpx.DecoderFunc[api.UpdateAssetClassRequest](func(r *http.Request) (api.UpdateAssetClassRequest, error) {
		return httpx.DecodeJSON[api.UpdateAssetClassRequest](r)
	})
	return httpx.Endpoint(decodeFn, log, endpoint)
}

// DeleteAssetClass deletes a manual class.
//
// @Summary Delete asset class
// @Description Deletes a manual class and its items/history.
// @Tags assets
// @Accept json
// @Produce json
// @Param class_id path string true "Class ID"
// @Param account_id query string true "Account ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assets/classes/{class_id} [delete]
func DeleteAssetClass(log logging.Logger, svc assetService) http.Handler {
	endpoint := func(ctx context.Context, req assetClassPathRequest) (status int, res any, err error) {
		err = svc.DeleteClass(ctx, req.AccountID, req.ClassID)
		if err != nil {
			if handledStatus, payload, handled := mapAssetServiceError(err); handled {
				return handledStatus, payload, nil
			}
			return http.StatusInternalServerError, struct{}{}, err
		}
		return http.StatusOK, map[string]string{"status": "deleted"}, nil
	}
	decodeFn := httpx.DecoderFunc[assetClassPathRequest](func(r *http.Request) (assetClassPathRequest, error) {
		query, err := httpx.DecodeQuery[api.GetAssetClassesRequest](r)
		if err != nil {
			return assetClassPathRequest{}, err
		}
		classID, err := uuid.Parse(r.PathValue("class_id"))
		if err != nil {
			return assetClassPathRequest{}, fmt.Errorf("invalid class_id")
		}
		return assetClassPathRequest{
			AccountID: query.AccountID,
			ClassID:   classID,
		}, nil
	})
	return httpx.Endpoint(decodeFn, log, endpoint)
}

// GetAssetClassDetails returns class details with items, growth, and history.
//
// @Summary Get asset class details
// @Description Returns class details for drawer rendering.
// @Tags assets
// @Accept json
// @Produce json
// @Param class_id path string true "Class ID"
// @Param account_id query string true "Account ID"
// @Success 200 {object} api.AssetClassDetailsResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assets/classes/{class_id} [get]
func GetAssetClassDetails(log logging.Logger, svc assetService) http.Handler {
	endpoint := func(ctx context.Context, req assetClassPathRequest) (status int, res any, err error) {
		details, err := svc.GetClassDetails(ctx, assets.ListClassDetailsInput{
			AccountID: req.AccountID,
			ClassID:   req.ClassID,
		})
		if err != nil {
			if handledStatus, payload, handled := mapAssetServiceError(err); handled {
				return handledStatus, payload, nil
			}
			return http.StatusInternalServerError, struct{}{}, err
		}
		if details == nil {
			return http.StatusNotFound, map[string]string{"class_id": assets.ErrAssetClassNotFound.Error()}, nil
		}

		items := make([]api.AssetItemResponse, 0, len(details.Items))
		for _, item := range details.Items {
			items = append(items, api.AssetItemResponse{
				ID:           item.ID,
				Name:         item.Name,
				CurrentWorth: formatDecimal(item.CurrentWorth.Float64()),
				Archived:     item.Archived,
				UpdatedAt:    item.UpdatedAt,
			})
		}
		growth := make([]api.AssetGrowthPointResponse, 0, len(details.Growth))
		for _, point := range details.Growth {
			growth = append(growth, api.AssetGrowthPointResponse{
				Date:       point.Date.Format("2006-01-02"),
				TotalWorth: formatDecimal(point.TotalWorth.Float64()),
			})
		}
		history := make([]api.AssetHistoryResponse, 0, len(details.History))
		for _, row := range details.History {
			var direction *string
			if row.Direction != nil {
				value := string(*row.Direction)
				direction = &value
			}
			history = append(history, api.AssetHistoryResponse{
				ID:              row.ID,
				ItemID:          row.ItemID,
				ChangeType:      string(row.ChangeType),
				Direction:       direction,
				Amount:          formatDecimal(row.Amount.Float64()),
				PreviousWorth:   formatDecimal(row.PreviousWorth.Float64()),
				NewWorth:        formatDecimal(row.NewWorth.Float64()),
				ClassTotalWorth: formatDecimal(row.ClassTotalWorth.Float64()),
				EffectiveDate:   row.EffectiveDate.Format("2006-01-02"),
				Note:            row.Note,
				CreatedAt:       row.CreatedAt,
			})
		}

		return http.StatusOK, api.AssetClassDetailsResponse{
			Class:   toAssetClassResponse(details.Class),
			Items:   items,
			Growth:  growth,
			History: history,
		}, nil
	}
	decodeFn := httpx.DecoderFunc[assetClassPathRequest](func(r *http.Request) (assetClassPathRequest, error) {
		query, err := httpx.DecodeQuery[api.GetAssetClassesRequest](r)
		if err != nil {
			return assetClassPathRequest{}, err
		}
		classID, err := uuid.Parse(r.PathValue("class_id"))
		if err != nil {
			return assetClassPathRequest{}, fmt.Errorf("invalid class_id")
		}
		return assetClassPathRequest{
			AccountID: query.AccountID,
			ClassID:   classID,
		}, nil
	})
	return httpx.Endpoint(decodeFn, log, endpoint)
}

// CreateAssetItem creates a tracked item in a manual class.
//
// @Summary Create asset item
// @Description Adds a tracked asset item and records its initial worth.
// @Tags assets
// @Accept json
// @Produce json
// @Param request body api.CreateAssetItemRequest true "Create asset item payload"
// @Success 201 {object} api.AssetItemResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assets/items [post]
func CreateAssetItem(log logging.Logger, svc assetService) http.Handler {
	endpoint := func(ctx context.Context, req api.CreateAssetItemRequest) (status int, res any, err error) {
		item, err := svc.CreateItem(ctx, assets.CreateItemInput{
			AccountID:     req.AccountID,
			ClassID:       req.ClassID,
			Name:          req.Name,
			InitialWorth:  req.InitialWorth,
			EffectiveDate: req.EffectiveDate,
			Note:          req.Note,
		})
		if err != nil {
			if handledStatus, payload, handled := mapAssetServiceError(err); handled {
				return handledStatus, payload, nil
			}
			return http.StatusInternalServerError, struct{}{}, err
		}
		return http.StatusCreated, api.AssetItemResponse{
			ID:           item.ID,
			Name:         item.Name,
			CurrentWorth: formatDecimal(item.CurrentWorth.Float64()),
			Archived:     item.Archived,
			UpdatedAt:    item.UpdatedAt,
		}, nil
	}
	decodeFn := httpx.DecoderFunc[api.CreateAssetItemRequest](func(r *http.Request) (api.CreateAssetItemRequest, error) {
		return httpx.DecodeJSON[api.CreateAssetItemRequest](r)
	})
	return httpx.Endpoint(decodeFn, log, endpoint)
}

// SetAssetItemWorth sets absolute worth for an item.
//
// @Summary Set asset item worth
// @Description Replaces item worth using a non-future effective date.
// @Tags assets
// @Accept json
// @Produce json
// @Param request body api.SetAssetItemWorthRequest true "Set worth payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assets/items/worth/set [post]
func SetAssetItemWorth(log logging.Logger, svc assetService) http.Handler {
	endpoint := func(ctx context.Context, req api.SetAssetItemWorthRequest) (status int, res any, err error) {
		err = svc.SetItemWorth(ctx, assets.SetItemWorthInput{
			AccountID:     req.AccountID,
			ClassID:       req.ClassID,
			ItemID:        req.ItemID,
			Worth:         req.Worth,
			EffectiveDate: req.EffectiveDate,
			Note:          req.Note,
		})
		if err != nil {
			if handledStatus, payload, handled := mapAssetServiceError(err); handled {
				return handledStatus, payload, nil
			}
			return http.StatusInternalServerError, struct{}{}, err
		}
		return http.StatusOK, map[string]string{"status": "updated"}, nil
	}
	decodeFn := httpx.DecoderFunc[api.SetAssetItemWorthRequest](func(r *http.Request) (api.SetAssetItemWorthRequest, error) {
		return httpx.DecodeJSON[api.SetAssetItemWorthRequest](r)
	})
	return httpx.Endpoint(decodeFn, log, endpoint)
}

// AdjustAssetItemWorth adjusts item worth by directional amount.
//
// @Summary Adjust asset item worth
// @Description Applies increase/decrease delta using non-future effective date.
// @Tags assets
// @Accept json
// @Produce json
// @Param request body api.AdjustAssetItemWorthRequest true "Adjust worth payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assets/items/worth/adjust [post]
func AdjustAssetItemWorth(log logging.Logger, svc assetService) http.Handler {
	endpoint := func(ctx context.Context, req api.AdjustAssetItemWorthRequest) (status int, res any, err error) {
		err = svc.AdjustItemWorth(ctx, assets.AdjustItemWorthInput{
			AccountID:     req.AccountID,
			ClassID:       req.ClassID,
			ItemID:        req.ItemID,
			Direction:     req.Direction,
			Amount:        req.Amount,
			EffectiveDate: req.EffectiveDate,
			Note:          req.Note,
		})
		if err != nil {
			if handledStatus, payload, handled := mapAssetServiceError(err); handled {
				return handledStatus, payload, nil
			}
			return http.StatusInternalServerError, struct{}{}, err
		}
		return http.StatusOK, map[string]string{"status": "updated"}, nil
	}
	decodeFn := httpx.DecoderFunc[api.AdjustAssetItemWorthRequest](func(r *http.Request) (api.AdjustAssetItemWorthRequest, error) {
		return httpx.DecodeJSON[api.AdjustAssetItemWorthRequest](r)
	})
	return httpx.Endpoint(decodeFn, log, endpoint)
}

func parseAssetDateRange(fromRaw, toRaw string) (*time.Time, *time.Time, error) {
	var from *time.Time
	if fromRaw != "" {
		d, err := time.Parse("2006-01-02", fromRaw)
		if err != nil {
			return nil, nil, fmt.Errorf("from must be in YYYY-MM-DD format")
		}
		value := d.UTC()
		from = &value
	}

	var to *time.Time
	if toRaw != "" {
		d, err := time.Parse("2006-01-02", toRaw)
		if err != nil {
			return nil, nil, fmt.Errorf("to must be in YYYY-MM-DD format")
		}
		value := d.UTC().AddDate(0, 0, 1).Add(-time.Nanosecond)
		to = &value
	}

	if from != nil && to != nil && from.After(*to) {
		return nil, nil, fmt.Errorf("from must be before or equal to to")
	}
	return from, to, nil
}

func assetDateRangeProblem(err error) map[string]string {
	if err == nil {
		return map[string]string{"date_range": "invalid date range"}
	}

	switch err.Error() {
	case "from must be in YYYY-MM-DD format":
		return map[string]string{"from": "from must be in YYYY-MM-DD format"}
	case "to must be in YYYY-MM-DD format":
		return map[string]string{"to": "to must be in YYYY-MM-DD format"}
	case "from must be before or equal to to":
		return map[string]string{"from": "from must be before or equal to to"}
	default:
		return map[string]string{"date_range": err.Error()}
	}
}

func toAssetClassResponse(row assets.ClassSummary) api.AssetClassResponse {
	return api.AssetClassResponse{
		ID:           row.ID,
		Name:         row.Name,
		Source:       string(row.Source),
		Archived:     row.Archived,
		CurrentWorth: formatDecimal(row.CurrentWorth.Float64()),
		LastChangeAt: row.LastChangeAt,
		GrowthPct:    row.GrowthPct,
		UpdatedAt:    row.UpdatedAt,
	}
}

func mapAssetServiceError(err error) (int, map[string]string, bool) {
	switch {
	case errors.Is(err, assets.ErrAssetAccountNotFound):
		return http.StatusNotFound, map[string]string{"account_id": err.Error()}, true
	case errors.Is(err, assets.ErrAssetClassNotFound):
		return http.StatusNotFound, map[string]string{"class_id": err.Error()}, true
	case errors.Is(err, assets.ErrAssetItemNotFound):
		return http.StatusNotFound, map[string]string{"item_id": err.Error()}, true
	case errors.Is(err, assets.ErrAssetClassAlreadyExists):
		return http.StatusConflict, map[string]string{"class": err.Error()}, true
	case errors.Is(err, assets.ErrAssetItemAlreadyExists):
		return http.StatusConflict, map[string]string{"item": err.Error()}, true
	case errors.Is(err, assets.ErrAssetClassNameRequired),
		errors.Is(err, assets.ErrAssetClassNameReserved),
		errors.Is(err, assets.ErrAssetClassNotManual),
		errors.Is(err, assets.ErrAssetItemNameRequired),
		errors.Is(err, assets.ErrAssetWorthInvalid),
		errors.Is(err, assets.ErrAssetAmountInvalid),
		errors.Is(err, assets.ErrAssetDirectionInvalid),
		errors.Is(err, assets.ErrAssetEffectiveDateInvalid),
		errors.Is(err, assets.ErrAssetEffectiveDateFuture):
		return http.StatusBadRequest, map[string]string{"asset": err.Error()}, true
	default:
		return 0, nil, false
	}
}
