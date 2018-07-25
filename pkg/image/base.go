package image

import (
	"fmt"

	"github.com/docker/cli/cli/command"
	commandImage "github.com/docker/cli/cli/command/image"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

type Base struct {
	Name    string
	Inspect *types.ImageInspect
}

func NewBaseImage(name string) *Base {
	image := &Base{}
	image.Name = name
	return image
}

func (i *Base) MustGetId(apiClient *client.Client) (string, error) {
	if inspect, err := i.MustGetInspect(apiClient); err == nil {
		return inspect.ID, nil
	} else {
		return "", err
	}
}

func (i *Base) MustGetInspect(apiClient *client.Client) (*types.ImageInspect, error) {
	if inspect, err := i.GetInspect(apiClient); err == nil && inspect != nil {
		return inspect, nil
	} else if err != nil {
		return nil, err
	} else {
		panic(fmt.Sprintf("runtime error: inspect must be (%s)", i.Name))
	}
}

func (i *Base) GetInspect(apiClient *client.Client) (*types.ImageInspect, error) {
	if i.Inspect == nil {
		if err := i.resetInspect(apiClient); err != nil {
			if client.IsErrNotFound(err) {
				return nil, nil
			} else {
				return nil, err
			}
		}
	}
	return i.Inspect, nil
}

func (i *Base) resetInspect(apiClient *client.Client) error {
	ctx := context.Background()
	inspect, _, err := apiClient.ImageInspectWithRaw(ctx, i.Name)
	if err != nil {
		return err
	}

	i.Inspect = &inspect
	return nil
}

func (i *Base) UnsetInspect() {
	i.Inspect = nil
}

func (i *Base) Pull(cli *command.DockerCli, apiClient *client.Client) error {
	cmd := commandImage.NewPullCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{i.Name})

	err := cmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func (i *Base) Push(cli *command.DockerCli, apiClient *client.Client) error {
	cmd := commandImage.NewPushCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{i.Name})

	err := cmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func (i *Base) Tag(tag string, cli *command.DockerCli, apiClient *client.Client) error {
	cmd := commandImage.NewTagCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{i.Name, tag})

	err := cmd.Execute()
	if err != nil {
		return err
	}

	i.UnsetInspect()

	return nil
}

func (i *Base) Untag(cli *command.DockerCli, apiClient *client.Client) error {
	cmd := commandImage.NewRemoveCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{i.Name})

	err := cmd.Execute()
	if err != nil {
		return err
	}

	i.UnsetInspect()

	return nil
}
