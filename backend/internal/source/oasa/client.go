// Package oasa implements the upstream sources for the OASA Telematics API.
//
// The API takes positional parameters (p1, p2) in the query string and
// answers with JSON. See docs/data-sources/oasa-telematics.md for the
// contract and the terms of use that constrain this package.
//
// The service is undocumented and has changed the type of its fields without
// notice, so responses are read leniently: a value that is expected to be a
// string may arrive as a number, and an error may arrive where a list was
// expected.
package oasa

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

	// maxAttempts is the number of tries for a request that fails in a way
	// that may be transient.
	maxAttempts = 3

	// backoffBase is the delay before the second attempt; it doubles for
	// every further attempt.
	backoffBase = 500 * time.Millisecond

	// rejectionCooldown is how long the client refuses to talk to the
	// upstream after it answers 403 or 429. OASA rate limits aggressively,
	// and continuing to hammer a limited endpoint risks a longer ban.
	rejectionCooldown = 10 * time.Minute
)

// Errors returned by this package.
var (
	// ErrRequestFailed means the upstream call did not succeed.
	ErrRequestFailed = errors.New("oasa: request failed")
	// ErrRejected means the upstream refused the request, usually because
	// the client exceeded its rate limit.
	ErrRejected = errors.New("oasa: request rejected")
	// ErrUpstreamError means the upstream answered with an error object
	// instead of the expected payload.
	ErrUpstreamError = errors.New("oasa: upstream error")
)

// object is a decoded JSON object. Numbers are kept as json.Number so that
// identifiers are never reformatted.
type object map[string]any

// str reads a field as text, accepting a string, a number, a boolean or
// null.
func (o object) str(key string) string {
	switch value := o[key].(type) {
	case nil:
		return ""
	case string:
		return value
	case json.Number:
		return value.String()
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(value)
	default:
		return fmt.Sprint(value)
	}
}

// Client talks to the OASA Telematics API.
//
// It identifies itself, applies a timeout, limits the request rate and backs
// off when the operator rejects a request. Requests are cached by the layer
// above, not here.
type Client struct {
	baseURL string
	http    *http.Client
	limiter *rate.Limiter
	ua      string

	// rejectedUntil is the time until which the upstream is considered to be
	// refusing requests.
	rejectedUntil time.Time
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

// list performs an API call that is expected to answer with an array.
func (c *Client) list(ctx context.Context, act string, params []string) ([]object, error) {
	body, err := c.call(ctx, act, params)
	if err != nil {
		return nil, err
	}

	var items []object
	if err := decode(body, &items); err != nil {
		// The upstream answers with an empty string when a query has no
		// results, rather than with an empty array. That is a valid answer,
		// not a failure.
		var text string
		if decodeErr := decode(body, &text); decodeErr == nil && strings.TrimSpace(text) == "" {
			return nil, nil
		}
		if message, ok := upstreamError(body); ok {
			return nil, fmt.Errorf("%w: %s", ErrUpstreamError, message)
		}
		return nil, fmt.Errorf("%w: decoding %s response: %w", ErrRequestFailed, act, err)
	}
	return items, nil
}

// object performs an API call that is expected to answer with an object.
func (c *Client) object(ctx context.Context, act string, params []string) (object, error) {
	body, err := c.call(ctx, act, params)
	if err != nil {
		return nil, err
	}

	var payload object
	if err := decode(body, &payload); err != nil {
		if message, ok := upstreamError(body); ok {
			return nil, fmt.Errorf("%w: %s", ErrUpstreamError, message)
		}
		return nil, fmt.Errorf("%w: decoding %s response: %w", ErrRequestFailed, act, err)
	}
	return payload, nil
}

// decode reads JSON while keeping numbers as json.Number, so that an
// identifier sent as a number is never reformatted on the way through.
func decode(body []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	return decoder.Decode(target)
}

// upstreamError reports the message of an error object answered by the
// upstream, for example {"error":"nothing requested"}.
func upstreamError(body []byte) (string, bool) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", false
	}
	message, ok := payload["error"].(string)
	if !ok || message == "" {
		return "", false
	}
	return message, true
}

// call performs the request, retrying transient failures.
func (c *Client) call(ctx context.Context, act string, params []string) ([]byte, error) {
	if wait := time.Until(c.rejectedUntil); wait > 0 {
		return nil, fmt.Errorf("%w: the operator is rate limiting this client, retrying in %s", ErrRejected, wait.Round(time.Second))
	}

	query := url.Values{}
	query.Set("act", act)
	for i, value := range params {
		query.Set("p"+strconv.Itoa(i+1), value)
	}
	endpoint := c.baseURL + "?" + query.Encode()

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			delay := backoffBase << (attempt - 1)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		body, err := c.attempt(ctx, endpoint)
		if err == nil {
			return body, nil
		}
		lastErr = err

		if errors.Is(err, ErrRejected) {
			c.rejectedUntil = time.Now().Add(rejectionCooldown)
			return nil, err
		}
		if !errors.Is(err, ErrRequestFailed) {
			return nil, err
		}
	}

	return nil, lastErr
}

// attempt performs a single request.
func (c *Client) attempt(ctx context.Context, endpoint string) ([]byte, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRequestFailed, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRequestFailed, err)
	}
	req.Header.Set("User-Agent", c.ua)
	req.Header.Set("Accept", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRequestFailed, err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRequestFailed, err)
	}

	switch {
	case res.StatusCode == http.StatusForbidden || res.StatusCode == http.StatusTooManyRequests:
		return nil, fmt.Errorf("%w: status %d", ErrRejected, res.StatusCode)
	case res.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("%w: status %d", ErrRequestFailed, res.StatusCode)
	}

	return body, nil
}

// splitList returns the items of a field that holds an array.
func splitList(value any) []object {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	result := make([]object, 0, len(items))
	for _, item := range items {
		if entry, ok := item.(map[string]any); ok {
			result = append(result, object(entry))
		}
	}
	return result
}

// trim removes surrounding whitespace.
func trim(value string) string {
	return strings.TrimSpace(value)
}
