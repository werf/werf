package true_git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"

	"github.com/werf/werf/pkg/util"
)

func UpwardLookupAndVerifyWorkTree(ctx context.Context, lookupPath string) (bool, string, error) {
	lookupPath = util.GetAbsoluteFilepath(lookupPath)

	for {
		dotGitPath := filepath.Join(lookupPath, git.GitDirName)

		if _, err := os.Stat(dotGitPath); os.IsNotExist(err) {
			if lookupPath != filepath.Dir(lookupPath) {
				lookupPath = filepath.Dir(lookupPath)
				continue
			}
		} else if err != nil {
			return false, "", fmt.Errorf("error accessing %q: %w", dotGitPath, err)
		} else if isValid, err := IsValidGitDir(ctx, dotGitPath); err != nil {
			return false, "", err
		} else if isValid {
			return true, lookupPath, nil
		}

		break
	}

	return false, "", nil
}

func IsValidWorkTree(ctx context.Context, workTree string) (bool, error) {
	return IsValidGitDir(ctx, filepath.Join(workTree, git.GitDirName))
}

func IsValidGitDir(ctx context.Context, gitDir string) (bool, error) {
	detectGitCmd := NewGitCmd(ctx, nil, "--git-dir", gitDir, "rev-parse")
	if err := detectGitCmd.Run(ctx); err != nil {
		if strings.HasPrefix(detectGitCmd.ErrBuf.String(), "fatal: not a git repository: ") {
			return false, nil
		}

		return false, fmt.Errorf("git rev parse command failed: %w", err)
	}

	return true, nil
}
