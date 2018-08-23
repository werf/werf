package dappdeps

import (
	"fmt"
	"time"

	"github.com/flant/dapp/pkg/docker"
	"github.com/flant/dapp/pkg/lock"
)

type container struct {
	Name      string
	ImageName string
	Volume    string
}

func (c *container) Create() error {
	name := fmt.Sprintf("--name=%s", c.Name)
	volume := fmt.Sprintf("--volume=%s", c.Volume)
	return docker.CliCreate(name, volume, c.ImageName)
}

func (c *container) CreateIfNotExist() error {
	exist, err := docker.ContainerExist(c.Name)
	if err != nil {
		return err
	}

	if !exist {
		err := lock.WithLock(fmt.Sprintf("dappdeps.container.%s", c.Name), lock.LockOptions{Timeout: time.Second * 600}, func() error {
			exist, err := docker.ContainerExist(c.Name)
			if err != nil {
				return err
			}

			if !exist {
				if err := c.Create(); err != nil {
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
