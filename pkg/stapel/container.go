package stapel

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/docker"
	"github.com/werf/werf/v2/pkg/werf"
)

type container struct {
	Name      string
	ImageName string
	Volume    string
	Platform  string
}

func (c *container) Create(ctx context.Context) error {
	name := fmt.Sprintf("--name=%s", c.Name)
	volume := fmt.Sprintf("--volume=%s", c.Volume)
	targetPlatform := c.Platform
	if targetPlatform == "" {
		targetPlatform = docker.GetDefaultPlatform()
	}

	if exist, err := docker.ImageExist(ctx, c.ImageName); err != nil {
		return err
	} else if exist {
		if targetPlatform != "" {
			inspect, err := docker.ImageInspect(ctx, c.ImageName)
			if err != nil {
				return err
			}

			actualPlatform := fmt.Sprintf("%s/%s", inspect.Os, inspect.Architecture)
			if inspect.Variant != "" {
				actualPlatform = fmt.Sprintf("%s/%s", actualPlatform, inspect.Variant)
			}

			platformMatches := targetPlatform == actualPlatform || strings.HasPrefix(targetPlatform, actualPlatform+"/") || strings.HasPrefix(actualPlatform, targetPlatform+"/")
			if !platformMatches {
				if err := docker.CliPullWithRetries(ctx, "--platform", targetPlatform, c.ImageName); err != nil {
					return err
				}
			}
		}
	} else {
		pullArgs := []string{c.ImageName}
		if targetPlatform != "" {
			pullArgs = append([]string{"--platform", targetPlatform}, pullArgs...)
		}

		if err := docker.CliPullWithRetries(ctx, pullArgs...); err != nil {
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
		err := werf.HostLocker().WithLock(ctx, fmt.Sprintf("stapel.container.%s", c.Name), lockgate.AcquireOptions{Timeout: time.Second * 600}, func() error {
			return logboek.Context(ctx).LogProcess("Creating container %s from image %s", c.Name, c.ImageName).DoError(func() error {
				exist, err := docker.ContainerExist(ctx, c.Name)
				if err != nil {
					return err
				}

				if !exist {
					if err := c.Create(ctx); err != nil {
						if docker.IsContainerNameConflict(err) {
							return nil
						}
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
