package true_git

import (
	"context"
	"fmt"
	"strings"
)

const submoduleFileMode = "160000"

type ChangedPath struct {
	Path        string
	IsSubmodule bool
}

// ListChangedPaths returns repo-relative paths changed between two commits without generating
// patch content. Submodule entries are reported with the submodule path itself (no recursion).
func ListChangedPaths(ctx context.Context, gitDir, fromCommit, toCommit string) ([]ChangedPath, error) {
	diffCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: gitDir},
		"-c", "core.quotePath=false",
		"diff", "--raw", "-z", "--no-renames", fromCommit, toCommit,
	)
	if err := diffCmd.Run(ctx); err != nil {
		return nil, fmt.Errorf("git diff --raw command failed: %w", err)
	}

	return parseRawDiffPaths(diffCmd.OutBuf.String())
}

func parseRawDiffPaths(out string) ([]ChangedPath, error) {
	var result []ChangedPath

	tokens := strings.Split(out, "\x00")
	for i := 0; i+1 < len(tokens); i += 2 {
		meta, path := tokens[i], tokens[i+1]

		if !strings.HasPrefix(meta, ":") {
			return nil, fmt.Errorf("unexpected git raw diff entry %q", meta)
		}

		metaFields := strings.Fields(strings.TrimPrefix(meta, ":"))
		if len(metaFields) < 5 {
			return nil, fmt.Errorf("unexpected git raw diff entry %q", meta)
		}

		srcMode, dstMode := metaFields[0], metaFields[1]

		result = append(result, ChangedPath{
			Path:        path,
			IsSubmodule: srcMode == submoduleFileMode || dstMode == submoduleFileMode,
		})
	}

	return result, nil
}
