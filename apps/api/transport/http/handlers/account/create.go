package account

import (
	"errors"
	"net/http"
	"strings"

	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	httpx "github.com/lennardclaproth/my-finances-tracker/transport/http"
)

// CreateAccountRequest creates an account record.
type CreateAccountRequest struct {
	Name       string  `json:"name"`
	ExternalID *string `json:"external_id,omitempty"`
}

type CreateAccountResponse struct {
	ID string `json:"id"`
}

// Valid validates required account creation fields.
func (r CreateAccountRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)
	if strings.TrimSpace(r.Name) == "" {
		problems["name"] = "name is required"
	}

	if name := strings.TrimSpace(r.Name); len(name) > 255 {
		problems["name"] = "name cannot exceed 255 characters"
	}

	return len(problems) == 0, problems
}

// CreateAccount creates an account for a user.
//
// @Summary Create account
// @Description Creates a new account.
// @Tags accounts
// @Accept json
// @Produce json
// @Param request body CreateAccountRequest true "Create account payload"
// @Success 201 {object} CreateAccountResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /accounts [post]
func Create(log logging.Logger, commands account.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.JSONDecode[CreateAccountRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err){
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

		id, err := commands.Create(r.Context(), nil, req.ExternalID, req.Name)
		if err != nil {
			if errors.Is(err, account.ErrAccountAlreadyExists) {
				httpx.JSONEncode(w, http.StatusConflict, map[string]string{"error": "account with the same external_id already exists"})
			}
			httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to create account"})
			log.Error(r.Context(), "create account: failed to create account", err)
			return
		}

		httpx.JSONEncode(w, http.StatusCreated, CreateAccountResponse{ID: id.String()})
	})
}
