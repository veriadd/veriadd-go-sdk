// Package wallet covers billing: balances, metered history and Bachs top-ups.
package wallet

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/veriadd/veriadd-go-sdk/internal/transport"
)

// Service bills and funds workspaces. Build it via [veriadd.New], not directly.
type Service struct {
	t *transport.Client
}

// New wires a Service to shared transport. Prefer [veriadd.New].
func New(t *transport.Client) *Service { return &Service{t: t} }

type Wallet struct {
	Client         string   `json:"client"`
	Email          string   `json:"email"`
	BalanceKobo    int64    `json:"balance_kobo"`
	BalanceNGN     float64  `json:"balance_ngn"`
	PriceL2Kobo    int64    `json:"price_l2_kobo"`
	PriceL3Kobo    int64    `json:"price_l3_kobo"`
	TopupProviders []string `json:"topup_providers"`
}

type UsageRow struct {
	Endpoint   string `json:"endpoint"`
	Status     string `json:"status"`
	BilledKobo int64  `json:"billed_kobo"`
	Confidence *int   `json:"confidence"`
	CreatedAt  string `json:"created_at"`
}

type TopupInit struct {
	Provider         string `json:"provider"`
	AuthorizationURL string `json:"authorization_url"`
	Reference        string `json:"reference"`
	CheckoutID       string `json:"checkout_id"`
}

type TopupVerify struct {
	CreditedKobo    int64   `json:"credited_kobo"`
	AlreadyCredited bool    `json:"already_credited"`
	BalanceKobo     int64   `json:"balance_kobo"`
	BalanceNGN      float64 `json:"balance_ngn"`
}

// Wallet returns balance + pricing for this key.
func (s *Service) Wallet(ctx context.Context) (*Wallet, error) {
	raw, err := s.t.Do(ctx, "GET", "/api/v1/wallet", nil, nil)
	if err != nil {
		return nil, err
	}
	return transport.Decode(raw, &Wallet{})
}

// Usage returns metered call history (1–200, default 50).
func (s *Service) Usage(ctx context.Context, limit int) ([]UsageRow, error) {
	if limit <= 0 {
		limit = 50
	}
	raw, err := s.t.Do(ctx, "GET", "/api/v1/usage", map[string]string{"limit": strconv.Itoa(limit)}, nil)
	if err != nil {
		return nil, err
	}
	var out []UsageRow
	if len(raw) == 0 || string(raw) == "null" {
		return out, nil
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, &transport.Error{Code: "bad_response", Message: "unexpected response shape: " + err.Error()}
	}
	return out, nil
}

// TopupInit starts a Bachs wallet top-up; redirect the user to AuthorizationURL.
func (s *Service) TopupInit(ctx context.Context, amountKobo int64, email, callbackURL string) (*TopupInit, error) {
	body := map[string]any{"amount_kobo": amountKobo, "email": email}
	if callbackURL != "" {
		body["callback_url"] = callbackURL
	}
	raw, err := s.t.Do(ctx, "POST", "/api/v1/wallet/topup/initialize", nil, body)
	if err != nil {
		return nil, err
	}
	return transport.Decode(raw, &TopupInit{})
}

// TopupVerify confirms a top-up reference and credits the wallet (idempotent).
func (s *Service) TopupVerify(ctx context.Context, reference string) (*TopupVerify, error) {
	raw, err := s.t.Do(ctx, "GET", "/api/v1/wallet/topup/verify", map[string]string{"reference": reference}, nil)
	if err != nil {
		return nil, err
	}
	return transport.Decode(raw, &TopupVerify{})
}
