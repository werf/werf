package dappdeps

import (
	"fmt"
	"time"

	"github.com/docker/cli/cli/command"
	commandContainer "github.com/docker/cli/cli/command/container"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"

	"github.com/flant/dapp/pkg/lock"
)

type container struct {
	Name      string
	ImageName string
	Volume    string
}

func (c *container) isExist(apiClient *client.Client) (bool, error) {
	ctx := context.Background()
	if _, err := apiClient.ContainerInspect(ctx, c.Name); err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *container) Create(cli *command.DockerCli) error {
	cmd := commandContainer.NewCreateCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	name := fmt.Sprintf("--name=%s", c.Name)
	volume := fmt.Sprintf("--volume=%s", c.Volume)
	cmd.SetArgs([]string{name, volume, c.ImageName})

	err := cmd.Execute()
	if err != nil {
		return err
	}
	return nil
}

func (c *container) CreateIfNotExist(cli *command.DockerCli, apiClient *client.Client) error {
	exist, err := c.isExist(apiClient)
	if err != nil {
		return err
	}

	if !exist {
		err := lock.WithLock(fmt.Sprintf("dappdeps.container.%s", c.Name), lock.LockOptions{Timeout: time.Second * 600}, func() error {
			exist, err := c.isExist(apiClient)
			if err != nil {
				return err
			}

			if !exist {
				if err := c.Create(cli); err != nil {
					return err
				}
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}
