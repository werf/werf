package stapel

import (
	"fmt"
	"time"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/lock"
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
			return logboek.LogProcess(fmt.Sprintf("Creating container %s from image %s", c.Name, c.ImageName), logboek.LogProcessOptions{}, func() error {
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

func (c *container) RmIfExist() error {
	exist, err := docker.ContainerExist(c.Name)
	if err != nil {
		return err
	}

	if exist {
		inspect, err := docker.ContainerInspect(c.Name)
		if err != nil {
			return err
		}

		if err := docker.CliRm(c.Name); err != nil {
			return err
		}

		for _, m := range inspect.Mounts {
			if m.Type == "volume" {
				if err := docker.VolumeRm(m.Name, true); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
