package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/google/uuid"
)

type DecoderFunc[T any] func(r *http.Request) (T, error)

type MultipartFileDecoderOptions struct {
	FieldName    string
	MaxBytes     int64
	MaxMemory    int64    // for ParseMultipartForm's in-memory part
	AllowedTypes []string // optional: []{"text/csv", "application/vnd.ms-excel"}
}

func JSONDecoder[T any](r *http.Request) (T, error) {
	var req T
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func QueryDecoder[T any](r *http.Request) (T, error) {
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
				return target, fmt.Errorf("invalid int for %s: %w", tag, err)
			}
			f.SetInt(int64(i))
		case reflect.Float64:
			fv, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return target, fmt.Errorf("invalid float for %s: %w", tag, err)
			}
			f.SetFloat(fv)
		case reflect.Bool:
			bv, err := strconv.ParseBool(val)
			if err != nil {
				return target, fmt.Errorf("invalid bool for %s: %w", tag, err)
			}
			f.SetBool(bv)
		default:
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
		return out, fmt.Errorf("parsing multipart form: %w", err)
	}

	f, fh, err := r.FormFile(field)
	if err != nil {
		return out, fmt.Errorf("getting form file %q: %w", field, err)
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
			f.Close()
			return out, fmt.Errorf("invalid content type %q", ct)
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
				return out, fmt.Errorf("invalid int for %s: %w", formTag, err)
			}
			fieldValue.SetInt(intVal)
		case reflect.Float64:
			floatVal, err := strconv.ParseFloat(formVal, 64)
			if err != nil {
				return out, fmt.Errorf("invalid float for %s: %w", formTag, err)
			}
			fieldValue.SetFloat(floatVal)
		case reflect.Bool:
			boolVal, err := strconv.ParseBool(formVal)
			if err != nil {
				return out, fmt.Errorf("invalid bool for %s: %w", formTag, err)
			}
			fieldValue.SetBool(boolVal)
		default:
			// Handle UUID type
			if fieldValue.Type() == reflect.TypeOf(uuid.UUID{}) {
				parsedUUID, err := uuid.Parse(formVal)
				if err != nil {
					return out, fmt.Errorf("invalid UUID for %s: %w", formTag, err)
				}
				fieldValue.Set(reflect.ValueOf(parsedUUID))
			}
		}
	}

	return out, nil
}
