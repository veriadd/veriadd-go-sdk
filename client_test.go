package veriadd_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	veriadd "github.com/veriadd/veriadd-go-sdk"
	"github.com/veriadd/veriadd-go-sdk/service/verify"
)

func jsonBody(t *testing.T, status int, body any) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(body)
	}
}

func TestNewRequiresAPIKey(t *testing.T) {
	if _, err := veriadd.New(""); err == nil {
		t.Fatal("expected error for empty apiKey")
	}
}

func TestVerifyPostsPayloadWithAuthHeaders(t *testing.T) {
	var gotMethod, gotPath, gotKey, gotReqID, gotCT string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		gotKey, gotReqID, gotCT = r.Header.Get("X-API-Key"), r.Header.Get("X-Request-ID"), r.Header.Get("Content-Type")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		jsonBody(t, 200, map[string]any{"data": map[string]any{
			"audit_id": "audit-1", "status": "verified", "confidence": 92,
			"reasons":            []string{"postcode valid in NIPOST registry (+40)"},
			"postcode_canonical": "LA-11-W06-TC-10",
			"nipost":             map[string]any{"postcode": "LA-11-W06-TC-10", "valid": true},
			"identity":           map[string]any{"provider": "dojah", "bvn_valid": true},
			"billed_kobo":        5000, "billed_ngn": 50,
		}})(w, r)
	}))
	defer srv.Close()

	c, err := veriadd.New("vr_live_test", veriadd.WithBaseURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	res, err := c.Verify.VerifyAddress(context.Background(), verify.Request{Postcode: "LA-11-W06-TC-10", Level: 3})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "verified" || res.Confidence != 92 {
		t.Fatalf("unexpected result: %+v", res)
	}
	if res.Identity == nil || res.Identity.BVNValid == nil || !*res.Identity.BVNValid {
		t.Fatalf("identity not parsed: %+v", res.Identity)
	}
	if gotMethod != "POST" || gotPath != "/api/v1/verify/address" {
		t.Fatalf("unexpected request: %s %s", gotMethod, gotPath)
	}
	if gotKey != "vr_live_test" || gotReqID == "" || gotCT != "application/json" {
		t.Fatalf("bad headers: key=%q reqid=%q ct=%q", gotKey, gotReqID, gotCT)
	}
	if gotBody["postcode"] != "LA-11-W06-TC-10" || gotBody["level"] != float64(3) {
		t.Fatalf("bad body: %v", gotBody)
	}
	// omitempty: unset optionals must be absent, not null
	for _, k := range []string{"street", "bvn", "phone"} {
		if _, ok := gotBody[k]; ok {
			t.Fatalf("expected %q absent, got %v", k, gotBody[k])
		}
	}
}

func TestBaseURLTrimmed(t *testing.T) {
	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotQuery = r.URL.Path, r.URL.RawQuery
		jsonBody(t, 200, map[string]any{"data": map[string]any{"postcode": "X", "valid": true}})(w, r)
	}))
	defer srv.Close()

	c, _ := veriadd.New("k", veriadd.WithBaseURL(srv.URL+"/"))
	l, err := c.Search.Lookup(context.Background(), "LA-11-W06-TC-10", 1)
	if err != nil {
		t.Fatal(err)
	}
	if !l.Valid {
		t.Fatal("expected valid lookup")
	}
	if gotPath != "/api/v1/lookup" || !strings.Contains(gotQuery, "code=LA-11-W06-TC-10") || !strings.Contains(gotQuery, "level=1") {
		t.Fatalf("bad path/query: %s?%s", gotPath, gotQuery)
	}
}

func TestErrorEnvelopeMapped(t *testing.T) {
	srv := httptest.NewServer(jsonBody(t, 402, map[string]any{"error": map[string]any{
		"code": "insufficient_credits", "message": "top up", "request_id": "req-1",
	}}))
	defer srv.Close()

	c, _ := veriadd.New("k", veriadd.WithBaseURL(srv.URL))
	_, err := c.Billing.Wallet(context.Background())
	verr, ok := err.(*veriadd.Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if verr.Code != "insufficient_credits" || verr.Status != 402 || verr.RequestID != "req-1" {
		t.Fatalf("bad mapping: %+v", verr)
	}
}

func TestNonJSONErrorBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(502)
		_, _ = w.Write([]byte("gateway exploded"))
	}))
	defer srv.Close()

	c, _ := veriadd.New("k", veriadd.WithBaseURL(srv.URL))
	_, err := c.Billing.Wallet(context.Background())
	verr, ok := err.(*veriadd.Error)
	if !ok || verr.Status != 502 {
		t.Fatalf("expected 502 *Error, got %v", err)
	}
}

func TestTimeoutMaps(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		jsonBody(t, 200, map[string]any{"data": map[string]any{}})(w, r)
	}))
	defer srv.Close()

	c, _ := veriadd.New("k", veriadd.WithBaseURL(srv.URL), veriadd.WithTimeout(30*time.Millisecond))
	_, err := c.Billing.Wallet(context.Background())
	verr, ok := err.(*veriadd.Error)
	if !ok || verr.Code != "network_error" {
		t.Fatalf("expected network_error *Error, got %v", err)
	}
}

func TestCreateKeyPostsEnv(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		jsonBody(t, 201, map[string]any{"data": map[string]any{
			"api_key": "vr_live_x", "key_prefix": "vr_live_x", "warning": "once",
		}})(w, r)
	}))
	defer srv.Close()

	c, _ := veriadd.New("k", veriadd.WithBaseURL(srv.URL))
	k, err := c.Keys.Create(context.Background(), "test", "Staging")
	if err != nil {
		t.Fatal(err)
	}
	if k.APIKey != "vr_live_x" {
		t.Fatalf("bad key: %+v", k)
	}
	if gotBody["env"] != "test" || gotBody["name"] != "Staging" {
		t.Fatalf("bad body: %v", gotBody)
	}
}

func TestRetrySucceedsAfterRetryable(t *testing.T) {
	n := 0
	out, err := veriadd.DoWithRetry(context.Background(), func() (string, error) {
		n++
		if n < 3 {
			return "", &veriadd.Error{Code: "rate_limited", Status: 429, Message: "slow"}
		}
		return "ok", nil
	}, veriadd.RetryOptions{MaxRetries: 5, BaseDelay: time.Millisecond})
	if err != nil || out != "ok" || n != 3 {
		t.Fatalf("out=%q err=%v n=%d", out, err, n)
	}
}

func TestRetryBubblesNonRetryable(t *testing.T) {
	n := 0
	_, err := veriadd.DoWithRetry(context.Background(), func() (string, error) {
		n++
		return "", &veriadd.Error{Code: "invalid_api_key", Status: 401, Message: "bad"}
	}, veriadd.RetryOptions{BaseDelay: time.Millisecond})
	verr, ok := err.(*veriadd.Error)
	if !ok || verr.Code != "invalid_api_key" || n != 1 {
		t.Fatalf("err=%v n=%d", err, n)
	}
}

func TestRetryGivesUp(t *testing.T) {
	n := 0
	_, err := veriadd.DoWithRetry(context.Background(), func() (string, error) {
		n++
		return "", &veriadd.Error{Code: "nipost_rate_limited", Status: 429, Message: "slow"}
	}, veriadd.RetryOptions{MaxRetries: 2, BaseDelay: time.Millisecond})
	if err == nil || n != 3 {
		t.Fatalf("err=%v n=%d", err, n)
	}
}
