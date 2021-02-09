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

type StageImage struct {
	*baseImage
	fromImage              *StageImage
	container              *StageImageContainer
	buildImage             *buildImage
	dockerfileImageBuilder *DockerfileImageBuilder
}

func NewStageImage(fromImage *StageImage, name string, localDockerServerRuntime *LocalDockerServerRuntime) *StageImage {
	stage := &StageImage{}
	stage.baseImage = newBaseImage(name, localDockerServerRuntime)
	stage.fromImage = fromImage
	stage.container = newStageImageContainer(stage)
	return stage
}

func (i *StageImage) Inspect() *types.ImageInspect {
	return i.inspect
}

func (i *StageImage) BuilderContainer() BuilderContainer {
	return &StageImageBuilderContainer{i}
}

func (i *StageImage) Container() Container {
	return i.container
}

func (i *StageImage) GetID() string {
	if i.buildImage != nil {
		return i.buildImage.Name()
	} else {
		return i.baseImage.GetStageDescription().Info.ID
	}
}

func (i *StageImage) Build(ctx context.Context, options BuildOptions) error {
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

	if inspect, err := i.LocalDockerServerRuntime.GetImageInspect(ctx, i.MustGetBuiltId()); err != nil {
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

func (i *StageImage) Commit(ctx context.Context) error {
	builtId, err := i.container.commit(ctx)
	if err != nil {
		return err
	}

	i.buildImage = newBuildImage(builtId, i.LocalDockerServerRuntime)

	return nil
}

func (i *StageImage) Introspect(ctx context.Context) error {
	if err := i.container.introspect(ctx); err != nil {
		return err
	}

	return nil
}

func (i *StageImage) introspectBefore(ctx context.Context) error {
	if err := i.container.introspectBefore(ctx); err != nil {
		return err
	}

	return nil
}

func (i *StageImage) MustResetInspect(ctx context.Context) error {
	if i.buildImage != nil {
		return i.buildImage.MustResetInspect(ctx)
	} else {
		return i.baseImage.MustResetInspect(ctx)
	}
}

func (i *StageImage) GetInspect() *types.ImageInspect {
	if i.buildImage != nil {
		return i.buildImage.GetInspect()
	} else {
		return i.baseImage.GetInspect()
	}
}

func (i *StageImage) MustGetBuiltId() string {
	builtId := i.GetBuiltId()
	if builtId == "" {
		panic(fmt.Sprintf("image %s built id is not available", i.Name()))
	}
	return builtId
}

func (i *StageImage) GetBuiltId() string {
	if i.dockerfileImageBuilder != nil {
		return i.dockerfileImageBuilder.GetBuiltId()
	} else if i.buildImage != nil {
		return i.buildImage.Name()
	} else {
		return ""
	}
}

func (i *StageImage) TagBuiltImage(ctx context.Context) error {
	return docker.CliTag(ctx, i.MustGetBuiltId(), i.name)
}

func (i *StageImage) Tag(ctx context.Context, name string) error {
	return docker.CliTag(ctx, i.GetID(), name)
}

func (i *StageImage) Pull(ctx context.Context) error {
	if err := docker.CliPullWithRetries(ctx, i.name); err != nil {
		return err
	}

	i.baseImage.UnsetInspect()

	return nil
}

func (i *StageImage) Push(ctx context.Context) error {
	return docker.CliPushWithRetries(ctx, i.name)
}

func (i *StageImage) DockerfileImageBuilder() *DockerfileImageBuilder {
	if i.dockerfileImageBuilder == nil {
		i.dockerfileImageBuilder = NewDockerfileImageBuilder()
	}
	return i.dockerfileImageBuilder
}

func debugDockerRunCommand() bool {
	return os.Getenv("WERF_DEBUG_DOCKER_RUN_COMMAND") == "1"
}
