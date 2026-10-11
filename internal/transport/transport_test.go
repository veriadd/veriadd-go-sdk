package transport_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/veriadd/veriadd-go-sdk/internal/transport"
)

func newTestTransport(srv *httptest.Server) *transport.Client {
	return transport.New("k", srv.URL, time.Second, nil)
}

func TestDoSendsAuthHeaders(t *testing.T) {
	var gotKey, gotReqID, gotCT, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey, gotReqID, gotCT, gotPath = r.Header.Get("X-API-Key"), r.Header.Get("X-Request-ID"), r.Header.Get("Content-Type"), r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"ok":true}}`))
	}))
	defer srv.Close()

	c := newTestTransport(srv)
	raw, err := c.Do(context.Background(), "GET", "/api/v1/status", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != `{"ok":true}` {
		t.Fatalf("bad payload: %s", raw)
	}
	if gotKey != "k" || gotReqID == "" || gotCT != "application/json" || gotPath != "/api/v1/status" {
		t.Fatalf("bad request: key=%q reqid=%q ct=%q path=%q", gotKey, gotReqID, gotCT, gotPath)
	}
}

func TestDoMapsEnvelopeErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(402)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{
			"code": "insufficient_credits", "message": "top up", "request_id": "req-1",
		}})
	}))
	defer srv.Close()

	c := newTestTransport(srv)
	_, err := c.Do(context.Background(), "GET", "/x", nil, nil)
	verr, ok := err.(*transport.Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if verr.Code != "insufficient_credits" || verr.Status != 402 || verr.RequestID != "req-1" {
		t.Fatalf("bad mapping: %+v", verr)
	}
}

func TestDoRejectsNonJSONErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(502)
		_, _ = w.Write([]byte("gateway exploded"))
	}))
	defer srv.Close()

	c := newTestTransport(srv)
	_, err := c.Do(context.Background(), "GET", "/x", nil, nil)
	verr, ok := err.(*transport.Error)
	if !ok || verr.Status != 502 {
		t.Fatalf("expected 502 *Error, got %v", err)
	}
}

func TestDoTimesOut(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
	}))
	defer srv.Close()

	c := transport.New("k", srv.URL, 30*time.Millisecond, nil)
	_, err := c.Do(context.Background(), "GET", "/x", nil, nil)
	verr, ok := err.(*transport.Error)
	if !ok || verr.Code != "network_error" {
		t.Fatalf("expected network_error *Error, got %v", err)
	}
}

func TestDecodeNullLeavesTarget(t *testing.T) {
	type shape struct {
		Name string `json:"name"`
	}
	out, err := transport.Decode([]byte("null"), &shape{Name: "keep"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Name != "keep" {
		t.Fatalf("null payload must leave target untouched, got %+v", out)
	}
	if _, err := transport.Decode([]byte("{oops"), &shape{}); err == nil {
		t.Fatal("expected bad_response for malformed JSON")
	}
}
