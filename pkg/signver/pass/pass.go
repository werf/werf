package pass

import (
	"io"
	"os"

	"github.com/sigstore/sigstore/pkg/cryptoutils"
	"golang.org/x/term"
)

const (
	SignVerPassword = "WERF_SIGNVER_PASSWORD"
)

// Read is for fuzzing
var Read = readPasswordFn

// readPasswordFn
// Copied from https://github.com/sigstore/cosign/blob/c948138c19691142c1e506e712b7c1646e8ceb21/cmd/cosign/cli/generate/generate_key_pair.go#L130
// and modified after.
func readPasswordFn(confirm bool) func() ([]byte, error) {
	if pw, ok := os.LookupEnv(SignVerPassword); ok {
		return func() ([]byte, error) {
			return []byte(pw), nil
		}
	}
	if term.IsTerminal(0) {
		return func() ([]byte, error) {
			return cryptoutils.GetPasswordFromStdIn(confirm)
		}
	}
	// Handle piped in passwords.
	return func() ([]byte, error) {
		return io.ReadAll(os.Stdin)
	}
}
