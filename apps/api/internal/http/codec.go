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

// DecoderFunc decodes an incoming HTTP request into a typed request model.
type DecoderFunc[T any] func(r *http.Request) (T, error)

// DecodeError captures structured decode failure details for endpoint responses.
type DecodeError struct {
	Status  int
	Field   string
	Message string
	Err     error
}

// Error returns the wrapped decode error message.
func (e *DecodeError) Error() string {
	if e == nil {
		return "decode error"
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	if e.Message != "" {
		return e.Message
	}
	return "decode error"
}

// Unwrap returns the underlying decode failure.
func (e *DecodeError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func newDecodeError(status int, field, message string, err error) error {
	return &DecodeError{
		Status:  status,
		Field:   field,
		Message: message,
		Err:     err,
	}
}

// MultipartFileDecoderOptions configures multipart request decoding behavior.
type MultipartFileDecoderOptions struct {
	FieldName    string
	MaxBytes     int64
	MaxMemory    int64    // for ParseMultipartForm's in-memory part
	AllowedTypes []string // optional: []{"text/csv", "application/vnd.ms-excel"}
}

// DecodeJSON decodes a strict single JSON object payload into T.
func DecodeJSON[T any](r *http.Request) (T, error) {
	var req T
	if r.Body == nil {
		return req, newDecodeError(http.StatusBadRequest, "body", "request body is required", io.EOF)
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		return req, classifyJSONDecodeError(err)
	}

	if err := decoder.Decode(&struct{}{}); err != nil && !errors.Is(err, io.EOF) {
		return req, newDecodeError(
			http.StatusBadRequest,
			"body",
			"request body must contain a single JSON value",
			err,
		)
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
				return target, newDecodeError(
					http.StatusBadRequest,
					tag,
					fmt.Sprintf("%s must be a valid integer", tag),
					err,
				)
			}
			f.SetInt(int64(i))
		case reflect.Float64:
			fv, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return target, newDecodeError(
					http.StatusBadRequest,
					tag,
					fmt.Sprintf("%s must be a valid float", tag),
					err,
				)
			}
			f.SetFloat(fv)
		case reflect.Bool:
			bv, err := strconv.ParseBool(val)
			if err != nil {
				return target, newDecodeError(
					http.StatusBadRequest,
					tag,
					fmt.Sprintf("%s must be a valid boolean", tag),
					err,
				)
			}
			f.SetBool(bv)
		default:
			// Handle UUID type.
			if f.Type() == reflect.TypeOf(uuid.UUID{}) {
				parsedUUID, err := uuid.Parse(val)
				if err != nil {
					return target, newDecodeError(
						http.StatusBadRequest,
						tag,
						fmt.Sprintf("%s must be a valid UUID", tag),
						err,
					)
				}
				f.Set(reflect.ValueOf(parsedUUID))
				continue
			}
			// Handle *UUID type.
			if f.Type() == reflect.TypeOf((*uuid.UUID)(nil)) {
				parsedUUID, err := uuid.Parse(val)
				if err != nil {
					return target, newDecodeError(
						http.StatusBadRequest,
						tag,
						fmt.Sprintf("%s must be a valid UUID", tag),
						err,
					)
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
			return out, newDecodeError(
				http.StatusRequestEntityTooLarge,
				field,
				fmt.Sprintf("%s exceeds max size of %d bytes", field, maxBytes),
				err,
			)
		}
		return out, newDecodeError(http.StatusBadRequest, "body", "invalid multipart form payload", err)
	}

	f, fh, err := r.FormFile(field)
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return out, newDecodeError(http.StatusBadRequest, field, fmt.Sprintf("%s is required", field), err)
		}
		return out, newDecodeError(http.StatusBadRequest, field, fmt.Sprintf("failed to read %s", field), err)
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
				return out, newDecodeError(http.StatusUnsupportedMediaType, field, "unsupported media type", closeErr)
			}
			return out, newDecodeError(http.StatusUnsupportedMediaType, field, "unsupported media type", nil)
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
				return out, newDecodeError(
					http.StatusBadRequest,
					formTag,
					fmt.Sprintf("%s must be a valid integer", formTag),
					err,
				)
			}
			fieldValue.SetInt(intVal)
		case reflect.Float64:
			floatVal, err := strconv.ParseFloat(formVal, 64)
			if err != nil {
				return out, newDecodeError(
					http.StatusBadRequest,
					formTag,
					fmt.Sprintf("%s must be a valid float", formTag),
					err,
				)
			}
			fieldValue.SetFloat(floatVal)
		case reflect.Bool:
			boolVal, err := strconv.ParseBool(formVal)
			if err != nil {
				return out, newDecodeError(
					http.StatusBadRequest,
					formTag,
					fmt.Sprintf("%s must be a valid boolean", formTag),
					err,
				)
			}
			fieldValue.SetBool(boolVal)
		default:
			// Handle UUID type
			if fieldValue.Type() == reflect.TypeOf(uuid.UUID{}) {
				parsedUUID, err := uuid.Parse(formVal)
				if err != nil {
					return out, newDecodeError(
						http.StatusBadRequest,
						formTag,
						fmt.Sprintf("%s must be a valid UUID", formTag),
						err,
					)
				}
				fieldValue.Set(reflect.ValueOf(parsedUUID))
				continue
			}
			// Handle *UUID type
			if fieldValue.Type() == reflect.TypeOf((*uuid.UUID)(nil)) {
				parsedUUID, err := uuid.Parse(formVal)
				if err != nil {
					return out, newDecodeError(
						http.StatusBadRequest,
						formTag,
						fmt.Sprintf("%s must be a valid UUID", formTag),
						err,
					)
				}
				fieldValue.Set(reflect.ValueOf(&parsedUUID))
			}
		}
	}

	return out, nil
}

func classifyJSONDecodeError(err error) error {
	if errors.Is(err, io.EOF) {
		return newDecodeError(http.StatusBadRequest, "body", "request body is required", err)
	}

	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return newDecodeError(http.StatusBadRequest, "body", "malformed JSON body", err)
	}

	if errors.Is(err, io.ErrUnexpectedEOF) {
		return newDecodeError(http.StatusBadRequest, "body", "malformed JSON body", err)
	}

	msg := err.Error()
	if strings.HasPrefix(msg, "json: unknown field ") {
		field := strings.TrimPrefix(msg, "json: unknown field ")
		field = strings.Trim(field, "\"")
		return newDecodeError(http.StatusBadRequest, field, fmt.Sprintf("unknown field %q", field), err)
	}

	if strings.HasPrefix(msg, "json: cannot unmarshal ") {
		return newDecodeError(http.StatusBadRequest, "body", "JSON body has invalid types", err)
	}

	return newDecodeError(http.StatusBadRequest, "body", "invalid JSON body", err)
}
