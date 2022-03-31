package true_git

import (
	"context"
	"fmt"

	"github.com/werf/logboek"
)

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
