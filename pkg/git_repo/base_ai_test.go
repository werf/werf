//go:build ai_tests

package git_repo

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/werf/werf/v2/pkg/true_git"
)

const envKey = "WERF_GIT_USE_WORKTREE"

func TestAI_UseGitWorktree_Default(t *testing.T) {
	reset := withEnvReset(t, envKey)
	defer reset()
	assert.False(t, useGitWorktree())
}

func TestAI_UseGitWorktree_SetToOne(t *testing.T) {
	reset := withEnvReset(t, envKey)
	defer reset()
	assert.NoError(t, os.Setenv(envKey, "1"))
	assert.True(t, useGitWorktree())
}

func TestAI_UseGitWorktree_SetToOther(t *testing.T) {
	reset := withEnvReset(t, envKey)
	defer reset()
	assert.NoError(t, os.Setenv(envKey, "foo"))
	assert.False(t, useGitWorktree())
}

func TestAI_FormatSubmoduleValidationError(t *testing.T) {
	result := &true_git.SubmoduleValidationResult{
		Valid: false,
		Errors: []true_git.SubmoduleValidationError{
			{
				SubmodulePath: "modules/a",
				Message:       "submodule \"modules/a\" is not initialized. Run: git submodule update --init --recursive",
			},
			{
				SubmodulePath: "modules/b",
				Message:       "submodule \"modules/b\": expected commit abc, got def. Run: git submodule update --init --recursive",
			},
		},
	}

	err := formatSubmoduleValidationError(result)
	assert.Equal(t, "submodule state validation failed, run 'git submodule update --init --recursive' to fix:\nsubmodule \"modules/a\": submodule \"modules/a\" is not initialized. Run: git submodule update --init --recursive\nsubmodule \"modules/b\": submodule \"modules/b\": expected commit abc, got def. Run: git submodule update --init --recursive", err.Error())
}

func withEnvReset(t *testing.T, key string) func() {
	original, ok := os.LookupEnv(key)
	if ok {
		assert.NoError(t, os.Unsetenv(key))
	}
	return func() {
		if ok {
			assert.NoError(t, os.Setenv(key, original))
		} else {
			assert.NoError(t, os.Unsetenv(key))
		}
	}
}
