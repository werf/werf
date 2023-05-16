package container_backend

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/werf"
)

type LegacyStageImage struct {
	*legacyBaseImage
	fromImage           *LegacyStageImage
	container           *LegacyStageImageContainer
	buildImage          *legacyBaseImage
	builtID             string
	commitChangeOptions LegacyCommitChangeOptions
	targetPlatform      string
}

func NewLegacyStageImage(fromImage *LegacyStageImage, name string, containerBackend ContainerBackend, targetPlatform string) *LegacyStageImage {
	stage := &LegacyStageImage{}
	stage.legacyBaseImage = newLegacyBaseImage(name, containerBackend)
	stage.fromImage = fromImage
	stage.container = newLegacyStageImageContainer(stage)
	stage.targetPlatform = targetPlatform
	return stage
}

func (i *LegacyStageImage) GetTargetPlatform() string {
	return i.targetPlatform
}

func (i *LegacyStageImage) GetCopy() LegacyImageInterface {
	ni := NewLegacyStageImage(i.fromImage, i.name, i.ContainerBackend, i.targetPlatform)
	if desc := i.GetStageDescription(); desc != nil {
		ni.SetStageDescription(desc.GetCopy())
	} else if info := i.GetInfo(); info != nil {
		ni.SetInfo(info.GetCopy())
	}
	return ni
}

func (i *LegacyStageImage) SetCommitChangeOptions(opts LegacyCommitChangeOptions) {
	i.commitChangeOptions = opts
}

func (i *LegacyStageImage) BuilderContainer() LegacyBuilderContainer {
	return &LegacyStageImageBuilderContainer{i}
}

func (i *LegacyStageImage) Container() LegacyContainer {
	return i.container
}

func (i *LegacyStageImage) GetID() string {
	if i.buildImage != nil {
		return i.buildImage.Name()
	} else {
		return i.legacyBaseImage.GetStageDescription().Info.ID
	}
}

func (i *LegacyStageImage) Build(ctx context.Context, options BuildOptions) error {
	if i.GetTargetPlatform() == i.ContainerBackend.GetDefaultPlatform() && i.ContainerBackend.GetDefaultPlatform() != "linux/amd64" {
		logboek.Context(ctx).Error().LogF("Detected your default build platform as %s.\n", i.ContainerBackend.GetDefaultPlatform())
		logboek.Context(ctx).Error().LogF("Building of stapel-type images using Docker-Server backend for platforms other than linux/amd64 is not supported.\n")
		logboek.Context(ctx).Error().LogF("Please either:\n * confirm emulation of linux/amd64 by exlicitly setting --platform=linux/amd64 param;\n * or use Dockerfile-type image instead.\n")
		logboek.Context(ctx).Error().LogLn()
		return fmt.Errorf("building of stapel image using Docker-Server backend is unsupported on your current platform %q", i.ContainerBackend.GetDefaultPlatform())
	}

	if i.GetTargetPlatform() != "" && i.GetTargetPlatform() != "linux/amd64" {
		logboek.Context(ctx).Error().LogF("Building of stapel-type images using Docker-Server backend for platforms other than linux/amd64 is not supported.\n")
		logboek.Context(ctx).Error().LogF("Please either:\n * use Buildah backend to build stapel-type images for arbitrary platforms;\n * or use Dockerfile-type images with any backend.\n")
		logboek.Context(ctx).Error().LogLn()
		return fmt.Errorf("building of stapel image using Docker-Server backend is unsupported for specified platform %q", i.GetTargetPlatform())
	}

	containerLockName := ContainerLockName(i.container.Name())
	if _, lock, err := werf.AcquireHostLock(ctx, containerLockName, lockgate.AcquireOptions{}); err != nil {
		return fmt.Errorf("failed to lock %s: %w", containerLockName, err)
	} else {
		defer werf.ReleaseHostLock(lock)
	}

	if debugDockerRunCommand() {
		runArgs, err := i.container.prepareRunArgs(ctx)
		if err != nil {
			return err
		}

		fmt.Printf("Docker run command:\ndocker run %s\n", strings.Join(runArgs, " "))

		if len(i.container.prepareAllRunCommands()) != 0 {
			fmt.Printf("Decoded command:\n%s\n", strings.Join(i.container.prepareAllRunCommands(), " && "))
		}
	}

	if containerRunErr := i.container.run(ctx); containerRunErr != nil {
		if strings.HasPrefix(containerRunErr.Error(), "container run failed") {
			if options.IntrospectBeforeError {
				logboek.Context(ctx).Default().LogFDetails("Launched command: %s\n", strings.Join(i.container.prepareAllRunCommands(), " && "))

				if err := logboek.Context(ctx).Streams().DoErrorWithoutProxyStreamDataFormatting(func() error {
					return i.introspectBefore(ctx)
				}); err != nil {
					return fmt.Errorf("introspect error failed: %w", err)
				}
			} else if options.IntrospectAfterError {
				if err := i.Commit(ctx); err != nil {
					return fmt.Errorf("introspect error failed: %w", err)
				}

				logboek.Context(ctx).Default().LogFDetails("Launched command: %s\n", strings.Join(i.container.prepareAllRunCommands(), " && "))

				if err := logboek.Context(ctx).Streams().DoErrorWithoutProxyStreamDataFormatting(func() error {
					return i.Introspect(ctx)
				}); err != nil {
					return fmt.Errorf("introspect error failed: %w", err)
				}
			}

			if err := i.container.rm(ctx); err != nil {
				return fmt.Errorf("unable to remove container (original error %s): %w", containerRunErr, err)
			}
		}

		return containerRunErr
	}

	if err := i.Commit(ctx); err != nil {
		return err
	}

	if err := i.container.rm(ctx); err != nil {
		return err
	}

	if info, err := i.ContainerBackend.GetImageInfo(ctx, i.MustGetBuiltID(), GetImageInfoOpts{}); err != nil {
		return err
	} else {
		i.SetInfo(info)
		i.SetStageDescription(&image.StageDescription{
			StageID: nil, // stage id does not available at the moment
			Info:    info,
		})
	}

	return nil
}

func (i *LegacyStageImage) Commit(ctx context.Context) error {
	builtId, err := i.container.commit(ctx)
	if err != nil {
		return err
	}

	i.buildImage = newLegacyBaseImage(builtId, i.ContainerBackend)
	i.builtID = builtId

	return nil
}

func (i *LegacyStageImage) Introspect(ctx context.Context) error {
	if err := i.container.introspect(ctx); err != nil {
		return err
	}

	return nil
}

func (i *LegacyStageImage) introspectBefore(ctx context.Context) error {
	if err := i.container.introspectBefore(ctx); err != nil {
		return err
	}

	return nil
}

func (i *LegacyStageImage) MustResetInfo(ctx context.Context) error {
	if i.buildImage != nil {
		return i.buildImage.MustResetInfo(ctx)
	} else {
		return i.legacyBaseImage.MustResetInfo(ctx)
	}
}

func (i *LegacyStageImage) GetInfo() *image.Info {
	if i.buildImage != nil {
		return i.buildImage.GetInfo()
	} else {
		return i.legacyBaseImage.GetInfo()
	}
}

func (i *LegacyStageImage) MustGetBuiltID() string {
	builtId := i.BuiltID()
	if builtId == "" {
		panic(fmt.Sprintf("image %s built id is not available", i.Name()))
	}
	return builtId
}

func (i *LegacyStageImage) SetBuiltID(builtID string) {
	i.builtID = builtID
}

func (i *LegacyStageImage) BuiltID() string {
	return i.builtID
}

func (i *LegacyStageImage) Tag(ctx context.Context, name string) error {
	_ = i.ContainerBackend.(*DockerServerBackend)
	return docker.CliTag(ctx, i.GetID(), name)
}

func (i *LegacyStageImage) Pull(ctx context.Context) error {
	_ = i.ContainerBackend.(*DockerServerBackend)

	var args []string
	if i.targetPlatform != "" {
		args = append(args, "--platform", i.targetPlatform)
	}
	args = append(args, i.name)

	if err := docker.CliPullWithRetries(ctx, args...); err != nil {
		return err
	}

	i.legacyBaseImage.UnsetInfo()

	return nil
}

func (i *LegacyStageImage) Push(ctx context.Context) error {
	_ = i.ContainerBackend.(*DockerServerBackend)

	return docker.CliPushWithRetries(ctx, i.name)
}

func debugDockerRunCommand() bool {
	return os.Getenv("WERF_DEBUG_DOCKER_RUN_COMMAND") == "1"
}
