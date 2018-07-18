package dappdeps

import (
	"fmt"
	dockerClient "github.com/docker/cli/cli/command"
	dockerClientContainer "github.com/docker/cli/cli/command/container"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

type Container struct {
	Ref       string
	ImageName string
	Volume    string
}

func (c *Container) isExist(dockerApiClient *client.Client) bool {
	ctx := context.Background()
	if _, err := dockerApiClient.ContainerInspect(ctx, c.Ref); err != nil {
		if client.IsErrNotFound(err) {
			return false
		}
		panic(fmt.Sprintf("dappdeps container inspect failed: %s", err.Error()))
	}
	return true
}

func (i *Container) Create(dockerClient *dockerClient.DockerCli) {
	command := dockerClientContainer.NewCreateCommand(dockerClient)
	command.SilenceErrors = true
	command.SilenceUsage = true

	name := fmt.Sprintf("--name=%s", i.Ref)
	volume := fmt.Sprintf("--volume=%s", i.Volume)
	command.SetArgs([]string{name, volume, i.ImageName})

	err := command.Execute()
	if err != nil {
		panic(fmt.Sprintf("dappdeps container create failed: %s", err.Error()))
	}
}
