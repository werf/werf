package util

import (
	"context"
	"fmt"
	"os"
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

func IsInContainer() bool {
	werfContainerized := GetBoolEnvironment("WERF_CONTAINERIZED")
	if werfContainerized != nil {
		return *werfContainerized
	}

	// Docker-daemon
	if isInContainer, err := RegularFileExists("/.dockerenv"); err == nil && isInContainer {
		return true
	}

	// Podman, CRI-O
	if isInContainer, err := RegularFileExists("/run/.containerenv"); err == nil && isInContainer {
		return true
	}

	cgroupsData, err := os.ReadFile("/proc/1/cgroup")
	if err != nil {
		return false
	}

	// containerd without Docker-daemon
	if strings.Contains(string(cgroupsData), "/cri-containerd-") ||
		strings.Contains(string(cgroupsData), "/containerd") {
		return true
	}

	// If in Kubernetes
	if strings.Contains(string(cgroupsData), "/kubepods") {
		return true
	}

	return false
}
