//go:build !embedwerfdeno

package deno

// EmbeddedDenoData reports that no embedded Deno binary is available, so
// callers fall back to letting nelm resolve one.
func EmbeddedDenoData() ([]byte, bool) {
	return nil, false
}
