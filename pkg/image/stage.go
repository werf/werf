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
				fmt.Printf("Launched command: %s\n", strings.Join(i.Container.AllRunCommands(), " && "))
				if err := i.IntrospectBefore(); err != nil {
					return fmt.Errorf("introspect error failed: %s", err)
				}
			} else if options.IntrospectAfterError {
				if err := i.Commit(); err != nil {
					return fmt.Errorf("introspect error failed: %s", err)
				}

				fmt.Printf("Launched command: %s\n", strings.Join(i.Container.AllRunCommands(), " && "))
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

	if err := docker.CliTag(buildImageId, i.Name); err != nil {
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
	if err := docker.CliPull(i.Name); err != nil {
		return err
	}

	i.Base.UnsetInspect()

	return nil
}

func (i *Stage) Push() error {
	return docker.CliPush(i.Name)
}

func (i *Stage) Import(name string) error {
	importedImage := NewBaseImage(name)

	if err := docker.CliPull(name); err != nil {
		return err
	}

	importedImageId, err := importedImage.MustGetId()
	if err != nil {
		return err
	}

	if err := docker.CliTag(importedImageId, i.Name); err != nil {
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

func (i *Stage) AddServiceRunCommands(commands []string) {
	i.Container.ServiceRunCommands = append(i.Container.ServiceRunCommands, commands...)
}

func (i *Stage) AddRunCommands(commands []string) {
	i.Container.RunCommands = append(i.Container.RunCommands, commands...)
}

func (i *Stage) AddEnv(env map[string]interface{}) {
	i.Container.RunOptions.AddEnv(env)
}

func (i *Stage) AddVolume(volume string) {
	i.Container.RunOptions.AddVolume([]string{volume})
}

func (i *Stage) AddServiceChangeLabel(name, value string) {
	i.Container.ServiceCommitChangeOptions.AddLabel(map[string]interface{}{name: value})
}

func (i *Stage) ReadDockerState() error {
	_, err := i.GetInspect()
	if err != nil {
		return fmt.Errorf("image %s inspect failed: %s", i.Name, err)
	}
	return nil
}

func (i *Stage) GetLabels() map[string]string {
	if i.Inspect != nil {
		return i.Inspect.Config.Labels
	}
	return nil
}

func (i *Stage) IsImageExists() bool {
	return i.Inspect != nil
}
