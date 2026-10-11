// Package keys manages workspace API keys: listing prefixes, issuing
// named keys (full secret shown once) and immediate revocation.
package keys

import (
	"context"
	"encoding/json"

	"github.com/veriadd/veriadd-go-sdk/internal/transport"
)

// Service manages API keys. Build it via [veriadd.New], not directly.
type Service struct {
	t *transport.Client
}

// New wires a Service to shared transport. Prefer [veriadd.New].
func New(t *transport.Client) *Service { return &Service{t: t} }

type Info struct {
	ID        string `json:"id"`
	KeyPrefix string `json:"key_prefix"`
	Env       string `json:"env"`
	Name      string `json:"name"`
	Revoked   bool   `json:"revoked"`
	CreatedAt string `json:"created_at"`
}

type Created struct {
	// Shown once — store it immediately, it is never returned again.
	APIKey    string `json:"api_key"`
	KeyPrefix string `json:"key_prefix"`
	Warning   string `json:"warning"`
}

// List returns key prefixes (full keys are never returned).
func (s *Service) List(ctx context.Context) ([]Info, error) {
	raw, err := s.t.Do(ctx, "GET", "/api/v1/keys", nil, nil)
	if err != nil {
		return nil, err
	}
	var out []Info
	if len(raw) == 0 || string(raw) == "null" {
		return out, nil
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, &transport.Error{Code: "bad_response", Message: "unexpected response shape: " + err.Error()}
	}
	return out, nil
}

// Create issues a named key in env ("live", default, or "test").
func (s *Service) Create(ctx context.Context, env, name string) (*Created, error) {
	if env == "" {
		env = "live"
	}
	raw, err := s.t.Do(ctx, "POST", "/api/v1/keys", nil, map[string]string{"env": env, "name": name})
	if err != nil {
		return nil, err
	}
	return transport.Decode(raw, &Created{})
}

// Revoke kills a key by id. Immediate.
func (s *Service) Revoke(ctx context.Context, id string) error {
	_, err := s.t.Do(ctx, "POST", "/api/v1/keys/"+id+"/revoke", nil, nil)
	return err
}
