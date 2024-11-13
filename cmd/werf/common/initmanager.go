package common

import (
	"context"
	"fmt"

	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/git_repo/gitdata"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/ssh_agent"
	"github.com/werf/werf/v2/pkg/storage/lrumeta"
	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/werf"
)

type InitCommonComponentsOptions struct {
	Cmd *CmdData

	InitTrueGit        InitTrueGitOptions
	InitDockerRegistry InitDockerRegistryOptions

	InitWerf     bool
	InitGitRepo  bool
	InitImage    bool
	InitLRUMeta  bool
	InitSSHAgent bool

	SetupOndemandKubeInitializer bool
}

type InitTrueGitOptions struct {
	Init    bool
	Options true_git.Options
}

type InitDockerRegistryOptions struct {
	Init            bool
	RegistryMirrors []string
}

func InitCommonComponents(ctx context.Context, opts InitCommonComponentsOptions) error {
	if opts.InitWerf {
		if err := werf.Init(*opts.Cmd.TmpDir, *opts.Cmd.HomeDir); err != nil {
			return fmt.Errorf("initialization error: %w", err)
		}
	}

	if opts.InitGitRepo {
		gitDataManager, err := gitdata.GetHostGitDataManager(ctx)
		if err != nil {
			return fmt.Errorf("error getting host git data manager: %w", err)
		}

		if err := git_repo.Init(gitDataManager); err != nil {
			return fmt.Errorf("cant' init git repo: %w", err)
		}
	}

	if opts.InitImage {
		if err := image.Init(); err != nil {
			return fmt.Errorf("manifest cache initialization error: %w", err)
		}
	}

	if opts.InitLRUMeta {
		if err := lrumeta.Init(); err != nil {
			return fmt.Errorf("lru cache initialization error: %w", err)
		}
	}

	if opts.InitTrueGit.Init {
		if err := true_git.Init(ctx, opts.InitTrueGit.Options); err != nil {
			return fmt.Errorf("git initialization error: %w", err)
		}
	}

	if opts.InitDockerRegistry.Init {
		if err := DockerRegistryInit(ctx, opts.Cmd, opts.InitDockerRegistry.RegistryMirrors); err != nil {
			return fmt.Errorf("docker registry initialization error: %w", err)
		}
	}

	if opts.InitSSHAgent {
		if err := ssh_agent.Init(ctx, GetSSHKey(opts.Cmd)); err != nil {
			return fmt.Errorf("cannot initialize ssh agent: %w", err)
		}
	}

	if opts.SetupOndemandKubeInitializer {
		SetupOndemandKubeInitializer(*opts.Cmd.KubeContext, *opts.Cmd.KubeConfig, *opts.Cmd.KubeConfigBase64, *opts.Cmd.KubeConfigPathMergeList)
		if err := GetOndemandKubeInitializer().Init(ctx); err != nil {
			return err
		}
	}

	return nil
}
