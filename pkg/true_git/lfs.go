package true_git

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
)

// Pointer files are below 1024 bytes by the Git LFS spec, so larger files are never read.
const lfsPointerMaxSize = 1024

var lfsPointerVersionPrefixes = [][]byte{
	[]byte("version https://git-lfs.github.com/spec/"),
	[]byte("version https://hawser.github.com/spec/"),
}

// pullLfsObjects fetches the LFS objects referenced by the checked out commit of workTreeDir and
// replaces the pointer files with their content. The filter config is passed explicitly because
// without it (`git lfs install` never run on the host) `git lfs pull` skips the checkout and
// exits 0; host-level lfs.fetchexclude is cleared for the same reason.
func pullLfsObjects(ctx context.Context, workTreeDir, pathScope string, env []string) error {
	args := []string{
		"-c", "filter.lfs.smudge=git-lfs smudge -- %f",
		"-c", "filter.lfs.process=git-lfs filter-process",
		"-c", "filter.lfs.clean=git-lfs clean -- %f",
		"-c", "filter.lfs.required=true",
		"lfs", "pull", "--exclude", "",
	}
	if pathScope != "" && pathScope != "." {
		args = append(args, "--include", pathScope)
	}

	pullCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: workTreeDir, Env: env}, args...)
	if err := pullCmd.Run(ctx); err != nil {
		if strings.Contains(pullCmd.ErrBuf.String(), "is not a git command") {
			return fmt.Errorf("git-lfs is not installed: %w", err)
		}
		return fmt.Errorf("git lfs pull command failed: %w", err)
	}

	return nil
}

func isLfsPointerFile(path string, size int64) (bool, error) {
	if size >= lfsPointerMaxSize {
		return false, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("read %q: %w", path, err)
	}

	return isLfsPointer(data), nil
}

func isLfsPointer(data []byte) bool {
	if len(data) >= lfsPointerMaxSize {
		return false
	}

	versionLine, rest, found := bytes.Cut(data, []byte("\n"))
	if !found {
		return false
	}

	hasVersionPrefix := false
	for _, prefix := range lfsPointerVersionPrefixes {
		if bytes.HasPrefix(versionLine, prefix) {
			hasVersionPrefix = true
			break
		}
	}
	if !hasVersionPrefix {
		return false
	}

	return bytes.HasPrefix(rest, []byte("oid sha256:")) || bytes.Contains(rest, []byte("\noid sha256:"))
}
