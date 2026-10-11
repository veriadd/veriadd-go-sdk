// Package kyb covers business verification: sandbox-by-default workspaces,
// RC auto-checks against CAC, and the review lifecycle.
package kyb

import (
	"context"

	"github.com/veriadd/veriadd-go-sdk/internal/transport"
)

// Service verifies businesses. Build it via [veriadd.New], not directly.
type Service struct {
	t *transport.Client
}

// New wires a Service to shared transport. Prefer [veriadd.New].
func New(t *transport.Client) *Service { return &Service{t: t} }

type KYB struct {
	ID                string  `json:"id"`
	BusinessName      string  `json:"business_name"`
	RCNumber          string  `json:"rc_number"`
	RegisteredAddress string  `json:"registered_address"`
	Website           string  `json:"website"`
	UseCase           string  `json:"use_case"`
	CompanyType       string  `json:"company_type"`
	CACVerified       bool    `json:"cac_verified"`
	CACLegalName      *string `json:"cac_legal_name"`
	Status            string  `json:"status"`
	ReviewNote        *string `json:"review_note"`
}

// Input submits (or resubmits) KYB business details. CompanyType is one of
// BUSINESS_NAME, COMPANY, INCORPORATED_TRUSTEES, LIMITED_PARTNERSHIP,
// LIMITED_LIABILITY_PARTNERSHIP.
type Input struct {
	BusinessName      string `json:"business_name"`
	RCNumber          string `json:"rc_number"`
	CompanyType       string `json:"company_type"`
	RegisteredAddress string `json:"registered_address,omitempty"`
	Website           string `json:"website,omitempty"`
	UseCase           string `json:"use_case,omitempty"`
}

// Get returns the workspace submission (nil when never submitted).
func (s *Service) Get(ctx context.Context) (*KYB, error) {
	raw, err := s.t.Do(ctx, "GET", "/api/v1/kyb", nil, nil)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	return transport.Decode(raw, &KYB{})
}

// Submit sends business details for review.
func (s *Service) Submit(ctx context.Context, in Input) (*KYB, error) {
	raw, err := s.t.Do(ctx, "POST", "/api/v1/kyb", nil, in)
	if err != nil {
		return nil, err
	}
	return transport.Decode(raw, &KYB{})
}
