package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/lennardclaproth/my-finances-tracker/api"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	httpx "github.com/lennardclaproth/my-finances-tracker/internal/http"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
)

// CreateAccount creates an account used to map imports and portfolio builds.
//
// @Summary Create account
// @Description Creates a new account.
// @Tags accounts
// @Accept json
// @Produce json
// @Param request body api.CreateAccountRequest true "Create account payload"
// @Success 200 {object} api.AccountResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /accounts [post]
func CreateAccount(log logging.Logger, creator account.Creator) http.Handler {
	endpoint := func(ctx context.Context, req api.CreateAccountRequest) (status int, res any, err error) {
		acc, err := account.NewAccount(req.Name, req.ExternalID)
		if err != nil {
			if errors.Is(err, account.ErrAccountNameRequired) {
				return http.StatusBadRequest, map[string]string{"name": account.ErrAccountNameRequired.Error()}, nil
			}
			return http.StatusBadRequest, map[string]string{"account": err.Error()}, nil
		}
		if err := creator.Create(ctx, acc); err != nil {
			if errors.Is(err, account.ErrAccountAlreadyExists) {
				return http.StatusConflict, map[string]string{"account": account.ErrAccountAlreadyExists.Error()}, nil
			}
			return http.StatusInternalServerError, struct{}{}, err
		}
		return http.StatusOK, api.AccountResponse{
			ID:         acc.ID,
			ExternalID: acc.ExternalID,
			Name:       acc.Name,
			CreatedAt:  acc.CreatedAt,
			UpdatedAt:  acc.UpdatedAt,
		}, nil
	}

	decodeFn := httpx.DecoderFunc[api.CreateAccountRequest](func(r *http.Request) (api.CreateAccountRequest, error) {
		return httpx.DecodeJSON[api.CreateAccountRequest](r)
	})

	return httpx.Endpoint(decodeFn, log, endpoint)
}

// GetAccounts lists all available accounts.
//
// @Summary List accounts
// @Description Lists all available accounts.
// @Tags accounts
// @Accept json
// @Produce json
// @Success 200 {array} api.AccountResponse
// @Failure 500 {object} map[string]string
// @Router /accounts [get]
func GetAccounts(log logging.Logger, lister account.Lister) http.Handler {
	endpoint := func(ctx context.Context, _ struct{}) (status int, res []api.AccountResponse, err error) {
		accounts, err := lister.List(ctx)
		if err != nil {
			return http.StatusInternalServerError, nil, err
		}
		out := make([]api.AccountResponse, 0, len(accounts))
		for _, acc := range accounts {
			if acc == nil {
				continue
			}
			out = append(out, api.AccountResponse{
				ID:         acc.ID,
				ExternalID: acc.ExternalID,
				Name:       acc.Name,
				CreatedAt:  acc.CreatedAt,
				UpdatedAt:  acc.UpdatedAt,
			})
		}
		return http.StatusOK, out, nil
	}

	decodeFn := httpx.DecoderFunc[struct{}](func(r *http.Request) (struct{}, error) {
		return struct{}{}, nil
	})

	return httpx.Endpoint(decodeFn, log, endpoint)
}
