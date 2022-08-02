package true_git

import (
	"context"
	"fmt"
	"strings"
)

type RefDescriptor struct {
	Commit   string
	FullName string
	IsHEAD   bool

	IsTag   bool
	TagName string

	IsBranch   bool
	BranchName string
	IsRemote   bool
	RemoteName string
}

type ShowRefResult struct {
	Refs []RefDescriptor
}

func ShowRef(ctx context.Context, repoDir string) (*ShowRefResult, error) {
	headRefCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: repoDir}, "show-ref", "--head")
	if err := headRefCmd.Run(ctx); err != nil {
		return nil, fmt.Errorf("git get refs from HEAD command failed: %w", err)
	}

	res := &ShowRefResult{}

	outputLines := strings.Split(headRefCmd.OutBuf.String(), "\n")
	for _, line := range outputLines {
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}

		ref := RefDescriptor{
			Commit:   parts[0],
			FullName: parts[1],
		}

		switch {
		case ref.FullName == "HEAD":
			ref.IsHEAD = true
		case strings.HasPrefix(ref.FullName, "refs/tags/"):
			ref.IsTag = true
			ref.TagName = strings.TrimPrefix(ref.FullName, "refs/tags/")
		case strings.HasPrefix(ref.FullName, "refs/heads/"):
			ref.IsBranch = true
			ref.BranchName = strings.TrimPrefix(ref.FullName, "refs/heads/")
		case strings.HasPrefix(ref.FullName, "refs/remotes/"):
			ref.IsBranch = true
			ref.IsRemote = true
			parts := strings.SplitN(strings.TrimPrefix(ref.FullName, "refs/remotes/"), "/", 2)
			if len(parts) != 2 {
				continue
			}
			ref.RemoteName, ref.BranchName = parts[0], parts[1]
		}

		res.Refs = append(res.Refs, ref)
	}

	return res, nil
}
