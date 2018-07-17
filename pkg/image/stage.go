package image

import (
	"fmt"

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
		return nil, fmt.Errorf("stage inspect failed: %s", err)
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

func (i *Stage) Build(dockerClient *command.DockerCli, dockerApiClient *client.Client) error {
	if err := i.Container.Run(dockerClient); err != nil {
		return fmt.Errorf("stage build failed: %s", err)
	}

	if err := i.Commit(dockerApiClient); err != nil {
		return fmt.Errorf("stage build failed: %s", err)
	}

	if err := i.Container.Rm(dockerApiClient); err != nil {
		return fmt.Errorf("stage build failed: %s", err)
	}

	return nil
}

func (i *Stage) Commit(dockerApiClient *client.Client) error {
	builtId, err := i.Container.Commit(dockerApiClient)
	if err != nil {
		return fmt.Errorf("stage commit failed: %s", err)
	}
	i.BuiltId = builtId

	return nil
}

func (i *Stage) Introspect(dockerClient *command.DockerCli) error {
	if err := i.Container.Introspect(dockerClient); err != nil {
		return fmt.Errorf("stage introspect failed: %s", err)
	}

	return nil
}
