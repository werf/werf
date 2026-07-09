package container_backend

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/frontend/dockerui"
	dockerspec "github.com/moby/docker-image-spec/specs-go/v1"
	"github.com/tonistiigi/fsutil"

	"github.com/werf/werf/v2/pkg/buildkit"
)

func (backend *BuildkitBackend) BuildDockerfileStage(ctx context.Context, baseImage string, opts BuildDockerfileStageOptions, instructions ...InstructionInterface) (string, error) {
	repo, err := backend.getStagesStorageRepo()
	if err != nil {
		return "", err
	}

	cl, err := backend.getClient(ctx)
	if err != nil {
		return "", err
	}

	platform, err := backend.parsePlatform(opts.TargetPlatform)
	if err != nil {
		return "", err
	}

	resolver := buildkit.NewImageMetaResolver(platform)
	pinnedBaseRef, baseConfig, err := resolver.ResolvePinnedRef(ctx, baseImage, platform)
	if err != nil {
		return "", fmt.Errorf("resolve base image %q: %w", baseImage, err)
	}

	img := &dockerspec.DockerOCIImage{}
	if err := json.Unmarshal(baseConfig, img); err != nil {
		return "", fmt.Errorf("unmarshal base image %q config: %w", baseImage, err)
	}

	state := llb.Image(pinnedBaseRef, llb.Platform(*platform))
	state, err = state.WithImageConfig(baseConfig)
	if err != nil {
		return "", fmt.Errorf("apply base image %q config to llb state: %w", baseImage, err)
	}

	stage := &buildkit.DockerfileStageState{
		State:    state,
		Image:    img,
		Platform: *platform,
	}

	for _, instr := range instructions {
		if err := instr.ApplyBuildkit(ctx, stage); err != nil {
			return "", fmt.Errorf("unable to apply instruction %s: %w", instr.Name(), err)
		}
	}

	def, err := stage.State.Marshal(ctx)
	if err != nil {
		return "", fmt.Errorf("marshal llb state: %w", err)
	}

	imageConfig, err := json.Marshal(stage.Image)
	if err != nil {
		return "", fmt.Errorf("marshal image config: %w", err)
	}

	attachables, err := backend.getSessionAttachables(stage.SSH, stage.Secrets)
	if err != nil {
		return "", err
	}

	localMounts := map[string]fsutil.FS{}
	if stage.UsesContext {
		if opts.BuildContextArchive == nil {
			panic(fmt.Sprintf("BuildContextArchive can't be nil: %+v", opts))
		}
		contextDir, err := opts.BuildContextArchive.ExtractOrGetExtractedDir(ctx)
		if err != nil {
			return "", fmt.Errorf("unable to extract build context: %w", err)
		}
		contextFS, err := fsutil.NewFS(contextDir)
		if err != nil {
			return "", fmt.Errorf("create fs for build context %q: %w", contextDir, err)
		}
		localMounts[dockerui.DefaultLocalNameContext] = contextFS
	}

	builtID, err := buildkit.Solve(ctx, cl, def, buildkit.SolveOptions{
		Repo:        repo,
		ImageConfig: imageConfig,
		LocalMounts: localMounts,
		Session:     attachables,
	})
	if err != nil {
		return "", fmt.Errorf("build dockerfile stage: %w", err)
	}

	return builtID, nil
}
