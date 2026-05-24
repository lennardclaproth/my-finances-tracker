// package http

// import (
// 	"context"
// 	"errors"
// 	"net/http"

// 	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
// 	"github.com/lennardclaproth/my-finances-tracker/internal/observability"
// )

// // EndpointFunc is the typed business function signature wrapped by Endpoint.
// type EndpointFunc[T any, R any] func(ctx context.Context, req T) (status int, res R, err error)

// // Validator is implemented by decoded request types that support semantic validation.
// type Validator interface {
// 	Valid(ctx context.Context) (map[string]string
// }

// // endpoint creates a wrapper for endpoint logic.
// // endpoint and returns a handler func. It decodes the request into a usable
// // model which it passes into the fn HandlerFunc.
// func Endpoint[T any, R any](decode DecoderFunc[T], log logging.Logger, fn EndpointFunc[T, R]) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		operationRoute := r.Pattern
// 		if operationRoute == "" {
// 			operationRoute = r.URL.Path
// 		}
// 		operation := observability.HTTPOperation(r.Method, operationRoute)

// 		// decode the request body or query paramaters based on the decode
// 		// function passed into the handle func
// 		req, err := decode(r)
// 		if err != nil {
// 			status, payload := decodeErrorResponse(err)
// 			log.Error(r.Context(), "handle: a decode error occurred", err,
// 				"operation", operation,
// 				"component", "http",
// 				"outcome", "failure",
// 				"error_class", "decode",
// 				"status", status,
// 			)
// 			if encErr := encode(w, status, payload); encErr != nil {
// 				log.Error(r.Context(), "handle: failed to encode decode error response", encErr,
// 					"operation", operation,
// 					"component", "http",
// 					"outcome", "failure",
// 					"error_class", "encode",
// 					"status", status,
// 				)
// 			}
// 			return
// 		}
// 		// if the request implements the Validator interface we execute the
// 		// valid function to add input validation to the request pipeline,
// 		// if we encounter any problems we return the problems to the client
// 		// as a bad request with the problems as a body
// 		if validator, ok := any(req).(Validator); ok {
// 			if problems := validator.Valid(r.Context()); len(problems) > 0 {
// 				log.Info(r.Context(), "handle: request validation failed",
// 					"operation", operation,
// 					"component", "http",
// 					"outcome", "failure",
// 					"error_class", "validation",
// 					"status", http.StatusBadRequest,
// 				)
// 				if encErr := encode(w, http.StatusBadRequest, problems); encErr != nil {
// 					log.Error(r.Context(), "handle: failed to encode validation response", encErr,
// 						"operation", operation,
// 						"component", "http",
// 						"outcome", "failure",
// 						"error_class", "encode",
// 						"status", http.StatusBadRequest,
// 					)
// 				}
// 				return
// 			}
// 		}
// 		// here we call the handler function that satisfies the type defined
// 		// above. we pass the request context and the decoded context.
// 		status, res, err := fn(r.Context(), req)
// 		if err != nil {
// 			log.Error(r.Context(), "handle: an error occurred while handling a request", err,
// 				"operation", operation,
// 				"component", "http",
// 				"outcome", "failure",
// 				"error_class", "internal",
// 				"status", http.StatusInternalServerError,
// 			)
// 			if encErr := encode(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"}); encErr != nil {
// 				log.Error(r.Context(), "handle: failed to encode internal error response", encErr,
// 					"operation", operation,
// 					"component", "http",
// 					"outcome", "failure",
// 					"error_class", "encode",
// 					"status", http.StatusInternalServerError,
// 				)
// 			}
// 			return
// 		}

// 		if encErr := encode(w, status, res); encErr != nil {
// 			log.Error(r.Context(), "handle: failed to encode success response", encErr,
// 				"operation", operation,
// 				"component", "http",
// 				"outcome", "failure",
// 				"error_class", "encode",
// 				"status", status,
// 			)
// 		}
// 	}
// }

// func decodeErrorResponse(err error) (int, map[string]string) {
// 	status := http.StatusBadRequest
// 	payload := map[string]string{"error": "invalid request"}

// 	var decErr *DecodeError
// 	if !errors.As(err, &decErr) {
// 		return status, payload
// 	}

// 	if decErr.Status != 0 {
// 		status = decErr.Status
// 	}
// 	if decErr.Field != "" && decErr.Message != "" {
// 		payload = map[string]string{decErr.Field: decErr.Message}
// 		return status, payload
// 	}
// 	if decErr.Message != "" {
// 		payload = map[string]string{"error": decErr.Message}
// 	}
// 	return status, payload
// }
