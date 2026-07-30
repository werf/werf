//go:build !embedwerfdeno

package deno

import "context"

// EmbeddedBinaryPath reports that no embedded Deno binary is available, so
// callers fall back to letting nelm download one.
func EmbeddedBinaryPath(_ context.Context) (string, bool, error) {
	return "", false, nil
}
