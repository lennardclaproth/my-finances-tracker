package agent

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.elastic.co/apm/module/apmhttp/v2"
)

type Option func(*Client)

type Client struct {
	http    *http.Client
	baseURL string
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
		return err
	}
	q := req.URL.Query()
	q.Add("message", msg)
	req.URL.RawQuery = q.Encode()
	req.Header.Add("Accept", "application/json")
	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	bodyBytes, err := io.ReadAll(res.Body)
	bodyString := string(bodyBytes)
	if err != nil {
		return err
	}
	if res.StatusCode >= 300 {
		return fmt.Errorf("Agent call failed with status code %d, message: %s", res.StatusCode, bodyString)
	}
	return nil
}
