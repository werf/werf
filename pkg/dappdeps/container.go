package dappdeps

import (
	"fmt"
	"time"

	"github.com/docker/docker/client"

	"github.com/flant/dapp/pkg/docker"
	"github.com/flant/dapp/pkg/lock"
)

type container struct {
	Name      string
	ImageName string
	Volume    string
}

func (c *container) isExist() (bool, error) {
	if _, err := docker.ContainerInspect(c.Name); err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *container) Create() error {
	name := fmt.Sprintf("--name=%s", c.Name)
	volume := fmt.Sprintf("--volume=%s", c.Volume)
	args := []string{name, volume, c.ImageName}

	return docker.ContainerCreate(args)
}

func (c *container) CreateIfNotExist() error {
	exist, err := c.isExist()
	if err != nil {
		return err
	}

	if !exist {
		err := lock.WithLock(fmt.Sprintf("dappdeps.container.%s", c.Name), lock.LockOptions{Timeout: time.Second * 600}, func() error {
			exist, err := c.isExist()
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
