package git_repo

import (
	"context"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAI_BuildCloneOptions(t *testing.T) {
	t.Run("branch mapping clones single branch without tags", func(t *testing.T) {
		opts := buildCloneOptions("https://example.com/repo.git", "main")
		assert.True(t, opts.SingleBranch)
		assert.Equal(t, plumbing.NewBranchReferenceName("main"), opts.ReferenceName)
		assert.Equal(t, git.NoTags, opts.Tags)
	})

	t.Run("no branch clones everything", func(t *testing.T) {
		opts := buildCloneOptions("https://example.com/repo.git", "")
		assert.False(t, opts.SingleBranch)
		assert.Empty(t, opts.ReferenceName)
		assert.NotEqual(t, git.NoTags, opts.Tags)
	})
}

func TestAI_BuildFetchOptions(t *testing.T) {
	t.Run("branch mapping fetches without tags", func(t *testing.T) {
		opts := buildFetchOptions("origin", "main")
		assert.Equal(t, git.NoTags, opts.Tags)
		assert.True(t, opts.Force)
	})

	t.Run("no branch fetches all tags", func(t *testing.T) {
		opts := buildFetchOptions("origin", "")
		assert.Equal(t, git.AllTags, opts.Tags)
		assert.True(t, opts.Force)
	})
}

func TestAI_SyncLocalBranches(t *testing.T) {
	tmpDir := t.TempDir()
	rawRepo, err := git.PlainInit(tmpDir, true)
	require.NoError(t, err)

	hashA := plumbing.NewHash("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	hashB := plumbing.NewHash("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	hashC := plumbing.NewHash("cccccccccccccccccccccccccccccccccccccccc")

	require.NoError(t, rawRepo.Storer.SetReference(
		plumbing.NewHashReference("refs/remotes/origin/main", hashA),
	))
	require.NoError(t, rawRepo.Storer.SetReference(
		plumbing.NewHashReference("refs/remotes/origin/feature-branch", hashB),
	))
	require.NoError(t, rawRepo.Storer.SetReference(
		plumbing.NewHashReference("refs/remotes/origin/devops/includes-problem", hashC),
	))

	require.NoError(t, rawRepo.Storer.SetReference(
		plumbing.NewHashReference("refs/heads/main", hashA),
	))

	remote := &Remote{Base: NewBase("test", nil)}
	err = remote.syncLocalBranches(context.Background(), rawRepo)
	require.NoError(t, err)

	ref, err := rawRepo.Storer.Reference(plumbing.ReferenceName("refs/heads/main"))
	require.NoError(t, err)
	assert.Equal(t, hashA, ref.Hash())

	ref, err = rawRepo.Storer.Reference(plumbing.ReferenceName("refs/heads/feature-branch"))
	require.NoError(t, err)
	assert.Equal(t, hashB, ref.Hash())

	ref, err = rawRepo.Storer.Reference(plumbing.ReferenceName("refs/heads/devops/includes-problem"))
	require.NoError(t, err)
	assert.Equal(t, hashC, ref.Hash())
}
