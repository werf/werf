package container_runtime

import (
	"fmt"
	"os"
	"strings"

	"github.com/flant/lockgate"
	"github.com/werf/werf/pkg/werf"

	"github.com/docker/docker/api/types"

	"github.com/flant/logboek"

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

func (i *StageImage) Build(options BuildOptions) error {
	if i.dockerfileImageBuilder != nil {
		if err := i.dockerfileImageBuilder.Build(); err != nil {
			return err
		}
	} else {
		containerLockName := ContainerLockName(i.container.Name())
		if _, lock, err := werf.AcquireHostLock(containerLockName, lockgate.AcquireOptions{}); err != nil {
			return fmt.Errorf("failed to lock %s: %s", containerLockName, err)
		} else {
			defer werf.ReleaseHostLock(lock)
		}

		if debugDockerRunCommand() {
			runArgs, err := i.container.prepareRunArgs()
			if err != nil {
				return err
			}

			fmt.Printf("Docker run command:\ndocker run %s\n", strings.Join(runArgs, " "))

			if len(i.container.prepareAllRunCommands()) != 0 {
				fmt.Printf("Decoded command:\n%s\n", strings.Join(i.container.prepareAllRunCommands(), " && "))
			}
		}

		if containerRunErr := i.container.run(); containerRunErr != nil {
			if strings.HasPrefix(containerRunErr.Error(), "container run failed") {
				if options.IntrospectBeforeError {
					logboek.Default.LogFDetails("Launched command: %s\n", strings.Join(i.container.prepareAllRunCommands(), " && "))

					if err := logboek.WithRawStreamsOutputModeOn(i.introspectBefore); err != nil {
						return fmt.Errorf("introspect error failed: %s", err)
					}
				} else if options.IntrospectAfterError {
					if err := i.Commit(); err != nil {
						return fmt.Errorf("introspect error failed: %s", err)
					}

					logboek.Default.LogFDetails("Launched command: %s\n", strings.Join(i.container.prepareAllRunCommands(), " && "))

					if err := logboek.WithRawStreamsOutputModeOn(i.Introspect); err != nil {
						return fmt.Errorf("introspect error failed: %s", err)
					}
				}

				if err := i.container.rm(); err != nil {
					return fmt.Errorf("introspect error failed: %s", err)
				}
			}

			return containerRunErr
		}

		if err := i.Commit(); err != nil {
			return err
		}

		if err := i.container.rm(); err != nil {
			return err
		}
	}

	if inspect, err := i.LocalDockerServerRuntime.GetImageInspect(i.MustGetBuiltId()); err != nil {
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

func (i *StageImage) Commit() error {
	builtId, err := i.container.commit()
	if err != nil {
		return err
	}

	i.buildImage = newBuildImage(builtId, i.LocalDockerServerRuntime)

	return nil
}

func (i *StageImage) Introspect() error {
	if err := i.container.introspect(); err != nil {
		return err
	}

	return nil
}

func (i *StageImage) introspectBefore() error {
	if err := i.container.introspectBefore(); err != nil {
		return err
	}

	return nil
}

func (i *StageImage) MustResetInspect() error {
	if i.buildImage != nil {
		return i.buildImage.MustResetInspect()
	} else {
		return i.baseImage.MustResetInspect()
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

func (i *StageImage) TagBuiltImage(name string) error {
	return docker.CliTag(i.MustGetBuiltId(), i.name)
}

func (i *StageImage) Tag(name string) error {
	return docker.CliTag(i.GetID(), name)
}

func (i *StageImage) Pull() error {
	if err := docker.CliPullWithRetries(i.name); err != nil {
		return err
	}

	i.baseImage.UnsetInspect()

	return nil
}

func (i *StageImage) Push() error {
	return docker.CliPushWithRetries(i.name)
}

func (i *StageImage) Import(name string) error {
	importedImage := newBaseImage(name, i.LocalDockerServerRuntime)

	if err := docker.CliPullWithRetries(name); err != nil {
		return err
	}

	importedImageId := importedImage.GetStageDescription().Info.ID

	if err := docker.CliTag(importedImageId, i.name); err != nil {
		return err
	}

	if err := docker.CliRmi(name); err != nil {
		return err
	}

	return nil
}

func (i *StageImage) Export(name string) error {
	if err := logboek.Info.LogProcess(fmt.Sprintf("Tagging %s", name), logboek.LevelLogProcessOptions{}, func() error {
		return i.Tag(name)
	}); err != nil {
		return err
	}

	defer func() {
		if err := logboek.Info.LogProcess(fmt.Sprintf("Untagging %s", name), logboek.LevelLogProcessOptions{}, func() error {
			return docker.CliRmi(name)
		}); err != nil {
			// TODO: errored image state
			logboek.Error.LogF("Unable to remote temporary image %q: %s", name, err)
		}
	}()

	if err := logboek.Info.LogProcess(fmt.Sprintf("Pushing %s", name), logboek.LevelLogProcessOptions{}, func() error {
		return docker.CliPushWithRetries(name)
	}); err != nil {
		return err
	}

	return nil
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
