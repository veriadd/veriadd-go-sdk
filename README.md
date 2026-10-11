# Veriadd Go SDK — `github.com/veriadd/veriadd-go-sdk`

Official Go client for the Veriadd address KYC API. Standard library only —
no runtime dependencies. Requires Go 1.21+.

## Install

```bash
go get github.com/veriadd/veriadd-go-sdk
```

## Quickstart

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	veriadd "github.com/veriadd/veriadd-go-sdk"
	"github.com/veriadd/veriadd-go-sdk/service/verify"
)

func main() {
	client, err := veriadd.New(os.Getenv("VERIADD_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	// Primary KYC endpoint: postcode + identity cross-check
	result, err := client.Verify.VerifyAddress(context.Background(), verify.Request{
		Postcode: "LA-11-W06-TC-10",
		State:    "LAGOS",
		BVN:      "22233344455",
		Phone:    "08031234567",
		Level:    3, // L1 free, L2 ₦30, L3 ₦50
	})
	if err != nil {
		log.Fatal(err)
	}
	if result.Status == "verified" && result.Confidence >= 80 {
		// proceed to onboarding — result.AuditID is your CBN trail
	}
	fmt.Println(result.Status, result.Confidence)
}
```

> **Secret keys stay server-side.** A `vr_live_…` key inside a shipped app
> can be extracted. Mobile apps should call your backend, which holds the key.

Point at a self-hosted backend with `veriadd.WithBaseURL`, inject transports
with `veriadd.WithHTTPClient`, tune timeouts with `veriadd.WithTimeout`.

## Error handling

Every failure returns `*VeriaddError` with `Code`, `Status` and `RequestID`:

```go
result, err := client.Verify.VerifyAddress(ctx, req)
if err != nil {
	var verr *veriadd.Error
	if errors.As(err, &verr) && verr.Code == "insufficient_credits" {
		// pause onboarding, trigger a wallet top-up
	}
}

// Retry transient states with backoff (context-aware)
result, err := veriadd.DoWithRetry(ctx, func() (*verify.Result, error) {
	return client.Verify.VerifyAddress(ctx, req)
}, veriadd.RetryOptions{})
```

## Methods

| Service | Methods | Endpoint prefix |
|---|---|---|
| `client.Verify` | `VerifyAddress` | `POST /v1/verify/address` |
| `client.Search` | `Lookup`, `Autocomplete`, `Nearby`, `Reverse`, `Disassemble`, `Assemble` | `GET /v1/lookup`, `/v1/search/*`, assembly |
| `client.Billing` | `Wallet`, `Usage`, `TopupInit`, `TopupVerify` | wallet + history |
| `client.KYB` | `Get`, `Submit` | business verification |
| `client.Keys` | `List`, `Create`, `Revoke` | API keys |
| `client` | `Status` | live dependency status |

Full endpoint semantics: https://veriadd.tech/docs/api
