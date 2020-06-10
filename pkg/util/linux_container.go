package util

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/werf/werf/pkg/docker"
)

func RemoveHostDirsWithLinuxContainer(mountDir string, dirs []string) error {
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

	return docker.CliRun(args...)
}

func ToLinuxContainerPath(path string) string {
	return filepath.ToSlash(
		strings.TrimPrefix(
			path,
			filepath.VolumeName(path),
		),
	)
}
