package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.elastic.co/apm/module/apmhttp/v2"
)

type Option func(*Client)

type Client struct {
	http    *http.Client
	baseURL string
}

type CallErrorKind string

const (
	CallErrorUnknown     CallErrorKind = "unknown"
	CallErrorUnreachable CallErrorKind = "unreachable"
	CallErrorServer      CallErrorKind = "server"
	CallErrorClient      CallErrorKind = "client"
)

type CallError struct {
	Kind       CallErrorKind
	StatusCode int
	Message    string
	Err        error
}

func (e *CallError) Error() string {
	if e == nil {
		return ""
	}
	switch {
	case e.StatusCode > 0 && e.Message != "":
		return fmt.Sprintf("agent call failed (%s) with status code %d: %s", e.Kind, e.StatusCode, e.Message)
	case e.StatusCode > 0:
		return fmt.Sprintf("agent call failed (%s) with status code %d", e.Kind, e.StatusCode)
	case e.Err != nil:
		return fmt.Sprintf("agent call failed (%s): %v", e.Kind, e.Err)
	default:
		return fmt.Sprintf("agent call failed (%s)", e.Kind)
	}
}

func (e *CallError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func IsUnreachableError(err error) bool {
	var callErr *CallError
	return errors.As(err, &callErr) && callErr.Kind == CallErrorUnreachable
}

func IsServerError(err error) bool {
	var callErr *CallError
	return errors.As(err, &callErr) && callErr.Kind == CallErrorServer
}

func IsClientError(err error) bool {
	var callErr *CallError
	return errors.As(err, &callErr) && callErr.Kind == CallErrorClient
}

func NewClient(baseURL string, opts ...Option) *Client {
	httpClient := &http.Client{
		Transport: apmhttp.WrapRoundTripper(http.DefaultTransport),
		Timeout:   5 * time.Minute,
	}

	c := &Client{
		http:    httpClient,
		baseURL: baseURL,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) CallAgent(ctx context.Context, ID uuid.UUID, msg string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/"+ID.String()+"/run", nil)
	if err != nil {
		return &CallError{Kind: CallErrorUnknown, Err: err}
	}
	q := req.URL.Query()
	q.Add("message", msg)
	req.URL.RawQuery = q.Encode()
	req.Header.Add("Accept", "application/json")
	res, err := c.http.Do(req)
	if err != nil {
		return &CallError{Kind: CallErrorUnreachable, Err: err}
	}
	bodyBytes, err := io.ReadAll(res.Body)
	closeErr := res.Body.Close()
	if err != nil {
		if closeErr != nil {
			return &CallError{Kind: CallErrorUnknown, Err: fmt.Errorf("read response body: %w (close failed: %v)", err, closeErr)}
		}
		return &CallError{Kind: CallErrorUnknown, Err: err}
	}
	if closeErr != nil {
		return &CallError{Kind: CallErrorUnknown, Err: fmt.Errorf("close response body: %w", closeErr)}
	}
	bodyString := strings.TrimSpace(string(bodyBytes))
	switch {
	case res.StatusCode >= http.StatusInternalServerError:
		return &CallError{
			Kind:       CallErrorServer,
			StatusCode: res.StatusCode,
			Message:    bodyString,
		}
	case res.StatusCode >= http.StatusBadRequest:
		return &CallError{
			Kind:       CallErrorClient,
			StatusCode: res.StatusCode,
			Message:    bodyString,
		}
	case res.StatusCode >= http.StatusMultipleChoices:
		return &CallError{
			Kind:       CallErrorUnknown,
			StatusCode: res.StatusCode,
			Message:    bodyString,
		}
	}
	return nil
}
