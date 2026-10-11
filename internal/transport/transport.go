// Package transport is the internal HTTP layer for the Veriadd SDK.
//
// It is internal on purpose: integrators program against the root Client
// and the service/* packages, never against these primitives. Keeping one
// transport means auth headers, timeouts, error envelopes and request IDs
// behave identically across every service.
package transport

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Error is the typed API failure parsed from the Veriadd
// `{error: {code, message}}` envelope. Every service method returns it
// instead of raw transport errors.
type Error struct {
	// Machine-readable code, e.g. `insufficient_credits`, `kyb_required`.
	Code string
	// HTTP status of the failed call (0 for timeouts / network failures).
	Status  int
	Message string
	// Correlation id — quote it when contacting support, if present.
	RequestID string
}

func (e *Error) Error() string {
	return fmt.Sprintf("veriadd: %s (%d): %s", e.Code, e.Status, e.Message)
}

// RetryableCodes are worth retrying with backoff (transient upstream/rate states).
var RetryableCodes = map[string]bool{
	"rate_limited":         true,
	"provider_unavailable": true,
	"nipost_rate_limited":  true,
}

// AsError unwraps err to *Error, following wrap chains.
func AsError(err error, target **Error) bool {
	return errors.As(err, target)
}

// Client carries auth, base URL and timeouts for every service.
type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	Timeout    time.Duration
}

// New builds a transport. An empty baseURL selects production.
func New(apiKey, baseURL string, timeout time.Duration, httpClient *http.Client) *Client {
	if baseURL == "" {
		baseURL = "https://api.veriadd.tech"
	}
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	}
	return &Client{
		APIKey:     apiKey,
		BaseURL:    baseURL,
		HTTPClient: httpClient,
		Timeout:    timeout,
	}
}

func newRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

type envelope struct {
	Data  json.RawMessage `json:"data"`
	Error *struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"request_id"`
	} `json:"error"`
}

// Do sends one JSON request and returns the unwrapped `data` payload.
func (c *Client) Do(ctx context.Context, method, path string, query map[string]string, body any) (json.RawMessage, error) {
	u := c.BaseURL + path
	if len(query) > 0 {
		v := url.Values{}
		for k, val := range query {
			v.Set(k, val)
		}
		u += "?" + v.Encode()
	}
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, &Error{Code: "encode_error", Message: err.Error()}
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, reader)
	if err != nil {
		return nil, &Error{Code: "network_error", Message: err.Error()}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.APIKey)
	req.Header.Set("X-Request-ID", newRequestID())

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, &Error{Code: "network_error", Message: err.Error()}
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var env envelope
		_ = json.Unmarshal(raw, &env)
		code := fmt.Sprintf("http_%d", resp.StatusCode)
		msg := fmt.Sprintf("Request failed (%d)", resp.StatusCode)
		reqID := ""
		if env.Error != nil {
			if env.Error.Code != "" {
				code = env.Error.Code
			}
			if env.Error.Message != "" {
				msg = env.Error.Message
			}
			reqID = env.Error.RequestID
		}
		return nil, &Error{Code: code, Status: resp.StatusCode, Message: msg, RequestID: reqID}
	}
	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, &Error{Code: "bad_response", Message: "invalid JSON response: " + err.Error()}
	}
	return env.Data, nil
}

// Decode unmarshals a `data` payload into out. Null/empty payloads leave
// out untouched and succeed, so optional objects decode cleanly.
func Decode[T any](raw json.RawMessage, out *T) (*T, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return out, nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return nil, &Error{Code: "bad_response", Message: "unexpected response shape: " + err.Error()}
	}
	return out, nil
}
