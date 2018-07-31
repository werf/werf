package image

import (
	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/client"
)

type Dimg struct {
	*Stage
}

func NewDimgImage(fromImage *Stage, name string) *Dimg {
	return &Dimg{Stage: NewStageImage(fromImage, name)}
}

func (i *Dimg) Tag(cli *command.DockerCli, apiClient *client.Client) error {
	return i.Stage.Tag(i.Name, cli, apiClient)
}

func (i *Dimg) Export(cli *command.DockerCli, apiClient *client.Client) error {
	return i.Stage.Export(i.Name, cli, apiClient)
}
