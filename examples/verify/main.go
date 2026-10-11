// Command verify runs a live L1 lookup against the Veriadd API.
// Requires a test key: VERIADD_KEY=vr_test_... go run ./examples/verify
package main

import (
	"context"
	"fmt"
	"os"

	veriadd "github.com/veriadd/veriadd-go-sdk"
	"github.com/veriadd/veriadd-go-sdk/service/verify"
)

func main() {
	key := os.Getenv("VERIADD_KEY")
	if key == "" {
		fmt.Println("Set VERIADD_KEY=vr_test_... and retry.")
		return
	}
	base := os.Getenv("VERIADD_BASE_URL")
	client, err := veriadd.New(key, veriadd.WithBaseURL(orDefault(base, "https://api.veriadd.tech")))
	if err != nil {
		panic(err)
	}
	ctx := context.Background()

	lookup, err := client.Search.Lookup(ctx, "LA-11-W06-TC-10", 1)
	if err != nil {
		panic(err)
	}
	fmt.Printf("lookup: %s valid=%v\n", lookup.Postcode, lookup.Valid)

	result, err := client.Verify.VerifyAddress(ctx, verify.Request{
		Postcode: "LA-11-W06-TC-10",
		Level:    1,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("verify: %s confidence=%d\n", result.Status, result.Confidence)
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
