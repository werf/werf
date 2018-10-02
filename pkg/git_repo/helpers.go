package git_repo

import (
	"encoding/hex"

	"gopkg.in/src-d/go-git.v4/plumbing"
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
