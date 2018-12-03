package image

import (
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"

	"github.com/flant/dapp/pkg/docker"
)

type Stage struct {
	*base
	fromImage  *Stage
	container  *StageContainer
	buildImage *build
}

func NewStageImage(fromImage *Stage, name string) *Stage {
	stage := &Stage{}
	stage.base = newBaseImage(name)
	stage.fromImage = fromImage
	stage.container = newStageImageContainer(stage)
	return stage
}

func (i *Stage) Labels() map[string]string {
	if i.inspect != nil {
		return i.inspect.Config.Labels
	}
	return nil
}

func (i *Stage) BuilderContainer() BuilderContainer {
	return &StageBuilderContainer{i}
}

func (i *Stage) Container() Container {
	return i.container
}

func (i *Stage) MustGetInspect() (*types.ImageInspect, error) {
	if i.buildImage != nil {
		return i.buildImage.MustGetInspect()
	} else {
		return i.base.MustGetInspect()
	}
}

func (i *Stage) MustGetId() (string, error) {
	if i.buildImage != nil {
		return i.buildImage.MustGetId()
	} else {
		return i.base.MustGetId()
	}
}

func (i *Stage) IsExists() (bool, error) {
	inspect, err := i.GetInspect()
	if err != nil {
		return false, err
	}

	exist := inspect != nil

	return exist, nil
}

func (i *Stage) ReadDockerState() error {
	_, err := i.GetInspect()
	if err != nil {
		return fmt.Errorf("image %s inspect failed: %s", i.name, err)
	}
	return nil
}

type StageBuildOptions struct {
	IntrospectBeforeError bool
	IntrospectAfterError  bool
}

func (i *Stage) Build(options *StageBuildOptions) error {
	if containerRunErr := i.container.run(); containerRunErr != nil {
		if strings.HasPrefix(containerRunErr.Error(), "container run failed") {
			if options.IntrospectBeforeError {
				fmt.Printf("Launched command: %s\n", strings.Join(i.container.prepareAllRunCommands(), " && "))
				if err := i.introspectBefore(); err != nil {
					return fmt.Errorf("introspect error failed: %s", err)
				}
			} else if options.IntrospectAfterError {
				if err := i.Commit(); err != nil {
					return fmt.Errorf("introspect error failed: %s", err)
				}

				fmt.Printf("Launched command: %s\n", strings.Join(i.container.prepareAllRunCommands(), " && "))
				if err := i.Introspect(); err != nil {
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

func (i *Stage) Commit() error {
	builtId, err := i.container.commit()
	if err != nil {
		return err
	}

	i.buildImage = newBuildImage(builtId)

	return nil
}

func (i *Stage) Introspect() error {
	if err := i.container.introspect(); err != nil {
		return err
	}

	return nil
}

func (i *Stage) introspectBefore() error {
	if err := i.container.introspectBefore(); err != nil {
		return err
	}

	return nil
}

func (i *Stage) SaveInCache() error {
	buildImageId, err := i.buildImage.MustGetId()
	if err != nil {
		return err
	}

	if err := docker.CliTag(buildImageId, i.name); err != nil {
		return err
	}

	return nil
}

func (i *Stage) Tag(name string) error {
	imageId, err := i.MustGetId()
	if err != nil {
		return err
	}

	if err := docker.CliTag(imageId, name); err != nil {
		return err
	}

	return nil
}

func (i *Stage) Pull() error {
	if err := docker.CliPull(i.name); err != nil {
		return err
	}

	i.base.unsetInspect()

	return nil
}

func (i *Stage) Push() error {
	return docker.CliPush(i.name)
}

func (i *Stage) Import(name string) error {
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

func (i *Stage) Export(name string) error {
	if err := i.Tag(name); err != nil {
		return err
	}

	if err := docker.CliPush(name); err != nil {
		return err
	}

	if err := docker.CliRmi(name); err != nil {
		return err
	}

	return nil
}
