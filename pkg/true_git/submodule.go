package true_git

import (
	"context"
	"fmt"
	"sync"

	"github.com/werf/logboek"
)

var (
	submoduleInitOnceMap sync.Map // map[string]*sync.Once,
)

func UpdateSubmodulesOnce(ctx context.Context, repoDir string) error {
	onceIface, _ := submoduleInitOnceMap.LoadOrStore(repoDir, new(sync.Once))
	once := onceIface.(*sync.Once)

	var err error
	once.Do(func() {
		if err := syncSubmodules(ctx, repoDir, repoDir); err != nil {
			err = fmt.Errorf("failed to sync submodules: %w", err)
			return
		}

		if err := updateSubmodules(ctx, repoDir, repoDir); err != nil {
			err = fmt.Errorf("failed to update submodules: %w", err)
			return
		}
	})
	return err
}

func syncSubmodules(ctx context.Context, repoDir, workTreeDir string) error {
	logProcessMsg := fmt.Sprintf("Sync submodules in work tree %q", workTreeDir)
	return logboek.Context(ctx).Info().LogProcess(logProcessMsg).DoError(func() error {
		submSyncCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: workTreeDir}, "submodule", "sync", "--recursive")
		if err := submSyncCmd.Run(ctx); err != nil {
			return fmt.Errorf("submodule sync command failed: %w", err)
		}

		return nil
	})
}

func updateSubmodules(ctx context.Context, repoDir, workTreeDir string) error {
	logProcessMsg := fmt.Sprintf("Update submodules in work tree %q", workTreeDir)
	return logboek.Context(ctx).Info().LogProcess(logProcessMsg).DoError(func() error {
		submUpdateCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: workTreeDir}, "submodule", "update", "--checkout", "--force", "--init", "--recursive")
		if err := submUpdateCmd.Run(ctx); err != nil {
			return fmt.Errorf("submodule update command failed: %w", err)
		}

		return nil
	})
}
