package git_repo

import (
	"encoding/hex"

	"github.com/go-git/go-git/v5/plumbing"
)

func newHash(s string) (plumbing.Hash, error) {
	var h plumbing.Hash

	b, err := hex.DecodeString(s)
	if err != nil {
		return h, err
	}

	copy(h[:], b)
	return h, nil
}
