//go:build embedwerfdeno

package deno

import (
	"context"
	"strings"

	"github.com/werf/nelm/pkg/ts"
)

// EmbeddedBinaryPath extracts the embedded Deno binary and returns its path.
// The second result reports whether an embedded binary is available at all.
func EmbeddedBinaryPath(ctx context.Context) (string, bool, error) {
	path, err := ts.ExtractEmbeddedDeno(ctx, embeddedDeno, strings.TrimSpace(embeddedDenoSHA256))
	if err != nil {
		return "", false, err
	}

	return path, true, nil
}
