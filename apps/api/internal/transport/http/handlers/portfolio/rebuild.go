package portfolio

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	portfoliodomain "github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
	httpx "github.com/lennardclaproth/my-finances-tracker/internal/transport/http"
)

// RebuildPortfolioRequest requests a portfolio rebuild for an account.
type RebuildPortfolioRequest struct {
	AccountID uuid.UUID `json:"account_id"`
}

func (r RebuildPortfolioRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)
	if r.AccountID == uuid.Nil {
		problems["account_id"] = "account_id is required"
	}
	return len(problems) == 0, problems
}

type portfolioRebuilder interface {
	Build(ctx context.Context, accountID uuid.UUID) error
}

// RebuildPortfolio rebuilds the portfolio for a specific account.
//
// @Summary Rebuild portfolio
// @Description Rebuilds portfolio positions and snapshots for the selected account.
// @Tags portfolio
// @Accept json
// @Produce json
// @Param request body RebuildPortfolioRequest true "Rebuild request payload"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /portfolio/rebuild [post]
func RebuildPortfolio(
	log logging.Logger,
	rebuilder portfolioRebuilder,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.JSONDecode[RebuildPortfolioRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
			return
		}

		isValid, problems := req.isValid()
		if !isValid {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, problems)
			return
		}

		if rebuilder == nil {
			log.Error(r.Context(), "portfolio rebuild: rebuilder is not configured", errors.New("portfolio rebuilder not configured"))
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to rebuild portfolio"})
			return
		}

		if err := rebuilder.Build(r.Context(), req.AccountID); err != nil {
			switch {
			case errors.Is(err, portfoliodomain.ErrAccountNotFound):
				_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"account_id": portfoliodomain.ErrAccountNotFound.Error()})
			case errors.Is(err, portfoliodomain.ErrBuildInProgress):
				_ = httpx.JSONEncode(w, http.StatusConflict, map[string]string{"rebuild": "portfolio rebuild already in progress"})
			case errors.Is(err, portfoliodomain.ErrPortfolioNoSnapshots):
				_ = httpx.JSONEncode(w, http.StatusUnprocessableEntity, map[string]string{"rebuild": portfoliodomain.ErrPortfolioNoSnapshots.Error()})
			default:
				log.Error(r.Context(), "portfolio rebuild: failed to rebuild portfolio", err, "account_id", req.AccountID.String())
				_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to rebuild portfolio"})
			}
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}
