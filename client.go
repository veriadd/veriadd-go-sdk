package veriadd

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/veriadd/veriadd-go-sdk/internal/transport"
	"github.com/veriadd/veriadd-go-sdk/service/keys"
	"github.com/veriadd/veriadd-go-sdk/service/kyb"
	"github.com/veriadd/veriadd-go-sdk/service/search"
	"github.com/veriadd/veriadd-go-sdk/service/verify"
	"github.com/veriadd/veriadd-go-sdk/service/wallet"
)

// DefaultBaseURL is the production host. Override with WithBaseURL for
// self-hosted backends or local development.
const DefaultBaseURL = "https://api.veriadd.tech"

// Client is the entry point of the SDK. Pick the service for the capability
// you need — every service shares one transport, so auth, timeouts, error
// envelopes and request IDs behave identically everywhere.
type Client struct {
	Verify  *verify.Service
	Search  *search.Service
	Billing *wallet.Service
	KYB     *kyb.Service
	Keys    *keys.Service

	transport *transport.Client
}

// Option configures a Client.
type Option func(*clientConfig)

type clientConfig struct {
	baseURL    string
	timeout    time.Duration
	httpClient *http.Client
}

// WithBaseURL points the client at a self-hosted backend or local dev server.
func WithBaseURL(url string) Option {
	return func(cfg *clientConfig) { cfg.baseURL = strings.TrimSuffix(url, "/") }
}

// WithTimeout sets the per-request timeout (default 15s).
func WithTimeout(d time.Duration) Option {
	return func(cfg *clientConfig) { cfg.timeout = d }
}

// WithHTTPClient injects a custom transport (proxies, mocks in tests).
// When set, the timeout option is ignored — configure it on your client.
func WithHTTPClient(h *http.Client) Option {
	return func(cfg *clientConfig) { cfg.httpClient = h }
}

// New builds a client. The apiKey is a secret (`vr_live_…` / `vr_test_…`) —
// never ship it in client-side code; mobile apps must call your backend.
func New(apiKey string, opts ...Option) (*Client, error) {
	if apiKey == "" {
		return nil, &Error{Code: "missing_api_key", Message: "veriadd.New requires an apiKey"}
	}
	cfg := &clientConfig{baseURL: DefaultBaseURL, timeout: 15 * time.Second}
	for _, o := range opts {
		o(cfg)
	}
	t := transport.New(apiKey, cfg.baseURL, cfg.timeout, cfg.httpClient)
	return &Client{
		Verify:    verify.New(t),
		Search:    search.New(t),
		Billing:   wallet.New(t),
		KYB:       kyb.New(t),
		Keys:      keys.New(t),
		transport: t,
	}, nil
}

// Status is the live dependency snapshot from GET /v1/status.
type Status struct {
	Service       string         `json:"service"`
	Version       string         `json:"version"`
	Time          string         `json:"time"`
	UptimeSeconds int64          `json:"uptime_seconds"`
	Deps          map[string]any `json:"deps"`
}

// Status returns live dependency status (public — key is harmless if set).
func (c *Client) Status(ctx context.Context) (*Status, error) {
	raw, err := c.transport.Do(ctx, "GET", "/api/v1/status", nil, nil)
	if err != nil {
		return nil, err
	}
	return transport.Decode(raw, &Status{})
}
