package true_git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func getCommonGitOptions() []string {
	return []string{"-c", "core.autocrlf=false", "-c", "gc.auto=0", "-c", "commit.gpgsign=false", "-c", "core.untrackedCache=false", "-c", "core.splitIndex=false"}
}

func getIncludePathOptions(ctx context.Context, repoDir string) ([]string, error) {
	configCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: repoDir}, "config", "--show-origin", "--get-all", "include.path")
	if err := configCmd.Run(ctx); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 && configCmd.OutBuf.String() == "" && configCmd.ErrBuf.String() == "" {
			return nil, nil
		}
		return nil, fmt.Errorf("get git include.path from %s: %w", repoDir, err)
	}

	var opts []string
	for _, line := range strings.Split(configCmd.OutBuf.String(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		origin, includePath, found := strings.Cut(line, "\t")
		if !found {
			continue
		}

		// include.path passed via -c has no declaring config file, so relative values are resolved against their origin up front
		if !strings.HasPrefix(includePath, "~") && !filepath.IsAbs(includePath) {
			if originFile, ok := strings.CutPrefix(origin, "file:"); ok {
				if !filepath.IsAbs(originFile) {
					originFile = filepath.Join(repoDir, originFile)
				}
				includePath = filepath.Join(filepath.Dir(originFile), includePath)
			} else {
				absIncludePath, err := filepath.Abs(includePath)
				if err != nil {
					return nil, fmt.Errorf("resolve include.path %q: %w", includePath, err)
				}
				includePath = absIncludePath
			}
		}

		opts = append(opts, "-c", "include.path="+includePath)
	}

	return opts, nil
}

func debug() bool {
	return os.Getenv("WERF_DEBUG_TRUE_GIT") == "1"
}
