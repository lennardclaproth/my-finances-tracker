package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// # DecoderFunc
//
// DecoderFunc is a function type that defines how to decode an http request into a specific type.
// It returns the decoded value and an error if the decoding fails. The error can be of type DecodeError to provide more
// context about the decoding failure.
type DecoderFunc[T any] func(r *http.Request) (T, error)

// # DecodeError
//
// DecodeError is a custom error type that provides additional context about decoding failures. It includes the field that caused the error,
// a message describing the error, and an optional HTTP status code to indicate the appropriate response status for the error.
type DecodeError struct {
	// Field is the name of the field that caused the decoding error, if applicable.
	Field string
	// Message provides a description of the decoding error.
	Message string
	// Status is an HTTP status code that can be used to indicate the appropriate response status for this error.
	Status int
	// Err is the underlying error that caused the decoding failure, if applicable.
	Err error
}

// Error implements the error interface for DecodeError. It returns a string representation of the error, including the field and message if available.
func (e *DecodeError) Error() string {
	if e == nil {
		return "decode error"
	}

	base := "decode error"

	// Construct the error message based on the available information.
	// 1. If both Field and Message are present, include both in the error message.
	// 2. If only Message is present, use it as the error message.
	// 3. If Err is present, include the underlying error message in the output.
	if e.Field != "" && e.Message != "" {
		base = fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	if e.Field == "" && e.Message != "" {
		base = e.Message
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", base, e.Err)
	}

	return base
}

// Unwrap returns the underlying decode failure.
func (e *DecodeError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// MultipartFileDecoderOptions configures multipart request decoding behavior.
type MultipartFileDecoderOptions struct {
	FieldName    string
	MaxBytes     int64
	MaxMemory    int64    // for ParseMultipartForm's in-memory part
	AllowedTypes []string // optional: []{"text/csv", "application/vnd.ms-excel"}
}

// JSONDecode decodes a strict single JSON object payload into T.
func JSONDecode[T any](r *http.Request) (T, error) {
	var req T
	// Check for empty body before decoding to provide a clearer error message.
	if r.Body == nil {
		return req, &DecodeError{
			Status:  http.StatusBadRequest,
			Field:   "body",
			Message: "request body is required",
			Err:     io.EOF,
		}
	}
	// Use a decoder with DisallowUnknownFields to catch extra fields and provide better error messages.
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&req)
	// Classify JSON decoding errors to provide more specific error messages.
	if err != nil {
		return req, classifyJSONDecodeError(err)
	}
	// Check for multiple JSON values in the body, which is not allowed.
	if err := decoder.Decode(&struct{}{}); err != nil && !errors.Is(err, io.EOF) {
		return req, &DecodeError{
			Status:  http.StatusBadRequest,
			Field:   "body",
			Message: "request body must contain a single JSON value",
			Err:     err,
		}
	}
	return req, nil
}

// DecodeQuery decodes query-string values into fields tagged with `query`.
func DecodeQuery[T any](r *http.Request) (T, error) {
	var target T
	values := r.URL.Query()

	v := reflect.ValueOf(&target).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("query")
		if tag == "" {
			continue
		}

		val := values.Get(tag)
		if val == "" {
			continue // optional param
		}

		f := v.Field(i)
		if !f.CanSet() {
			continue
		}

		switch f.Kind() {
		case reflect.String:
			f.SetString(val)
		case reflect.Int, reflect.Int64:
			i, err := strconv.Atoi(val)
			if err != nil {
				return target, &DecodeError{
					Status:  http.StatusBadRequest,
					Field:   tag,
					Message: fmt.Sprintf("%s must be a valid integer", tag),
					Err:     err,
				}
			}
			f.SetInt(int64(i))
		case reflect.Float64:
			fv, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return target, &DecodeError{
					Status:  http.StatusBadRequest,
					Field:   tag,
					Message: fmt.Sprintf("%s must be a valid float", tag),
					Err:     err,
				}
			}
			f.SetFloat(fv)
		case reflect.Bool:
			bv, err := strconv.ParseBool(val)
			if err != nil {
				return target, &DecodeError{
					Status:  http.StatusBadRequest,
					Field:   tag,
					Message: fmt.Sprintf("%s must be a valid boolean", tag),
					Err:     err,
				}
			}
			f.SetBool(bv)
		default:
			// Handle UUID type.
			if f.Type() == reflect.TypeOf(uuid.UUID{}) {
				parsedUUID, err := uuid.Parse(val)
				if err != nil {
					return target, &DecodeError{
						Status:  http.StatusBadRequest,
						Field:   tag,
						Message: fmt.Sprintf("%s must be a valid UUID", tag),
						Err:     err,
					}
				}
				f.Set(reflect.ValueOf(parsedUUID))
				continue
			}
			// Handle *UUID type.
			if f.Type() == reflect.TypeOf((*uuid.UUID)(nil)) {
				parsedUUID, err := uuid.Parse(val)
				if err != nil {
					return target, &DecodeError{
						Status:  http.StatusBadRequest,
						Field:   tag,
						Message: fmt.Sprintf("%s must be a valid UUID", tag),
						Err:     err,
					}
				}
				f.Set(reflect.ValueOf(&parsedUUID))
				continue
			}
			// silently ignore unsupported types
		}
	}
	return target, nil
}

func encode[T any](w http.ResponseWriter, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

// DecodeMultipartFile is a generic decoder that extracts a multipart file
// and populates a struct with both file metadata and extra form fields.
//
// Struct tags:
//
// - `multipart:"file"` for the file field (type: multipart.File)
// - `multipart:"filename"` for the filename (type: string)
// - `multipart:"size"` for file size (type: int64)
// - `multipart:"header"` for MIME headers (type: textproto.MIMEHeader)
// - `form:"field_name"` for extra form fields (supports: string, int, int64, float64, bool, uuid.UUID)
// DecodeMultipartFile decodes a multipart request with one file and optional form fields.
func DecodeMultipartFile[T any](r *http.Request, opt MultipartFileDecoderOptions) (T, error) {
	var out T

	field := opt.FieldName
	if field == "" {
		field = "file"
	}
	maxBytes := opt.MaxBytes
	if maxBytes == 0 {
		maxBytes = 10 << 20 // 10MB
	}
	maxMem := opt.MaxMemory
	if maxMem == 0 {
		maxMem = maxBytes
	}

	// Hard cap request body size (important for DoS protection).
	r.Body = http.MaxBytesReader(nil, r.Body, maxBytes)

	if err := r.ParseMultipartForm(maxMem); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) || strings.Contains(strings.ToLower(err.Error()), "request body too large") {
			return out, &DecodeError{
				Status:  http.StatusRequestEntityTooLarge,
				Field:   field,
				Message: fmt.Sprintf("%s exceeds max size of %d bytes", field, maxBytes),
				Err:     err,
			}
		}
		return out, &DecodeError{
			Status:  http.StatusBadRequest,
			Field:   "body",
			Message: "invalid multipart form payload",
			Err:     err,
		}
	}

	f, fh, err := r.FormFile(field)
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return out, &DecodeError{
				Status:  http.StatusBadRequest,
				Field:   field,
				Message: fmt.Sprintf("%s is required", field),
				Err:     err,
			}
		}
		return out, &DecodeError{
			Status:  http.StatusBadRequest,
			Field:   field,
			Message: fmt.Sprintf("failed to read %s", field),
			Err:     err,
		}
	}

	// Optional content-type check (best-effort; can be missing/lying).
	if len(opt.AllowedTypes) > 0 {
		ct := fh.Header.Get("Content-Type")
		allowed := false
		for _, a := range opt.AllowedTypes {
			if ct == a {
				allowed = true
				break
			}
		}
		if !allowed {
			if closeErr := f.Close(); closeErr != nil {
				return out, &DecodeError{
					Status:  http.StatusUnsupportedMediaType,
					Field:   field,
					Message: "unsupported media type",
					Err:     closeErr,
				}
			}
			return out, &DecodeError{
				Status:  http.StatusUnsupportedMediaType,
				Field:   field,
				Message: "unsupported media type",
				Err:     nil,
			}
		}
	}

	// Use reflection to populate struct fields
	v := reflect.ValueOf(&out).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		structField := t.Field(i)
		fieldValue := v.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		// Handle multipart-specific fields (File, Filename, Size, Header)
		if multipartTag := structField.Tag.Get("multipart"); multipartTag != "" {
			switch multipartTag {
			case "file":
				// multipart.File is an interface type
				if fieldValue.Type().String() == "multipart.File" {
					fieldValue.Set(reflect.ValueOf(f))
				}
			case "filename":
				if fieldValue.Kind() == reflect.String {
					fieldValue.SetString(fh.Filename)
				}
			case "size":
				if fieldValue.Kind() == reflect.Int64 {
					fieldValue.SetInt(fh.Size)
				}
			case "header":
				if fieldValue.Type().Name() == "MIMEHeader" {
					fieldValue.Set(reflect.ValueOf(fh.Header))
				}
			}
			continue
		}

		// Handle form fields
		formTag := structField.Tag.Get("form")
		if formTag == "" {
			continue
		}

		formVal := r.FormValue(formTag)
		if formVal == "" {
			continue // optional param
		}

		// Parse form value based on field type
		switch fieldValue.Kind() {
		case reflect.String:
			fieldValue.SetString(formVal)
		case reflect.Int, reflect.Int64:
			intVal, err := strconv.ParseInt(formVal, 10, 64)
			if err != nil {
				return out, &DecodeError{
					Status:  http.StatusBadRequest,
					Field:   formTag,
					Message: fmt.Sprintf("%s must be a valid integer", formTag),
					Err:     err,
				}
			}
			fieldValue.SetInt(intVal)
		case reflect.Float64:
			floatVal, err := strconv.ParseFloat(formVal, 64)
			if err != nil {
				return out, &DecodeError{
					Status:  http.StatusBadRequest,
					Field:   formTag,
					Message: fmt.Sprintf("%s must be a valid float", formTag),
					Err:     err,
				}		
			}
			fieldValue.SetFloat(floatVal)
		case reflect.Bool:
			boolVal, err := strconv.ParseBool(formVal)
			if err != nil {
				return out, &DecodeError{
					Status:  http.StatusBadRequest,
					Field:   formTag,
					Message: fmt.Sprintf("%s must be a valid boolean", formTag),
					Err:     err,
				}
			}
			fieldValue.SetBool(boolVal)
		default:
			// Handle UUID type
			if fieldValue.Type() == reflect.TypeOf(uuid.UUID{}) {
				parsedUUID, err := uuid.Parse(formVal)
				if err != nil {
					return out, &DecodeError{
						Status:  http.StatusBadRequest,
						Field:   formTag,
						Message: fmt.Sprintf("%s must be a valid UUID", formTag),
						Err:     err,
					}
				}
				fieldValue.Set(reflect.ValueOf(parsedUUID))
				continue
			}
			// Handle *UUID type
			if fieldValue.Type() == reflect.TypeOf((*uuid.UUID)(nil)) {
				parsedUUID, err := uuid.Parse(formVal)
				if err != nil {
					return out, &DecodeError{
						Status:  http.StatusBadRequest,
						Field:   formTag,
						Message: fmt.Sprintf("%s must be a valid UUID", formTag),
						Err:     err,
					}
				}
				fieldValue.Set(reflect.ValueOf(&parsedUUID))
			}
		}
	}

	return out, nil
}

func classifyJSONDecodeError(err error) error {
	if errors.Is(err, io.EOF) {
		return &DecodeError{
			Status:  http.StatusBadRequest,
			Field:   "body",
			Message: "request body is required",
			Err:     err,
		}
	}

	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return &DecodeError{
			Status:  http.StatusBadRequest,
			Field:   "body",
			Message: "malformed JSON body",
			Err:     err,
		}
	}

	if errors.Is(err, io.ErrUnexpectedEOF) {
		return &DecodeError{
			Status:  http.StatusBadRequest,
			Field:   "body",
			Message: "malformed JSON body",
			Err:     err,
		}
	}

	msg := err.Error()
	if strings.HasPrefix(msg, "json: unknown field ") {
		field := strings.TrimPrefix(msg, "json: unknown field ")
		field = strings.Trim(field, "\"")
		return &DecodeError{
			Status:  http.StatusBadRequest,
			Field:   field,
			Message: fmt.Sprintf("unknown field %q", field),
			Err:     err,
		}
	}

	if strings.HasPrefix(msg, "json: cannot unmarshal ") {
		return &DecodeError{
			Status:  http.StatusBadRequest,
			Field:   "body",
			Message: "JSON body has invalid types",
			Err:     err,
		}
	}

	return &DecodeError{
		Status:  http.StatusBadRequest,
		Field:   "body",
		Message: "invalid JSON body",
		Err:     err,
	}
}
