// Package verify covers address verification: NIPOST registry resolution
// plus BVN/phone/NIN identity cross-checks, scored with an auditable trail.
package verify

import (
	"context"

	"github.com/veriadd/veriadd-go-sdk/internal/transport"
)

// Service verifies addresses. Build it via [veriadd.New], not directly.
type Service struct {
	t *transport.Client
}

// New wires a Service to shared transport. Prefer [veriadd.New].
func New(t *transport.Client) *Service { return &Service{t: t} }

// Request is the input for Service.VerifyAddress. Only Postcode is required.
type Request struct {
	Postcode  string   `json:"postcode"`
	Street    string   `json:"street,omitempty"`
	LGA       string   `json:"lga,omitempty"`
	State     string   `json:"state,omitempty"`
	FirstName string   `json:"first_name,omitempty"`
	LastName  string   `json:"last_name,omitempty"`
	DOB       string   `json:"dob,omitempty"` // yyyy-mm-dd
	Phone     string   `json:"phone,omitempty"`
	BVN       string   `json:"bvn,omitempty"` // 11 digits
	NIN       string   `json:"nin,omitempty"` // 11 digits
	Lat       *float64 `json:"lat,omitempty"`
	Lng       *float64 `json:"lng,omitempty"`
	// NIPOST lookup depth 1–5. Billing: L1 free, L2 ₦30, L3+ ₦50. Defaults to 3.
	Level int `json:"level,omitempty"`
}

// Lookup is the graded NIPOST payload (levels are cumulative).
type Lookup struct {
	Postcode              string         `json:"postcode"`
	Valid                 bool           `json:"valid"`
	AdministrativeAddress map[string]any `json:"administrative_address"`
	RecentHouseAddress    map[string]any `json:"recent_house_address"`
	BuildingUseStatus     string         `json:"building_use_status"`
}

// Identity is the Dojah cross-check breakdown. Fields appear when checked.
type Identity struct {
	Provider       string `json:"provider"`
	BVNValid       *bool  `json:"bvn_valid"`
	BVNNameMatch   *bool  `json:"bvn_name_match"`
	PhoneLinked    *bool  `json:"phone_linked"`
	PhoneNameMatch *bool  `json:"phone_name_match"`
	NINValid       *bool  `json:"nin_valid"`
}

type Result struct {
	AuditID           string    `json:"audit_id"`
	Status            string    `json:"status"` // verified | partial | failed | invalid
	Confidence        int       `json:"confidence"`
	Reasons           []string  `json:"reasons"`
	PostcodeCanonical string    `json:"postcode_canonical"`
	Nipost            *Lookup   `json:"nipost"`
	Identity          *Identity `json:"identity"`
	BilledKobo        int64     `json:"billed_kobo"`
	BilledNGN         float64   `json:"billed_ngn"`
}

// VerifyAddress verifies an address + optional identity cross-checks. The money endpoint.
func (s *Service) VerifyAddress(ctx context.Context, req Request) (*Result, error) {
	raw, err := s.t.Do(ctx, "POST", "/api/v1/verify/address", nil, req)
	if err != nil {
		return nil, err
	}
	return transport.Decode(raw, &Result{})
}
