package image

import (
	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

type Stage struct {
	Name         string
	Id           string
	From         *Stage
	BuiltId      string
	Container    *StageContainer
	BuiltInspect *types.ImageInspect
	Inspect      *types.ImageInspect
}

func (i *Stage) ResetBuiltInspect(dockerApiClient *client.Client) error {
	inspect, err := inspect(dockerApiClient, i.BuiltId)
	if err != nil {
		return err
	}

	i.BuiltInspect = inspect
	return nil
}

func (i *Stage) ResetInspect(dockerApiClient *client.Client) error {
	inspect, err := inspect(dockerApiClient, i.Name)
	if err != nil {
		return err
	}

	i.Inspect = inspect
	return nil
}

func inspect(dockerApiClient *client.Client, imageId string) (*types.ImageInspect, error) {
	ctx := context.Background()
	inspect, _, err := dockerApiClient.ImageInspectWithRaw(ctx, imageId)
	if err != nil {
		if client.IsErrNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &inspect, nil
}

func NewStageImage(from *Stage, name string, builtId string) *Stage {
	stage := &Stage{}
	stage.From = from
	stage.BuiltId = builtId
	stage.Name = name
	stage.Container = NewStageImageContainer()
	stage.Container.Image = stage
	return stage
}

func (i *Stage) NewStageImage(name string) *Stage {
	stage := NewStageImage(i.From, i.Name, i.BuiltId)
	return stage
}

func (i *Stage) Build(dockerClient *command.DockerCli, dockerApiClient *client.Client) error {
	if err := i.Container.Run(dockerClient); err != nil {
		return err
	}

	builtId, err := i.Container.Commit(dockerApiClient)
	if err != nil {
		return err
	}
	i.BuiltId = builtId

	if err := i.Container.Rm(dockerApiClient); err != nil {
		return err
	}

	return nil
}
