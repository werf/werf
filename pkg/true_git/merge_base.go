package true_git

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func IsAncestor(ctx context.Context, ancestorCommit, descendantCommit, gitDir string) (bool, error) {
	repository, err := PlainOpenWithOptions(gitDir, &PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return false, fmt.Errorf("open repo %q: %w", gitDir, err)
	}

	ancestorHash := plumbing.NewHash(ancestorCommit)
	descendantHash := plumbing.NewHash(descendantCommit)

	if ancestorHash == descendantHash {
		return true, nil
	}

	descendantObj, err := repository.CommitObject(descendantHash)
	if err != nil {
		return false, nil
	}

	if _, err := repository.CommitObject(ancestorHash); err != nil {
		return false, nil
	}

	visited := map[plumbing.Hash]bool{descendantHash: true}
	queue := []*object.Commit{descendantObj}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		parentIter := current.Parents()
		for {
			parent, err := parentIter.Next()
			if err != nil {
				break
			}
			if parent.Hash == ancestorHash {
				return true, nil
			}
			if !visited[parent.Hash] {
				visited[parent.Hash] = true
				queue = append(queue, parent)
			}
		}
	}

	return false, nil
}
