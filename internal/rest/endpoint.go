package rest

import (
	"context"
	"net/http"

	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
)

type EndpointFunc[T any, R any] func(ctx context.Context, req T) (status int, res R, err error)

type Validator interface {
	Valid(ctx context.Context) map[string]string
}

// endpoint creates a wrapper for endpoint logic.
// endpoint and returns a handler func. It decodes the request into a usable
// model which it passes into the fn HandlerFunc.
func Endpoint[T any, R any](decode DecoderFunc[T], log logging.Logger, fn EndpointFunc[T, R]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// decode the request body or query paramaters based on the decode
		// function passed into the handle func
		req, err := decode(r)
		if err != nil {
			// if the decoding of the request fails we return an error of a
			// internal server error to the client
			log.Error(r.Context(), "handle: a decode error occurred", err)
			_ = encode(w, http.StatusInternalServerError, map[string]string{"error": "internal server eror"})
			return
		}
		// if the request implements the Validator interface we execute the
		// valid function to add input validation to the request pipeline,
		// if we encounter any problems we return the problems to the client
		// as a bad request with the problems as a body
		if validator, ok := any(req).(Validator); ok {
			if problems := validator.Valid(r.Context()); len(problems) > 0 {
				_ = encode(w, http.StatusBadRequest, problems)
				return
			}
		}
		// here we call the handler function that satisfies the type defined
		// above. we pass the request context and the decoded context.
		status, res, err := fn(r.Context(), req)
		if err != nil {
			log.Error(r.Context(), "handle: an error occurred while handling a request", err)
			_ = encode(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}
		_ = encode(w, status, res)
	}
}
