package common

import (
	"context"
	"fmt"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/buildah"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker"
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
	buildahMode      buildah.Mode
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

	if opts.InitProcessContainerBackend || opts.InitDockerRegistry {
		newCtx, err := cmanager.InitContainerBackendComponents(ctx, opts.Cmd, opts.InitDockerRegistry, opts.InitProcessContainerBackend)
		if err != nil {
			return nil, ctx, err
		}
		ctx = newCtx
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

	return cmanager, ctx, nil
}

// InitContainerBackendComponents initializes buildah mode, docker config, registry mirrors,
// docker registry and/or container backend, storing results on m. Safe to call multiple times;
// each initXxx section only runs the parts requested by initDockerRegistry/initProcessContainerBackend.
func (m *ComponentsManager) InitContainerBackendComponents(ctx context.Context, cmd *CmdData, initDockerRegistry, initProcessContainerBackend bool) (context.Context, error) {
	buildahMode, _, err := GetBuildahMode()
	if err != nil {
		return ctx, fmt.Errorf("unable to determine buildah mode: %w", err)
	}
	m.buildahMode = *buildahMode

	// Set DOCKER_CONFIG early so that authn.DefaultKeychain (used by go-containerregistry)
	// picks up custom credentials even when the full container backend is not initialized.
	if err := docker.InitDockerConfig(docker.InitOptions{DockerConfigDir: *cmd.DockerConfig}); err != nil {
		return ctx, fmt.Errorf("init docker config: %w", err)
	}

	if initProcessContainerBackend && m.buildahMode == buildah.ModeDisabled {
		newCtx, err := InitProcessDocker(ctx, cmd)
		if err != nil {
			return ctx, fmt.Errorf("unable to init docker: %w", err)
		}
		ctx = newCtx
	}

	rm, err := GetContainerRegistryMirror(ctx, cmd, m.buildahMode)
	if err != nil {
		return ctx, fmt.Errorf("error get container registry mirrors: %w", err)
	}
	m.registryMirrors = &rm

	if initDockerRegistry {
		if err := DockerRegistryInit(ctx, cmd, *m.registryMirrors, m.buildahMode); err != nil {
			return ctx, fmt.Errorf("docker registry initialization error: %w", err)
		}
	}

	if initProcessContainerBackend {
		cb, newCtx, err := InitProcessContainerBackend(ctx, cmd, *m.registryMirrors)
		if err != nil {
			return ctx, fmt.Errorf("container backend initialization error: %w", err)
		}
		m.containerBackend = cb
		ctx = newCtx
	}

	return ctx, nil
}

// EnsureContainerBackend lazily initializes the container backend (and, optionally, the docker
// registry) on first call and reuses it on subsequent calls. Intended for commands where the
// backend is only required conditionally (e.g. project has images to build).
func (m *ComponentsManager) EnsureContainerBackend(ctx context.Context, cmd *CmdData, initDockerRegistry bool) (container_backend.ContainerBackend, context.Context, error) {
	if m.containerBackend != nil {
		return m.containerBackend, ctx, nil
	}

	newCtx, err := m.InitContainerBackendComponents(ctx, cmd, initDockerRegistry, true)
	if err != nil {
		return nil, ctx, err
	}

	return m.containerBackend, newCtx, nil
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

// TryContainerBackend returns the container backend and true if it has been initialized,
// or nil and false otherwise. Use this instead of ContainerBackend() when backend
// initialization is optional.
func (m *ComponentsManager) TryContainerBackend() (container_backend.ContainerBackend, bool) {
	return m.containerBackend, m.containerBackend != nil
}

func (m *ComponentsManager) BuildahMode() buildah.Mode {
	return m.buildahMode
}

func (m *ComponentsManager) TerminateSSHAgent() {
	err := ssh_agent.Terminate()
	if err != nil {
		logboek.Warn().LogF("WARNING: ssh agent termination failed: %s\n", err)
	}
}
