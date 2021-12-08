package util

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/werf/werf/pkg/docker"
)

func RemoveHostDirsWithLinuxContainer(ctx context.Context, mountDir string, dirs []string) error {
	var containerDirs []string
	for _, dir := range dirs {
		containerDirs = append(containerDirs, ToLinuxContainerPath(dir))
	}

	args := []string{
		"--rm",
		"--volume", fmt.Sprintf("%s:%s", mountDir, ToLinuxContainerPath(mountDir)),
		"alpine",
		"rm", "-rf",
	}

	args = append(args, containerDirs...)

	return docker.CliRun(ctx, args...)
}

func ToLinuxContainerPath(path string) string {
	return filepath.ToSlash(
		strings.TrimPrefix(
			path,
			filepath.VolumeName(path),
		),
	)
}

func IsInContainer() (bool, error) {
	if dockerEnvExist, err := RegularFileExists("/.dockerenv"); err != nil {
		return false, fmt.Errorf("unable to check for /.dockerenv existence: %s", err)
	} else if dockerEnvExist {
		return true, nil
	}

	if containerEnvExist, err := RegularFileExists("/run/.containerenv"); err != nil {
		return false, fmt.Errorf("unable to check for /run/.containerenv existence: %s", err)
	} else if containerEnvExist {
		return true, nil
	}

	return false, nil
}
