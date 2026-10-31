// Package oasa implements the upstream sources for the OASA Telematics API.
//
// The API takes positional parameters (p1, p2) in the query string and
// answers with JSON. See docs/data-sources/oasa-telematics.md for the
// contract and the terms of use that constrain this package.
package oasa

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"golang.org/x/time/rate"
)

// DefaultBaseURL is the upstream service address.
const DefaultBaseURL = "https://telematics.oasa.gr/api/"

// DefaultUserAgent identifies this project to the upstream operator.
const DefaultUserAgent = "leoforio/0.1 (+https://github.com/ksohos333/leoforio)"

const (
	defaultTimeout   = 15 * time.Second
	defaultRate      = 4.0
	defaultRateBurst = 8
)

// Client talks to the OASA Telematics API.
//
// It identifies itself, applies a timeout, and limits the request rate so
// that the upstream is not overwhelmed. Requests are cached by the layer
// above, not here.
type Client struct {
	baseURL string
	http    *http.Client
	limiter *rate.Limiter
	ua      string
}

// NewClient returns a client for the given base URL. An empty base URL uses
// DefaultBaseURL.
func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: defaultTimeout},
		limiter: rate.NewLimiter(rate.Limit(defaultRate), defaultRateBurst),
		ua:      DefaultUserAgent,
	}
}

// get performs an API call and decodes the JSON response into target.
//
// Parameters are positional: the first becomes p1, the second p2 and so on.
// The upstream ignores named parameters and answers with an error, so the
// order matters.
func (c *Client) get(ctx context.Context, act string, params []string, target any) error {
	if err := c.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("%w: %w", ErrRequestFailed, err)
	}

	query := url.Values{}
	query.Set("act", act)
	for i, value := range params {
		query.Set("p"+strconv.Itoa(i+1), value)
	}

	endpoint := c.baseURL + "?" + query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRequestFailed, err)
	}
	req.Header.Set("User-Agent", c.ua)
	req.Header.Set("Accept", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRequestFailed, err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRequestFailed, err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: status %d", ErrRequestFailed, res.StatusCode)
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("%w: decoding %s response: %w", ErrRequestFailed, act, err)
	}

	return nil
}
