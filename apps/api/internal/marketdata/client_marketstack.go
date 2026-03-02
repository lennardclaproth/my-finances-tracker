package marketdata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.elastic.co/apm/module/apmhttp/v2"
)

type marketstackTime struct {
	time.Time
}

func (m *marketstackTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		m.Time = time.Time{}
		return nil
	}
	// Marketstack: 2026-02-06T00:00:00+0000  (no colon in offset)
	t, err := time.Parse("2006-01-02T15:04:05-0700", s)
	if err != nil {
		return err
	}
	m.Time = t
	return nil
}

type marketstackInt64 int64

func (m *marketstackInt64) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	if s == "" || s == "null" {
		*m = 0
		return nil
	}
	// Accept plain numbers (e.g. 11580, 11580.0) and quoted numbers.
	s = strings.Trim(s, `"`)
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	*m = marketstackInt64(int64(math.Round(f)))
	return nil
}

type ProviderStore interface {
	GetByName(ctx context.Context, name ProviderName) (*Provider, error)
	DeductTokens(ctx context.Context, name ProviderName, count int32) error
}

type MarketStackClient struct {
	http         *http.Client
	ps           ProviderStore
	providerName ProviderName
}

func NewMarketStackClient(ps ProviderStore, providerName ProviderName) *MarketStackClient {
	hc := &http.Client{
		Transport: apmhttp.WrapRoundTripper(http.DefaultTransport),
		Timeout:   20 * time.Second,
	}
	return &MarketStackClient{
		http:         hc,
		ps:           ps,
		providerName: providerName,
	}
}

type marketstackEODResponse struct {
	Pagination struct {
		Limit  int `json:"limit"`
		Total  int `json:"total"`
		Offset int `json:"offset"`
		Count  int `json:"count"`
	} `json:"pagination"`
	Data []struct {
		Symbol string           `json:"symbol"`
		Date   marketstackTime  `json:"date"`
		Open   float64          `json:"open"`
		Close  float64          `json:"close"`
		High   float64          `json:"high"`
		Low    float64          `json:"low"`
		Volume marketstackInt64 `json:"volume"`
	} `json:"data"`
}

// GetEOD builds a http request following the format:
//
// 'https://api.marketstack.com/v2/eod?symbols=tdt.as,aapl&limit=1000&access_key=XXXX'
//
// it constructs history objects from the response and returns them as an iterator.
// It handles pagination by making multiple requests until all data is retrieved.
// If any request fails, it yields an error and stops the iteration.
func (c *MarketStackClient) GetEOD(ctx context.Context, symbols []string, from, to *time.Time) iter.Seq2[Daily, error] {
	const limit int32 = 1000

	return func(yield func(Daily, error) bool) {
		var (
			offset int32 = 0
			total        = -1
		)
		for {
			// Fetch one page of EOD data
			page, err := c.fetchEODPage(ctx, symbols, offset, limit, from, to)
			if err != nil {
				yield(Daily{}, fmt.Errorf("GetEOD failed: %w", err))
				return
			}
			// On the first page, we get the total number of records to calculate how many
			// pages we need to fetch. We also deduct the appropriate number of tokens
			// from the provider based on the total.
			if total == -1 {
				err := c.deductEODTokens(ctx, page.Pagination.Total, limit)
				if err != nil {
					yield(Daily{}, fmt.Errorf("GetEOD failed: %w", err))
					return
				}
				total = page.Pagination.Total
			}
			// If the page is empty, we are done.
			if page.Pagination.Count == 0 || len(page.Data) == 0 {
				return
			}
			// Yield the histories from this page. If the consumer signals to stop, we return early.
			if !c.yieldDailies(yield, page) {
				return
			}
			// Move the offset for the next page. If we've reached or exceeded the total, we are done.
			offset += int32(page.Pagination.Count)
			if total >= 0 && int(offset) >= total {
				return
			}
		}
	}
}

// fetchEODPage does exactly one HTTP request for one page and decodes it.
// It constructs the request with the appropriate query parameters for symbols, limit, and offset, and executes it.
// It returns an error if the request fails or if the response cannot be decoded, which should cause GetEOD to yield an error and stop iteration.
func (c *MarketStackClient) fetchEODPage(ctx context.Context, symbols []string, offset, limit int32, from, to *time.Time) (marketstackEODResponse, error) {
	req, err := c.constructEODRequest(ctx, symbols, offset, limit, from, to)
	if err != nil {
		return marketstackEODResponse{}, fmt.Errorf("fetchEODPage failed to construct request: %w", err)
	}

	res, err := c.http.Do(req)
	if err != nil {
		return marketstackEODResponse{}, fmt.Errorf("fetchEODPage failed to execute request: %w", err)
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return marketstackEODResponse{}, fmt.Errorf("fetchEODPage failed to read response body: %w", err)
	}

	if res.StatusCode >= 300 {
		return marketstackEODResponse{}, fmt.Errorf("fetchEODPage received status %d: %s", res.StatusCode, string(bodyBytes))
	}

	var parsed marketstackEODResponse
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return marketstackEODResponse{}, fmt.Errorf("fetchEODPage failed to unmarshal response: %w", err)
	}

	return parsed, nil
}

// constructEODRequest builds the HTTP request for fetching EOD data from MarketStack.
func (c *MarketStackClient) constructEODRequest(ctx context.Context, symbols []string, offset, limit int32, from, to *time.Time) (*http.Request, error) {
	provider, err := c.ps.GetByName(ctx, c.providerName)
	if err != nil {
		return nil, fmt.Errorf("constructEODRequest failed to fetch provider %s: %w", c.providerName, err)
	}
	if provider == nil {
		return nil, fmt.Errorf("constructEODRequest failed, provider %s not found", c.providerName)
	}
	if provider.BaseURI == "" {
		return nil, fmt.Errorf("constructEODRequest failed, provider %s base URI is empty", c.providerName)
	}
	if provider.ApiKey == "" {
		return nil, fmt.Errorf("constructEODRequest failed, provider %s api key is empty", c.providerName)
	}

	// Set up the request to the MarketStack API
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, provider.BaseURI+"/eod", nil)
	if err != nil {
		return nil, fmt.Errorf("constructEODRequest failed to construct request: %w", err)
	}
	// Add query parameters for symbols, limit, and access key
	q := req.URL.Query()
	q.Add("symbols", strings.Join(symbols, ","))
	q.Add("limit", fmt.Sprintf("%d", limit))
	q.Add("offset", fmt.Sprintf("%d", offset))
	q.Add("access_key", provider.ApiKey)
	if from != nil {
		q.Add("date_from", from.Format("2006-01-02"))
	}
	if to != nil {
		q.Add("date_to", to.Format("2006-01-02"))
	}
	req.URL.RawQuery = q.Encode()
	req.Header.Add("Accept", "application/json")
	return req, nil
}

// deductEODTokens calculates how many API tokens to deduct based on the total number of records and the limit per page.
// It performs a ceiling division to determine the number of requests needed and calls the provider's DeductTokens method.
// For example, if there are 2500 total records and the limit is 1000, we need 3 requests (1000 + 1000 + 500), so we deduct 3 tokens.
// It returns an error if the provider fails to deduct the tokens, which should cause GetEOD to yield an error and stop iteration.
func (c *MarketStackClient) deductEODTokens(ctx context.Context, total int, limit int32) error {
	requestCount := (int32(total) + limit - 1) / limit // ceiling division
	return c.ps.DeductTokens(ctx, c.providerName, requestCount)
}

// yieldDailies takes a MarketStack API response page and yields History objects to the consumer.
// It converts each data entry in the page to a History object and yields it. If any entry is malformed, it skips it.
// If the consumer signals to stop (by returning false), it stops yielding and returns.
func (c *MarketStackClient) yieldDailies(
	yield func(Daily, error) bool,
	page marketstackEODResponse,
) bool {
	for _, d := range page.Data {
		h, err := NewDaily(d.Symbol, d.Date.Time, d.Open, d.Close, d.High, d.Low, int64(d.Volume))
		if err != nil {
			continue // skip malformed rows
		}

		if !yield(h, nil) {
			return false
		}
	}
	return true
}

// // 'https://api.marketstack.com/v2/tickers/TDT.AS?access_key=XXXX'
// func (c *MarketStackClient) GetInformation(ctx context.Context, symbol string) (Listing, error) {

// }
