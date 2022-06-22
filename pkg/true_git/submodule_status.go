package true_git

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/werf/werf/pkg/git_repo/repo_handle"
)

type SubmodulesStatus []*SubmoduleStatus

type SubmoduleStatus struct {
	GitDir string `json:"-"`

	Path       string           `json:"path"`
	Commit     string           `json:"commit"`
	Submodules SubmodulesStatus `json:"submodules"`
}

func GetSubmodulesStatus(ctx context.Context, commit, gitDir, workTreeCacheDir string) (status SubmodulesStatus, resErr error) {
	resErr = withWorkTreeCacheLock(ctx, workTreeCacheDir, func() error {
		var err error
		status, err = getSubmodulesStatus(ctx, commit, gitDir, workTreeCacheDir)
		return err
	})

	return
}

func getSubmodulesStatus(ctx context.Context, commit, gitDir, workTreeCacheDir string) (SubmodulesStatus, error) {
	var err error

	gitDir, err = filepath.Abs(gitDir)
	if err != nil {
		return nil, fmt.Errorf("bad git dir %s: %w", gitDir, err)
	}

	workTreeCacheDir, err = filepath.Abs(workTreeCacheDir)
	if err != nil {
		return nil, fmt.Errorf("bad work tree cache dir %s: %w", workTreeCacheDir, err)
	}

	workTreeDir, err := prepareWorkTree(ctx, gitDir, workTreeCacheDir, commit, true)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare work tree in cache %s for commit %s: %w", workTreeCacheDir, commit, err)
	}

	repository, err := GitOpenWithCustomWorktreeDir(gitDir, workTreeDir)
	if err != nil {
		return nil, fmt.Errorf("git open failed: %w", err)
	}

	repoHandle, err := repo_handle.NewHandle(repository)
	if err != nil {
		return nil, fmt.Errorf("unable to create repo handle: %w", err)
	}

	for _, sub := range repoHandle.Submodules() {
		fmt.Printf("Name=%q Path=%q Url=%q Branch=%q\n", sub.Config().Name, sub.Config().Path, sub.Config().URL, sub.Config().Branch)
		fmt.Printf("Status Path=%q Current=%q Expected=%q Branch=%q\n", sub.Status().Path, sub.Status().Current, sub.Status().Expected, sub.Status().Branch)
	}

	return nil, nil
}
