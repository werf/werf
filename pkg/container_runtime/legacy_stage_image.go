package container_runtime

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/werf/lockgate"
	"github.com/werf/werf/pkg/werf"

	"github.com/docker/docker/api/types"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/image"
)

type LegacyStageImage struct {
	*legacyBaseImage
	fromImage              *LegacyStageImage
	container              *LegacyStageImageContainer
	buildImage             *legacyBaseImage
	dockerfileImageBuilder *DockerfileImageBuilder
}

func NewLegacyStageImage(fromImage *LegacyStageImage, name string, localDockerServerRuntime *DockerServerRuntime) *LegacyStageImage {
	stage := &LegacyStageImage{}
	stage.legacyBaseImage = newLegacyBaseImage(name, localDockerServerRuntime)
	stage.fromImage = fromImage
	stage.container = newLegacyStageImageContainer(stage)
	return stage
}

func (i *LegacyStageImage) Inspect() *types.ImageInspect {
	return i.inspect
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

func (i *LegacyStageImage) Build(ctx context.Context, options LegacyBuildOptions) error {
	if i.dockerfileImageBuilder != nil {
		if err := i.dockerfileImageBuilder.Build(ctx); err != nil {
			return err
		}
	} else {
		containerLockName := ContainerLockName(i.container.Name())
		if _, lock, err := werf.AcquireHostLock(ctx, containerLockName, lockgate.AcquireOptions{}); err != nil {
			return fmt.Errorf("failed to lock %s: %s", containerLockName, err)
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
						return fmt.Errorf("introspect error failed: %s", err)
					}
				} else if options.IntrospectAfterError {
					if err := i.Commit(ctx); err != nil {
						return fmt.Errorf("introspect error failed: %s", err)
					}

					logboek.Context(ctx).Default().LogFDetails("Launched command: %s\n", strings.Join(i.container.prepareAllRunCommands(), " && "))

					if err := logboek.Context(ctx).Streams().DoErrorWithoutProxyStreamDataFormatting(func() error {
						return i.Introspect(ctx)
					}); err != nil {
						return fmt.Errorf("introspect error failed: %s", err)
					}
				}

				if err := i.container.rm(ctx); err != nil {
					return fmt.Errorf("introspect error failed: %s", err)
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
	}

	if inspect, err := i.DockerServerRuntime.GetImageInspect(ctx, i.MustGetBuiltId()); err != nil {
		return err
	} else {
		i.SetInspect(inspect)
		i.SetStageDescription(&image.StageDescription{
			StageID: nil, // stage id does not available at the moment
			Info:    image.NewInfoFromInspect(i.Name(), inspect),
		})
	}

	return nil
}

func (i *LegacyStageImage) Commit(ctx context.Context) error {
	builtId, err := i.container.commit(ctx)
	if err != nil {
		return err
	}

	i.buildImage = newLegacyBaseImage(builtId, i.DockerServerRuntime)

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

func (i *LegacyStageImage) MustResetInspect(ctx context.Context) error {
	if i.buildImage != nil {
		return i.buildImage.MustResetInspect(ctx)
	} else {
		return i.legacyBaseImage.MustResetInspect(ctx)
	}
}

func (i *LegacyStageImage) GetInspect() *types.ImageInspect {
	if i.buildImage != nil {
		return i.buildImage.GetInspect()
	} else {
		return i.legacyBaseImage.GetInspect()
	}
}

func (i *LegacyStageImage) MustGetBuiltId() string {
	builtId := i.GetBuiltId()
	if builtId == "" {
		panic(fmt.Sprintf("image %s built id is not available", i.Name()))
	}
	return builtId
}

func (i *LegacyStageImage) GetBuiltId() string {
	if i.dockerfileImageBuilder != nil {
		return i.dockerfileImageBuilder.GetBuiltId()
	} else if i.buildImage != nil {
		return i.buildImage.Name()
	} else {
		return ""
	}
}

func (i *LegacyStageImage) TagBuiltImage(ctx context.Context) error {
	return docker.CliTag(ctx, i.MustGetBuiltId(), i.name)
}

func (i *LegacyStageImage) Tag(ctx context.Context, name string) error {
	return docker.CliTag(ctx, i.GetID(), name)
}

func (i *LegacyStageImage) Pull(ctx context.Context) error {
	if err := docker.CliPullWithRetries(ctx, i.name); err != nil {
		return err
	}

	i.legacyBaseImage.UnsetInspect()

	return nil
}

func (i *LegacyStageImage) Push(ctx context.Context) error {
	return docker.CliPushWithRetries(ctx, i.name)
}

func (i *LegacyStageImage) DockerfileImageBuilder() *DockerfileImageBuilder {
	if i.dockerfileImageBuilder == nil {
		i.dockerfileImageBuilder = NewDockerfileImageBuilder()
	}
	return i.dockerfileImageBuilder
}

func debugDockerRunCommand() bool {
	return os.Getenv("WERF_DEBUG_DOCKER_RUN_COMMAND") == "1"
}
