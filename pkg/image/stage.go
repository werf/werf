package image

import (
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"

	"github.com/flant/dapp/pkg/docker"
)

type Stage struct {
	*Base
	FromImage  *Stage
	Container  *StageContainer
	BuildImage *Build
}

func NewStageImage(fromImage *Stage, name string) *Stage {
	stage := &Stage{}
	stage.Base = NewBaseImage(name)
	stage.FromImage = fromImage
	stage.Container = NewStageImageContainer(stage)
	return stage
}

func (i *Stage) BuilderContainer() *StageBuilderContainer {
	return &StageBuilderContainer{i}
}

func (i *Stage) MustGetInspect() (*types.ImageInspect, error) {
	if i.BuildImage != nil {
		return i.BuildImage.MustGetInspect()
	} else {
		return i.Base.MustGetInspect()
	}
}

func (i *Stage) MustGetId() (string, error) {
	if i.BuildImage != nil {
		return i.BuildImage.MustGetId()
	} else {
		return i.Base.MustGetId()
	}
}

func (i *Stage) Build(options *StageBuildOptions) error {
	if containerRunErr := i.Container.Run(); containerRunErr != nil {
		if strings.HasPrefix(containerRunErr.Error(), "container run failed") {
			if options.IntrospectBeforeError {
				if err := i.IntrospectBefore(); err != nil {
					return fmt.Errorf("introspect error failed: %s", err)
				}
			} else if options.IntrospectAfterError {
				if err := i.Commit(); err != nil {
					return fmt.Errorf("introspect error failed: %s", err)
				}
				if err := i.Introspect(); err != nil {
					return fmt.Errorf("introspect error failed: %s", err)
				}
			}

			if err := i.Container.Rm(); err != nil {
				return fmt.Errorf("introspect error failed: %s", err)
			}
		}

		return containerRunErr
	}

	if err := i.Commit(); err != nil {
		return err
	}

	if err := i.Container.Rm(); err != nil {
		return err
	}

	return nil
}

type StageBuildOptions struct {
	IntrospectBeforeError bool
	IntrospectAfterError  bool
}

func (i *Stage) Commit() error {
	builtId, err := i.Container.Commit()
	if err != nil {
		return err
	}

	i.BuildImage = NewBuildImage(builtId)

	return nil
}

func (i *Stage) Introspect() error {
	if err := i.Container.Introspect(); err != nil {
		return err
	}

	return nil
}

func (i *Stage) IntrospectBefore() error {
	if err := i.Container.IntrospectBefore(); err != nil {
		return err
	}

	return nil
}

func (i *Stage) SaveInCache() error {
	buildImageId, err := i.BuildImage.MustGetId()
	if err != nil {
		return err
	}

	if err := docker.ImageTag(buildImageId, i.Name); err != nil {
		return err
	}

	return nil
}

func (i *Stage) Tag(name string) error {
	imageId, err := i.MustGetId()
	if err != nil {
		return err
	}

	if err := docker.ImageTag(imageId, name); err != nil {
		return err
	}

	return nil
}

func (i *Stage) Pull() error {
	if err := docker.ImagePull(i.Name); err != nil {
		return err
	}

	i.Base.UnsetInspect()

	return nil
}

func (i *Stage) Push() error {
	return docker.ImagePush(i.Name)
}

func (i *Stage) Import(name string) error {
	importedImage := NewBaseImage(name)

	if err := docker.ImagePull(name); err != nil {
		return err
	}

	importedImageId, err := importedImage.MustGetId()
	if err != nil {
		return err
	}

	if err := docker.ImageTag(importedImageId, i.Name); err != nil {
		return err
	}

	if err := docker.ImageUntag(name); err != nil {
		return err
	}

	return nil
}

func (i *Stage) Export(name string) error {
	if err := i.Tag(name); err != nil {
		return err
	}

	if err := docker.ImagePush(name); err != nil {
		return err
	}

	if err := docker.ImageUntag(name); err != nil {
		return err
	}

	return nil
}
