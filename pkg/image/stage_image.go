package image

import (
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"

	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/lock"
)

type StageImage struct {
	*base
	fromImage  *StageImage
	container  *StageImageContainer
	buildImage *build
}

func NewStageImage(fromImage *StageImage, name string) *StageImage {
	stage := &StageImage{}
	stage.base = newBaseImage(name)
	stage.fromImage = fromImage
	stage.container = newStageImageContainer(stage)
	return stage
}

func (i *StageImage) Inspect() *types.ImageInspect {
	return i.inspect
}

func (i *StageImage) Labels() map[string]string {
	if i.inspect != nil {
		return i.inspect.Config.Labels
	}
	return nil
}

func (i *StageImage) BuilderContainer() BuilderContainer {
	return &StageImageBuilderContainer{i}
}

func (i *StageImage) Container() Container {
	return i.container
}

func (i *StageImage) MustGetInspect() (*types.ImageInspect, error) {
	if i.buildImage != nil {
		return i.buildImage.MustGetInspect()
	} else {
		return i.base.MustGetInspect()
	}
}

func (i *StageImage) MustGetId() (string, error) {
	if i.buildImage != nil {
		return i.buildImage.MustGetId()
	} else {
		return i.base.MustGetId()
	}
}

func (i *StageImage) ID() string {
	if i.inspect != nil {
		return i.inspect.ID
	}
	return ""
}

func (i *StageImage) IsExists() bool {
	return i.inspect != nil
}

func (i *StageImage) SyncDockerState() error {
	if err := i.ResetInspect(); err != nil {
		return fmt.Errorf("image %s inspect failed: %s", i.name, err)
	}
	return nil
}

func (i *StageImage) Build(options BuildOptions) error {
	containerLockName := ContainerLockName(i.container.Name())
	if err := lock.Lock(containerLockName, lock.LockOptions{}); err != nil {
		return fmt.Errorf("failed to lock %s: %s", containerLockName, err)
	}
	defer lock.Unlock(containerLockName)

	if containerRunErr := i.container.run(); containerRunErr != nil {
		if strings.HasPrefix(containerRunErr.Error(), "container run failed") {
			if options.IntrospectBeforeError {
				logboek.LogInfoF("Launched command: %s\n", strings.Join(i.container.prepareAllRunCommands(), " && "))
				if err := logboek.WithRawStreamsOutputModeOn(i.introspectBefore); err != nil {
					return fmt.Errorf("introspect error failed: %s", err)
				}
			} else if options.IntrospectAfterError {
				if err := i.Commit(); err != nil {
					return fmt.Errorf("introspect error failed: %s", err)
				}

				logboek.LogInfoF("Launched command: %s\n", strings.Join(i.container.prepareAllRunCommands(), " && "))
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

	return nil
}

func (i *StageImage) Commit() error {
	builtId, err := i.container.commit()
	if err != nil {
		return err
	}

	i.buildImage = newBuildImage(builtId)

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

func (i *StageImage) SaveInCache() error {
	buildImageId, err := i.buildImage.MustGetId()
	if err != nil {
		return err
	}

	if err := docker.CliTag(buildImageId, i.name); err != nil {
		return err
	}

	if err := i.SyncDockerState(); err != nil {
		return err
	}

	return nil
}

func (i *StageImage) Tag(name string) error {
	imageId, err := i.MustGetId()
	if err != nil {
		return err
	}

	if err := docker.CliTag(imageId, name); err != nil {
		return err
	}

	return nil
}

func (i *StageImage) Pull() error {
	if err := docker.CliPull(i.name); err != nil {
		return err
	}

	i.base.unsetInspect()

	return nil
}

func (i *StageImage) Push() error {
	return docker.CliPush(i.name)
}

func (i *StageImage) Import(name string) error {
	importedImage := newBaseImage(name)

	if err := docker.CliPull(name); err != nil {
		return err
	}

	importedImageId, err := importedImage.MustGetId()
	if err != nil {
		return err
	}

	if err := docker.CliTag(importedImageId, i.name); err != nil {
		return err
	}

	if err := docker.CliRmi(name); err != nil {
		return err
	}

	return nil
}

func (i *StageImage) Export(name string) error {
	if err := logboek.LogProcess(fmt.Sprintf("Tagging %s", name), logboek.LogProcessOptions{}, func() error {
		return i.Tag(name)
	}); err != nil {
		return err
	}

	if err := logboek.LogProcess(fmt.Sprintf("Pushing %s", name), logboek.LogProcessOptions{}, func() error {
		return docker.CliPush(name)
	}); err != nil {
		return err
	}

	if err := logboek.LogProcess(fmt.Sprintf("Untagging %s", name), logboek.LogProcessOptions{}, func() error {
		return docker.CliRmi(name)
	}); err != nil {
		return err
	}

	return nil
}
