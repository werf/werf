package stapel

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/containerd/containerd/platforms"

	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/container_backend/thirdparty/platformutil"
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
	imageLockName := stapelImageLockName(c.ImageName)
	return werf.HostLocker().WithLock(ctx, imageLockName, lockgate.AcquireOptions{Timeout: time.Second * 600}, func() error {
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

				actualSpec, err := platformutil.ParsePlatform(fmt.Sprintf("%s/%s", inspect.Os, inspect.Architecture))
				if err != nil {
					return fmt.Errorf("parse cached image platform: %w", err)
				}
				if inspect.Variant != "" {
					actualSpec.Variant = inspect.Variant
				}

				desiredSpec, err := platformutil.ParsePlatform(targetPlatform)
				if err != nil {
					return fmt.Errorf("parse target platform: %w", err)
				}

				if !platforms.Only(platforms.Normalize(desiredSpec)).Match(actualSpec) {
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
	})
}

func stapelImageLockName(imageName string) string {
	return fmt.Sprintf("stapel.image.%s", strings.NewReplacer("/", "_", ":", "_", "@", "_").Replace(imageName))
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
