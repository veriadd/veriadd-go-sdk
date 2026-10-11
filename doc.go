// Package veriadd is the official Go client for the Veriadd address KYC API.
//
// Pick the service for the capability you need:
//
//	client, err := veriadd.New(os.Getenv("VERIADD_KEY"))
//	if err != nil {
//		log.Fatal(err)
//	}
//	result, err := client.Verify.VerifyAddress(ctx, verify.Request{
//		Postcode: "LA-11-W06-TC-10",
//		State:    "LAGOS",
//		Level:    3,
//	})
package veriadd
