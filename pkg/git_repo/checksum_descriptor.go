package git_repo

import (
	"fmt"
	"hash"
)

type ChecksumDescriptor struct {
	NoMatchPaths []string
	Hash         hash.Hash
}

func (c *ChecksumDescriptor) String() string {
	return fmt.Sprintf("%x", c.Hash.Sum(nil))
}

func (c *ChecksumDescriptor) GetNoMatchPaths() []string {
	return c.NoMatchPaths
}
