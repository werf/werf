package ls_tree

import (
	"context"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
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

func TestResultChecksum(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Result checksum")
}

var _ = ginkgo.DescribeTable("checksum includes full paths", func(oldPath, newPath string) {
	ctx := context.Background()
	before := newTestResultWithMode(filemode.Regular)
	before.lsTreeEntries[0].FullFilepath = oldPath
	after := newTestResultWithMode(filemode.Regular)
	after.lsTreeEntries[0].FullFilepath = newPath

	gomega.Expect(before.Checksum(ctx)).NotTo(gomega.Equal(after.Checksum(ctx)))
},
	ginkgo.Entry("file rename", "a.txt", "b.txt"),
	ginkgo.Entry("directory rename", "a/file.txt", "b/file.txt"),
	ginkgo.Entry("newline in filename", "a\n.txt", "b\n.txt"),
)

func TestResultChecksum_FileModeChangeFlipsChecksum(t *testing.T) {
	ctx := context.Background()

	regular := newTestResultWithMode(filemode.Regular).Checksum(ctx)
	executable := newTestResultWithMode(filemode.Executable).Checksum(ctx)

	require.NotEmpty(t, regular)
	require.NotEmpty(t, executable)
	require.NotEqual(t, regular, executable)
}

func TestResultChecksum_SameModeIsDeterministic(t *testing.T) {
	ctx := context.Background()

	first := newTestResultWithMode(filemode.Regular).Checksum(ctx)
	second := newTestResultWithMode(filemode.Regular).Checksum(ctx)

	require.Equal(t, first, second)
}
