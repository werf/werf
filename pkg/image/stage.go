package image

import (
	"fmt"
	"strings"

	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
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

func (i *Stage) MustGetInspect(apiClient *client.Client) (*types.ImageInspect, error) {
	if i.BuildImage != nil {
		return i.BuildImage.MustGetInspect(apiClient)
	} else {
		return i.Base.MustGetInspect(apiClient)
	}
}

func (i *Stage) MustGetId(apiClient *client.Client) (string, error) {
	if i.BuildImage != nil {
		return i.BuildImage.MustGetId(apiClient)
	} else {
		return i.Base.MustGetId(apiClient)
	}
}

func (i *Stage) Build(options *StageBuildOptions, cli *command.DockerCli, apiClient *client.Client) error {
	if containerRunErr := i.Container.Run(cli, apiClient); containerRunErr != nil {
		if strings.HasPrefix(containerRunErr.Error(), "container run failed") {
			if options.IntrospectBeforeError {
				if err := i.IntrospectBefore(cli, apiClient); err != nil {
					return fmt.Errorf("introspect error failed: %s", err)
				}
			} else if options.IntrospectAfterError {
				if err := i.Commit(apiClient); err != nil {
					return fmt.Errorf("introspect error failed: %s", err)
				}
				if err := i.Introspect(cli, apiClient); err != nil {
					return fmt.Errorf("introspect error failed: %s", err)
				}
			}

			if err := i.Container.Rm(apiClient); err != nil {
				return fmt.Errorf("introspect error failed: %s", err)
			}
		}

		return containerRunErr
	}

	if err := i.Commit(apiClient); err != nil {
		return err
	}

	if err := i.Container.Rm(apiClient); err != nil {
		return err
	}

	return nil
}

type StageBuildOptions struct {
	IntrospectBeforeError bool
	IntrospectAfterError  bool
}

func (i *Stage) Commit(apiClient *client.Client) error {
	builtId, err := i.Container.Commit(apiClient)
	if err != nil {
		return err
	}

	i.BuildImage = NewBuildImage(builtId)

	return nil
}

func (i *Stage) Introspect(cli *command.DockerCli, apiClient *client.Client) error {
	if err := i.Container.Introspect(cli, apiClient); err != nil {
		return err
	}

	return nil
}

func (i *Stage) IntrospectBefore(cli *command.DockerCli, apiClient *client.Client) error {
	if err := i.Container.IntrospectBefore(cli, apiClient); err != nil {
		return err
	}

	return nil
}

func (i *Stage) SaveInCache(cli *command.DockerCli, apiClient *client.Client) error {
	if err := i.BuildImage.Tag(i.Name, cli, apiClient); err != nil {
		return err
	}
	return nil
}
