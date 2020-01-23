package tmp_manager

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/flant/werf/pkg/werf"
)

const (
	projectsServiceDir          = "projects"
	dockerConfigsServiceDir     = "docker_configs"
	werfConfigRendersServiceDir = "werf_config_renders"

	CommonPrefix           = "werf-"
	ProjectDirPrefix       = CommonPrefix + "project-data-"
	DockerConfigDirPrefix  = CommonPrefix + "docker-config-"
	WerfConfigRenderPrefix = CommonPrefix + "config-render-"
	HelmTmpChartDestPrefix = CommonPrefix + "helm-chart-dest-"
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

func CreateHelmTmpChartDestDir() (string, error) {
	newDir, err := newTmpDir(HelmTmpChartDestPrefix)
	if err != nil {
		return "", err
	}

	if err := os.Chmod(newDir, 0700); err != nil {
		return "", err
	}

	return newDir, nil
}

func registerCreatedPath(newPath, createdPathsDir string) error {
	if err := os.MkdirAll(createdPathsDir, os.ModePerm); err != nil {
		return fmt.Errorf("unable to create dir %s: %s", createdPathsDir, err)
	}

	createdPath := filepath.Join(createdPathsDir, filepath.Base(newPath))
	if err := os.Symlink(newPath, createdPath); err != nil {
		return fmt.Errorf("unable to create symlink %s -> %s: %s", createdPath, newPath, err)
	}

	return nil
}

func releasePath(path, createdPathsDir, releasedPathsDir string) error {
	if err := os.MkdirAll(releasedPathsDir, os.ModePerm); err != nil {
		return fmt.Errorf("unable to create dir %s: %s", releasedPathsDir, err)
	}

	releasedPath := filepath.Join(releasedPathsDir, filepath.Base(path))
	if err := os.Symlink(path, releasedPath); err != nil {
		return fmt.Errorf("unable to create symlink %s -> %s: %s", releasedPath, path, err)
	}

	createdPath := filepath.Join(createdPathsDir, filepath.Base(path))
	if err := os.Remove(createdPath); err != nil {
		return fmt.Errorf("unable to remove %s: %s", createdPath, err)
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
