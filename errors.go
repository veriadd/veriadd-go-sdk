package veriadd

import "github.com/veriadd/veriadd-go-sdk/internal/transport"

// Error is the typed API failure. It aliases the transport error so every
// service surfaces the identical type: match it with errors.As.
//
//	var verr *veriadd.Error
//	if errors.As(err, &verr) && verr.Code == "insufficient_credits" { ... }
type Error = transport.Error
