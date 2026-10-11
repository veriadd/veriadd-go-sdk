// Package search covers postcode discovery: graded lookups, typeahead,
// geo search and postcode assembly.
package search

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/veriadd/veriadd-go-sdk/internal/transport"
	"github.com/veriadd/veriadd-go-sdk/service/verify"
)

// Service searches the postcode registry. Build it via [veriadd.New], not directly.
type Service struct {
	t *transport.Client
}

// New wires a Service to shared transport. Prefer [veriadd.New].
func New(t *transport.Client) *Service { return &Service{t: t} }

// Lookup resolves a postcode. Level 1 is free.
func (s *Service) Lookup(ctx context.Context, code string, level int) (*verify.Lookup, error) {
	if level <= 0 {
		level = 1
	}
	raw, err := s.t.Do(ctx, "GET", "/api/v1/lookup", map[string]string{
		"code":  code,
		"level": strconv.Itoa(level),
	}, nil)
	if err != nil {
		return nil, err
	}
	return transport.Decode(raw, &verify.Lookup{})
}

// Autocomplete is segment-aware postcode typeahead.
func (s *Service) Autocomplete(ctx context.Context, q string) (any, error) {
	raw, err := s.t.Do(ctx, "GET", "/api/v1/search/autocomplete", map[string]string{"q": q}, nil)
	if err != nil {
		return nil, err
	}
	var out any
	return out, json.Unmarshal(raw, &out)
}

// NearbyOptions scopes a nearby search. Radius is metres (default 300).
type NearbyOptions struct {
	Lat    float64
	Lng    float64
	Radius *float64
}

// Nearby finds postcodes within a radius of a coordinate.
func (s *Service) Nearby(ctx context.Context, o NearbyOptions) (any, error) {
	q := map[string]string{"lat": fmt.Sprint(o.Lat), "lng": fmt.Sprint(o.Lng)}
	if o.Radius != nil {
		q["radius"] = fmt.Sprint(*o.Radius)
	}
	raw, err := s.t.Do(ctx, "GET", "/api/v1/search/nearby", q, nil)
	if err != nil {
		return nil, err
	}
	var out any
	return out, json.Unmarshal(raw, &out)
}

// ReverseOptions scopes a reverse-geocode. MaxDistanceM defaults to 25, max 250.
type ReverseOptions struct {
	Lat          float64
	Lng          float64
	MaxDistanceM *float64
}

// Reverse resolves a coordinate to the nearest postcode.
func (s *Service) Reverse(ctx context.Context, o ReverseOptions) (any, error) {
	q := map[string]string{"lat": fmt.Sprint(o.Lat), "lng": fmt.Sprint(o.Lng)}
	if o.MaxDistanceM != nil {
		q["max_distance_m"] = fmt.Sprint(*o.MaxDistanceM)
	}
	raw, err := s.t.Do(ctx, "GET", "/api/v1/search/reverse", q, nil)
	if err != nil {
		return nil, err
	}
	var out any
	return out, json.Unmarshal(raw, &out)
}

// Disassemble parses a postcode into segments (native NIPOST shape).
func (s *Service) Disassemble(ctx context.Context, code string) (any, error) {
	raw, err := s.t.Do(ctx, "GET", "/api/v1/assembly/disassemble", map[string]string{"code": code}, nil)
	if err != nil {
		return nil, err
	}
	var out any
	return out, json.Unmarshal(raw, &out)
}

// AssembleInput builds a canonical postcode from segments.
type AssembleInput struct {
	State    string `json:"state"`
	LGA      string `json:"lga"`
	District string `json:"district"`
	Area     string `json:"area"`
	Unit     string `json:"unit"`
}

// Assemble builds segments into a canonical postcode.
func (s *Service) Assemble(ctx context.Context, in AssembleInput) (any, error) {
	raw, err := s.t.Do(ctx, "POST", "/api/v1/assembly/assemble", nil, in)
	if err != nil {
		return nil, err
	}
	var out any
	return out, json.Unmarshal(raw, &out)
}
