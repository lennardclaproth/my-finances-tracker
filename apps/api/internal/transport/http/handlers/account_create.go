package handlers

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
func CreateAccount(log logging.Logger, cah handlers.CreateAccountHandler) http.Handler {
	return func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		req, err := apphttp.JSONDecode[api.CreateAccountRequest](r)
	}
}
// func CreateAccount(log logging.Logger, creator account.Creator) http.Handler {
// 	endpoint := func(ctx context.Context, req api.CreateAccountRequest) (status int, res any, err error) {
// 		acc, err := account.NewAccount(req.Name, req.ExternalID)
// 		if err != nil {
// 			if errors.Is(err, account.ErrAccountNameRequired) {
// 				return http.StatusBadRequest, map[string]string{"name": account.ErrAccountNameRequired.Error()}, nil
// 			}
// 			return http.StatusBadRequest, map[string]string{"account": err.Error()}, nil
// 		}
// 		if err := creator.Create(ctx, acc); err != nil {
// 			if errors.Is(err, account.ErrAccountAlreadyExists) {
// 				return http.StatusConflict, map[string]string{"account": account.ErrAccountAlreadyExists.Error()}, nil
// 			}
// 			return http.StatusInternalServerError, struct{}{}, err
// 		}
// 		return http.StatusOK, api.AccountResponse{
// 			ID:         acc.ID,
// 			ExternalID: acc.ExternalID,
// 			Name:       acc.Name,
// 			CreatedAt:  acc.CreatedAt,
// 			UpdatedAt:  acc.UpdatedAt,
// 		}, nil
// 	}

// 	decodeFn := httpx.DecoderFunc[api.CreateAccountRequest](func(r *http.Request) (api.CreateAccountRequest, error) {
// 		return httpx.DecodeJSON[api.CreateAccountRequest](r)
// 	})

// 	return httpx.Endpoint(decodeFn, log, endpoint)
// }
