package stapel

import (
	"context"

	"github.com/werf/werf/v2/pkg/docker"
)

type container struct {
	Name      string
	ImageName string
	Volume    string
	Platform  string
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
