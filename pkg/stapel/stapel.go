package stapel

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/werf/werf/v2/pkg/docker"
	"github.com/werf/werf/v2/pkg/image"
)

const (
	VERSION              = "0.7.1"
	IMAGE                = "registry.werf.io/werf/stapel"
	CONTAINER_MOUNT_ROOT = "/.werf"
)

func getVersion() string {
	version := VERSION
	if v := os.Getenv("WERF_STAPEL_IMAGE_VERSION"); v != "" {
		version = v
	}
	return version
}

func getImage() string {
	image := IMAGE
	if i := os.Getenv("WERF_STAPEL_IMAGE_NAME"); i != "" {
		image = i
	}
	return image
}

func ImageName() string {
	return fmt.Sprintf("%s:%s", getImage(), getVersion())
}

func getContainer(targetPlatform string) container {
	containerNameSuffix := getVersion()
	if targetPlatform != "" {
		containerNameSuffix = fmt.Sprintf("%s_%s", containerNameSuffix, strings.ReplaceAll(targetPlatform, "/", "_"))
	}

	return container{
		Name:      fmt.Sprintf("%s%s", image.AssemblingContainerNamePrefix, containerNameSuffix),
		ImageName: ImageName(),
		Volume:    path.Join(CONTAINER_MOUNT_ROOT, "stapel"),
		Platform:  targetPlatform,
	}
}

func Purge(ctx context.Context) error {
	container := getContainer("")
	if err := container.RmIfExist(ctx); err != nil {
		return err
	}

	if err := rmiIfExist(ctx); err != nil {
		return err
	}

	return nil
}

func rmiIfExist(ctx context.Context) error {
	exist, err := docker.ImageExist(ctx, ImageName())
	if err != nil {
		return err
	}

	if exist {
		return docker.CliRmi(ctx, ImageName())
	}

	return nil
}
