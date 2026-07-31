//go:build embedwerfdeno

package deno

import (
	"runtime"

	"github.com/werf/nelm/pkg/ts/denolock"
)

// EmbeddedDenoData returns the compressed embedded Deno binary. The last
// result reports whether an embedded binary is available at all.
func EmbeddedDenoData() ([]byte, bool) {
	// The blob is downloaded by nelm's embed-deno, which verifies it against the release nelm pins,
	// so a platform nelm does not pin cannot have produced one.
	if _, err := denolock.Get(runtime.GOOS, runtime.GOARCH); err != nil {
		return nil, false
	}

	return embeddedDeno, true
}
