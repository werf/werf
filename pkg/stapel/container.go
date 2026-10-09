package stapel

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/containerd/containerd/platforms"

	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/werf/v3/pkg/container_backend/thirdparty/platformutil"
	"github.com/werf/werf/v3/pkg/docker"
	"github.com/werf/werf/v3/pkg/werf"
)

type container struct {
	Name      string
	ImageName string
	Volume    string
	Platform  string
}

func EnsureImage(ctx context.Context, imageName, targetPlatform string) error {
	return werf.HostLocker().WithLock(ctx, stapelImageLockName(imageName), lockgate.AcquireOptions{Timeout: time.Second * 600}, func() error {
		return ensureImage(ctx, imageName, targetPlatform)
	})
}

func (c *container) Create(ctx context.Context) error {
	return werf.HostLocker().WithLock(ctx, stapelImageLockName(c.ImageName), lockgate.AcquireOptions{Timeout: time.Second * 600}, func() error {
		if err := ensureImage(ctx, c.ImageName, c.Platform); err != nil {
			return err
		}
		return docker.CliCreate(ctx, fmt.Sprintf("--name=%s", c.Name), fmt.Sprintf("--volume=%s", c.Volume), c.ImageName)
	})
}

func ensureImage(ctx context.Context, imageName, targetPlatform string) error {
	if targetPlatform == "" {
		targetPlatform = docker.GetDefaultPlatform()
	}

	if exist, err := docker.ImageExist(ctx, imageName); err != nil {
		return err
	} else if exist {
		if targetPlatform != "" {
			inspect, err := docker.ImageInspect(ctx, imageName)
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
				if err := acquireImage(ctx, imageName, targetPlatform); err != nil {
					return err
				}
			}
		}
	} else {
		if err := acquireImage(ctx, imageName, targetPlatform); err != nil {
			return err
		}
	}
	return nil
}

func acquireImage(ctx context.Context, imageName, targetPlatform string) error {
	if isDefaultImageRef() {
		if _, ok := embeddedImageForPlatform(targetPlatform); ok {
			return loadEmbeddedImage(ctx, targetPlatform)
		}
	}

	pullArgs := []string{imageName}
	if targetPlatform != "" {
		pullArgs = append([]string{"--platform", targetPlatform}, pullArgs...)
	}

	return docker.CliPullWithRetries(ctx, pullArgs...)
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
