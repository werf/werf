package signver

import (
	"github.com/sigstore/sigstore/pkg/cryptoutils"
)

// KeyOpts
// Copied from https://github.com/sigstore/cosign/blob/c948138c19691142c1e506e712b7c1646e8ceb21/cmd/cosign/cli/options/key.go#L20
// and modified after.
type KeyOpts struct {
	KeyRef   string
	PassFunc cryptoutils.PassFunc
}
