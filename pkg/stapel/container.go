package stapel

import (
	"fmt"
	"time"

	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/logger"
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
		err := lock.WithLock(fmt.Sprintf("stapel.container.%s", c.Name), lock.LockOptions{Timeout: time.Second * 600}, func() error {
			return logger.LogSecondaryProcess(fmt.Sprintf("Creating container %s from image %s", c.Name, c.ImageName), logger.LogProcessOptions{}, func() error {
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
		})

		if err != nil {
			return err
		}
	}

	return nil
}
