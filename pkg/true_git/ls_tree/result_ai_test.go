package ls_tree

import (
	"context"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"
)

func newTestResultWithMode(mode filemode.FileMode) *Result {
	hash := plumbing.NewHash("e69de29bb2d1d6434b8b29ae775ad8c2e48c5391")
	return NewResult("commit", "", []*LsTreeEntry{
		{
			FullFilepath: "app/main.sh",
			TreeEntry: object.TreeEntry{
				Name: "main.sh",
				Mode: mode,
				Hash: hash,
			},
		},
	}, []*SubmoduleResult{})
}

func TestAI_ResultChecksum_FileModeChangeFlipsChecksum(t *testing.T) {
	ctx := context.Background()

	regular := newTestResultWithMode(filemode.Regular).Checksum(ctx)
	executable := newTestResultWithMode(filemode.Executable).Checksum(ctx)

	require.NotEmpty(t, regular)
	require.NotEmpty(t, executable)
	require.NotEqual(t, regular, executable)
}

func TestAI_ResultChecksum_SameModeIsDeterministic(t *testing.T) {
	ctx := context.Background()

	first := newTestResultWithMode(filemode.Regular).Checksum(ctx)
	second := newTestResultWithMode(filemode.Regular).Checksum(ctx)

	require.Equal(t, first, second)
}
