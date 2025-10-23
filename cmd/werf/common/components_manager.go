package common

import (
	"context"
	"fmt"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/git_repo/gitdata"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/logging"
	"github.com/werf/werf/v2/pkg/ssh_agent"
	"github.com/werf/werf/v2/pkg/storage/lrumeta"
	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

type ComponentsManager struct {
	registryMirrors  *[]string
	containerBackend container_backend.ContainerBackend
}

type InitCommonComponentsOptions struct {
	// Command data
	Cmd *CmdData

	// Initialize git
	InitTrueGitWithOptions *InitTrueGitOptions

	// Initialize Docker Registry
	InitDockerRegistry bool
	// Initialize Container Backend
	InitProcessContainerBackend bool
	// Initialize werf
	InitWerf bool
	// Initialize CommonGitDataManager
	InitGitDataManager bool
	// Initialize CommonManifestCache
	InitManifestCache bool
	// Initialize CommonLRUImagesCache
	InitLRUImagesCache bool
	// Initialize SSH agent. Should be used with defer call TerminateSSHAgent()
	InitSSHAgent bool

	// Setup OndemandKubeInitializer
	SetupOndemandKubeInitializer bool
}

type InitTrueGitOptions struct {
	Options true_git.Options
}

func InitCommonComponents(ctx context.Context, opts InitCommonComponentsOptions) (*ComponentsManager, context.Context, error) {
	if err := opts.Cmd.ProcessFlags(); err != nil {
		return nil, ctx, fmt.Errorf("process flags: %w", err)
	}

	cmanager := &ComponentsManager{}
	if opts.InitWerf {
		if err := werf.Init(*opts.Cmd.TmpDir, *opts.Cmd.HomeDir); err != nil {
			return nil, ctx, fmt.Errorf("initialization error: %w", err)
		}

		if ok, warning, err := logging.BackgroundWarning(werf.GetServiceDir()); err != nil {
			return nil, ctx, err
		} else if ok {
			global_warnings.GlobalWarningLn(ctx, warning)
		}
	}

	if opts.InitDockerRegistry || opts.InitProcessContainerBackend {
		rm, err := GetContainerRegistryMirror(ctx, opts.Cmd)
		if err != nil {
			return nil, ctx, fmt.Errorf("error get container registry mirrors: %w", err)
		}
		cmanager.registryMirrors = &rm
	}

	if opts.InitDockerRegistry {
		if err := DockerRegistryInit(ctx, opts.Cmd, *cmanager.registryMirrors); err != nil {
			return nil, ctx, fmt.Errorf("docker registry initialization error: %w", err)
		}
	}

	if opts.InitProcessContainerBackend {
		cb, newCtx, err := InitProcessContainerBackend(ctx, opts.Cmd, *cmanager.registryMirrors)
		if err != nil {
			return nil, ctx, fmt.Errorf("container backend initialization error: %w", err)
		}
		cmanager.containerBackend = cb
		ctx = newCtx // context reinitialization
	}

	if opts.InitGitDataManager {
		gitDataManager, err := gitdata.GetHostGitDataManager(ctx)
		if err != nil {
			return nil, ctx, fmt.Errorf("error getting host git data manager: %w", err)
		}

		if err := git_repo.Init(gitDataManager); err != nil {
			return nil, ctx, fmt.Errorf("cannot initialize git repo: %w", err)
		}
	}

	if opts.InitManifestCache {
		if err := image.Init(); err != nil {
			return nil, ctx, fmt.Errorf("manifest cache initialization error: %w", err)
		}
	}

	if opts.InitLRUImagesCache {
		if err := lrumeta.Init(); err != nil {
			return nil, ctx, fmt.Errorf("lru cache initialization error: %w", err)
		}
	}

	if opts.InitTrueGitWithOptions != nil {
		if err := true_git.Init(ctx, opts.InitTrueGitWithOptions.Options); err != nil {
			return nil, ctx, fmt.Errorf("git initialization error: %w", err)
		}
	}

	if opts.InitSSHAgent {
		if err := ssh_agent.Init(ctx, GetSSHKey(opts.Cmd)); err != nil {
			return nil, ctx, fmt.Errorf("cannot initialize ssh agent: %w", err)
		}
	}

	if opts.SetupOndemandKubeInitializer {
		SetupOndemandKubeInitializer(opts.Cmd.KubeContextCurrent, opts.Cmd.LegacyKubeConfigPath, opts.Cmd.KubeConfigBase64, opts.Cmd.LegacyKubeConfigPathsMergeList)
		if err := GetOndemandKubeInitializer().Init(ctx); err != nil {
			return nil, ctx, err
		}
	}

	return cmanager, ctx, nil
}

func (m *ComponentsManager) RegistryMirrors() []string {
	if m.registryMirrors == nil {
		panic("bug: init required!")
	}
	return *m.registryMirrors
}

func (m *ComponentsManager) ContainerBackend() container_backend.ContainerBackend {
	if m.containerBackend == nil {
		panic("bug: init required!")
	}
	return m.containerBackend
}

func (m *ComponentsManager) TerminateSSHAgent() {
	err := ssh_agent.Terminate()
	if err != nil {
		logboek.Warn().LogF("WARNING: ssh agent termination failed: %s\n", err)
	}
}
