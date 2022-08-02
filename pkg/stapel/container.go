package stapel

import (
	"context"
	"fmt"
	"time"

	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/werf"
)

type container struct {
	Name      string
	ImageName string
	Volume    string
}

func (c *container) Create(ctx context.Context) error {
	name := fmt.Sprintf("--name=%s", c.Name)
	volume := fmt.Sprintf("--volume=%s", c.Volume)

	if exist, err := docker.ImageExist(ctx, c.ImageName); err != nil {
		return err
	} else if !exist {
		if err := docker.CliPullWithRetries(ctx, c.ImageName); err != nil {
			return err
		}
	}

	return docker.CliCreate(ctx, name, volume, c.ImageName)
}

func (c *container) CreateIfNotExist(ctx context.Context) error {
	exist, err := docker.ContainerExist(ctx, c.Name)
	if err != nil {
		return err
	}

	if !exist {
		err := werf.WithHostLock(ctx, fmt.Sprintf("stapel.container.%s", c.Name), lockgate.AcquireOptions{Timeout: time.Second * 600}, func() error {
			return logboek.Context(ctx).LogProcess("Creating container %s from image %s", c.Name, c.ImageName).DoError(func() error {
				exist, err := docker.ContainerExist(ctx, c.Name)
				if err != nil {
					return err
				}

				if !exist {
					if err := c.Create(ctx); err != nil {
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

func (c *container) RmIfExist(ctx context.Context) error {
	exist, err := docker.ContainerExist(ctx, c.Name)
	if err != nil {
		return err
	}

	if exist {
		inspect, err := docker.ContainerInspect(ctx, c.Name)
		if err != nil {
			return err
		}

		if err := docker.CliRm(ctx, c.Name); err != nil {
			return err
		}

		for _, m := range inspect.Mounts {
			if m.Type == "volume" {
				if err := docker.VolumeRm(ctx, m.Name, true); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
