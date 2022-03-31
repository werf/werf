package tmp_manager

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/werf/werf/pkg/werf"
)

const (
	projectsServiceDir          = "projects"
	dockerConfigsServiceDir     = "docker_configs"
	kubeConfigsServiceDir       = "kubeconfigs"
	werfConfigRendersServiceDir = "werf_config_renders"

	CommonPrefix           = "werf-"
	ProjectDirPrefix       = CommonPrefix + "project-data-"
	DockerConfigDirPrefix  = CommonPrefix + "docker-config-"
	KubeConfigDirPrefix    = CommonPrefix + "kubeconfig-"
	WerfConfigRenderPrefix = CommonPrefix + "config-render-"
)

func GetServiceTmpDir() string {
	return filepath.Join(werf.GetServiceDir(), "tmp")
}

func GetCreatedTmpDirs() string {
	return filepath.Join(GetServiceTmpDir(), "created")
}

func GetReleasedTmpDirs() string {
	return filepath.Join(GetServiceTmpDir(), "released")
}

func registerCreatedPath(newPath, createdPathsDir string) error {
	if err := os.MkdirAll(createdPathsDir, os.ModePerm); err != nil {
		return fmt.Errorf("unable to create dir %s: %w", createdPathsDir, err)
	}

	createdPath := filepath.Join(createdPathsDir, filepath.Base(newPath))
	if err := os.Symlink(newPath, createdPath); err != nil {
		return fmt.Errorf("unable to create symlink %s -> %s: %w", createdPath, newPath, err)
	}

	return nil
}

func releasePath(path, createdPathsDir, releasedPathsDir string) error {
	if err := os.MkdirAll(releasedPathsDir, os.ModePerm); err != nil {
		return fmt.Errorf("unable to create dir %s: %w", releasedPathsDir, err)
	}

	releasedPath := filepath.Join(releasedPathsDir, filepath.Base(path))
	if err := os.Symlink(path, releasedPath); err != nil {
		return fmt.Errorf("unable to create symlink %s -> %s: %w", releasedPath, path, err)
	}

	createdPath := filepath.Join(createdPathsDir, filepath.Base(path))
	if err := os.Remove(createdPath); err != nil {
		return fmt.Errorf("unable to remove %s: %w", createdPath, err)
	}

	return nil
}

func newTmpDir(prefix string) (string, error) {
	newDir, err := ioutil.TempDir(werf.GetTmpDir(), prefix)
	if err != nil {
		return "", err
	}

	return newDir, nil
}

func newTmpFile(prefix string) (string, error) {
	newFile, err := ioutil.TempFile(werf.GetTmpDir(), prefix)
	if err != nil {
		return "", err
	}

	path := newFile.Name()

	err = newFile.Close()
	if err != nil {
		return "", err
	}

	return path, nil
}
