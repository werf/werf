package git_repo

import (
	"context"
	"fmt"
)

func GetVirtualMergeParents(ctx context.Context, gitRepo GitRepo, virtualMergeCommit string) (string, string, error) {
	if parents, err := gitRepo.GetMergeCommitParents(ctx, virtualMergeCommit); err != nil {
		return "", "", err
	} else if len(parents) == 2 {
		return parents[1], parents[0], nil
	} else {
		return "", "", fmt.Errorf("got unexpected parents: %v", parents)
	}
}
