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
	// Command data
	Cmd *CmdData

	// Initialize git
	InitTrueGitWithOptions *InitTrueGitOptions
	// Initialize Docker Registry
	InitDockerRegistryWithOptions *InitDockerRegistryOptions

	// Initialize werf
	InitWerf bool
	// Initialize CommonGitDataManager
	InitGitDataManager bool
	// Initialize CommonManifestCache
	InitManifestCache bool
	// Initialize CommonLRUImagesCache
	InitLRUImagesCache bool
	// Initialize SSH agent. Should be used with defer ssh_agent.Terminate()
	InitSSHAgent bool

	// Setup OndemandKubeInitializer
	SetupOndemandKubeInitializer bool
}

type InitTrueGitOptions struct {
	Options true_git.Options
}

type InitDockerRegistryOptions struct {
	RegistryMirrors []string
}

func InitCommonComponents(ctx context.Context, opts InitCommonComponentsOptions) error {
	if opts.InitWerf {
		if err := werf.Init(*opts.Cmd.TmpDir, *opts.Cmd.HomeDir); err != nil {
			return fmt.Errorf("initialization error: %w", err)
		}
	}

	if opts.InitGitDataManager {
		gitDataManager, err := gitdata.GetHostGitDataManager(ctx)
		if err != nil {
			return fmt.Errorf("error getting host git data manager: %w", err)
		}

		if err := git_repo.Init(gitDataManager); err != nil {
			return fmt.Errorf("cant' init git repo: %w", err)
		}
	}

	if opts.InitManifestCache {
		if err := image.Init(); err != nil {
			return fmt.Errorf("manifest cache initialization error: %w", err)
		}
	}

	if opts.InitLRUImagesCache {
		if err := lrumeta.Init(); err != nil {
			return fmt.Errorf("lru cache initialization error: %w", err)
		}
	}

	if opts.InitTrueGitWithOptions != nil {
		if err := true_git.Init(ctx, opts.InitTrueGitWithOptions.Options); err != nil {
			return fmt.Errorf("git initialization error: %w", err)
		}
	}

	if opts.InitDockerRegistryWithOptions != nil {
		if err := DockerRegistryInit(ctx, opts.Cmd, opts.InitDockerRegistryWithOptions.RegistryMirrors); err != nil {
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
