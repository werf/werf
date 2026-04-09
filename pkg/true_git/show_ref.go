package true_git

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
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
	repository, err := PlainOpenWithOptions(repoDir, &PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return nil, fmt.Errorf("open repository: %w", err)
	}

	res := &ShowRefResult{}

	head, err := repository.Head()
	if err == nil {
		res.Refs = append(res.Refs, RefDescriptor{
			IsHEAD:   true,
			Commit:   head.Hash().String(),
			FullName: "HEAD",
		})
	}

	refIter, err := repository.References()
	if err != nil {
		return nil, fmt.Errorf("iterate references: %w", err)
	}

	err = refIter.ForEach(func(ref *plumbing.Reference) error {
		commit := ref.Hash().String()
		fullName := ref.Name().String()

		if ref.Name().IsTag() {
			if tag, err := repository.TagObject(ref.Hash()); err == nil {
				commit = tag.Target.String()
			}
		}

		refDesc := RefDescriptor{
			Commit:   commit,
			FullName: fullName,
		}

		switch {
		case fullName == "HEAD":
			refDesc.IsHEAD = true
		case strings.HasPrefix(fullName, "refs/tags/"):
			refDesc.IsTag = true
			refDesc.TagName = strings.TrimPrefix(fullName, "refs/tags/")
		case strings.HasPrefix(fullName, "refs/heads/"):
			refDesc.IsBranch = true
			refDesc.BranchName = strings.TrimPrefix(fullName, "refs/heads/")
		case strings.HasPrefix(fullName, "refs/remotes/"):
			refDesc.IsBranch = true
			refDesc.IsRemote = true
			parts := strings.SplitN(strings.TrimPrefix(fullName, "refs/remotes/"), "/", 2)
			if len(parts) == 2 {
				refDesc.RemoteName = parts[0]
				refDesc.BranchName = parts[1]
			}
		}

		res.Refs = append(res.Refs, refDesc)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("process references: %w", err)
	}

	return res, nil
}
