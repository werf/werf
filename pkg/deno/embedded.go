//go:build embedwerfdeno

package deno

import "strings"

// EmbeddedDenoData returns the compressed embedded Deno binary and its
// expected sha256. The last result reports whether an embedded binary is
// available at all.
func EmbeddedDenoData() ([]byte, string, bool) {
	return embeddedDeno, strings.TrimSpace(embeddedDenoSHA256), true
}
