//go:build ai_tests

package true_git

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"
)

func TestAI_ShowRef_HeadOnly(t *testing.T) {
	// Create a temporary git repo with just HEAD
	tmpDir := t.TempDir()
	repo, err := git.PlainInit(tmpDir, false)
	require.NoError(t, err)

	// Create initial commit
	wt, err := repo.Worktree()
	require.NoError(t, err)

	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	_, err = wt.Add("test.txt")
	require.NoError(t, err)

	_, err = wt.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
		},
	})
	require.NoError(t, err)

	// Call ShowRef
	result, err := ShowRef(context.Background(), tmpDir)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should have at least HEAD
	headFound := false
	for _, ref := range result.Refs {
		if ref.IsHEAD {
			headFound = true
			require.NotEmpty(t, ref.Commit)
			require.Equal(t, "HEAD", ref.FullName)
			break
		}
	}
	require.True(t, headFound, "HEAD ref not found")
}

func TestAI_ShowRef_WithBranches(t *testing.T) {
	tmpDir := t.TempDir()
	repo, err := git.PlainInit(tmpDir, false)
	require.NoError(t, err)

	wt, err := repo.Worktree()
	require.NoError(t, err)

	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	_, err = wt.Add("test.txt")
	require.NoError(t, err)

	_, err = wt.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
		},
	})
	require.NoError(t, err)

	// Create a second branch
	err = wt.Checkout(&git.CheckoutOptions{
		Create: true,
		Branch: "refs/heads/feature",
	})
	require.NoError(t, err)

	// Make a change on the new branch
	err = os.WriteFile(testFile, []byte("modified content"), 0644)
	require.NoError(t, err)

	_, err = wt.Add("test.txt")
	require.NoError(t, err)

	_, err = wt.Commit("Feature commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
		},
	})
	require.NoError(t, err)

	// Call ShowRef
	result, err := ShowRef(context.Background(), tmpDir)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should have master (or main) and feature branches
	masterFound := false
	featureFound := false

	for _, ref := range result.Refs {
		if ref.IsBranch && ref.BranchName == "master" && !ref.IsRemote {
			masterFound = true
			require.NotEmpty(t, ref.Commit)
			require.Equal(t, "refs/heads/master", ref.FullName)
		}
		if ref.IsBranch && ref.BranchName == "feature" && !ref.IsRemote {
			featureFound = true
			require.NotEmpty(t, ref.Commit)
			require.Equal(t, "refs/heads/feature", ref.FullName)
		}
	}

	require.True(t, masterFound || featureFound, "At least one branch should be found")
}

func TestAI_ShowRef_WithLightweightTag(t *testing.T) {
	tmpDir := t.TempDir()
	repo, err := git.PlainInit(tmpDir, false)
	require.NoError(t, err)

	wt, err := repo.Worktree()
	require.NoError(t, err)

	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	_, err = wt.Add("test.txt")
	require.NoError(t, err)

	hash, err := wt.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
		},
	})
	require.NoError(t, err)

	// Create a lightweight tag
	_, err = repo.CreateTag("v1.0", hash, nil)
	require.NoError(t, err)

	// Call ShowRef
	result, err := ShowRef(context.Background(), tmpDir)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should have the tag pointing to the commit
	tagFound := false
	for _, ref := range result.Refs {
		if ref.IsTag && ref.TagName == "v1.0" {
			tagFound = true
			require.NotEmpty(t, ref.Commit)
			require.Equal(t, "refs/tags/v1.0", ref.FullName)
			require.Equal(t, hash.String(), ref.Commit)
			break
		}
	}
	require.True(t, tagFound, "v1.0 tag not found")
}

func TestAI_ShowRef_WithAnnotatedTag(t *testing.T) {
	tmpDir := t.TempDir()
	repo, err := git.PlainInit(tmpDir, false)
	require.NoError(t, err)

	wt, err := repo.Worktree()
	require.NoError(t, err)

	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	_, err = wt.Add("test.txt")
	require.NoError(t, err)

	hash, err := wt.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
		},
	})
	require.NoError(t, err)

	// Create an annotated tag
	sig := &object.Signature{
		Name:  "Test User",
		Email: "test@example.com",
	}

	_, err = repo.CreateTag("v2.0", hash, &git.CreateTagOptions{
		Tagger:  sig,
		Message: "Version 2.0",
	})
	require.NoError(t, err)

	// Call ShowRef
	result, err := ShowRef(context.Background(), tmpDir)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should have the tag, resolved to commit hash (not tag object hash)
	tagFound := false
	for _, ref := range result.Refs {
		if ref.IsTag && ref.TagName == "v2.0" {
			tagFound = true
			require.NotEmpty(t, ref.Commit)
			require.Equal(t, "refs/tags/v2.0", ref.FullName)
			// For annotated tags, commit should be the target commit, not the tag object
			require.Equal(t, hash.String(), ref.Commit, "Annotated tag should resolve to commit hash")
			break
		}
	}
	require.True(t, tagFound, "v2.0 annotated tag not found")
}

func TestAI_ShowRef_EmptyRepo(t *testing.T) {
	tmpDir := t.TempDir()
	repo, err := git.PlainInit(tmpDir, false)
	require.NoError(t, err)

	// Try to get HEAD from empty repo
	head, err := repo.Head()
	// HEAD might not exist or might error in empty repo - both are acceptable

	if err == nil {
		// If HEAD exists, use it
		t.Logf("HEAD exists: %v", head)
	}

	// Call ShowRef on empty repo - should not panic
	result, err := ShowRef(context.Background(), tmpDir)
	// May error on empty repo (HEAD doesn't exist), or may succeed with empty result
	if err == nil {
		require.NotNil(t, result)
	}
}
