package veriadd

import (
	"context"
	"time"

	"github.com/veriadd/veriadd-go-sdk/internal/transport"
)

// RetryOptions tunes DoWithRetry.
type RetryOptions struct {
	// MaxRetries caps retries after the first attempt. The zero value means
	// defaults (3 retries); pass a negative number with other fields set to
	// disable retrying.
	MaxRetries int
	// Retryable overrides the code set (default transport.RetryableCodes).
	Retryable map[string]bool
	// BaseDelay doubles every attempt (default 1s).
	BaseDelay time.Duration
}

// RetryableCodes are worth retrying with backoff (transient upstream/rate states).
var RetryableCodes = transport.RetryableCodes

// DoWithRetry runs fn with exponential backoff on retryable *Error codes.
// Non-retryable errors return immediately. Context cancellation aborts waits.
func DoWithRetry[T any](ctx context.Context, fn func() (T, error), opts RetryOptions) (T, error) {
	maxRetries := opts.MaxRetries
	retryable := opts.Retryable
	base := opts.BaseDelay
	if maxRetries == 0 && retryable == nil && base == 0 {
		maxRetries = 3
	}
	if retryable == nil {
		retryable = transport.RetryableCodes
	}
	if base <= 0 {
		base = time.Second
	}
	var zero T
	for attempt := 0; ; attempt++ {
		out, err := fn()
		if err == nil {
			return out, nil
		}
		var verr *transport.Error
		if transport.AsError(err, &verr) {
			if attempt >= maxRetries || !retryable[verr.Code] {
				return zero, err
			}
		} else {
			return zero, err
		}
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-time.After(base << attempt):
		}
	}
}
