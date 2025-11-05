package tmp_manager

import (
	"os"
	"path/filepath"

	"github.com/werf/werf/v2/pkg/werf"
)

const (
	projectsServiceDir          = "projects"
	dockerConfigsServiceDir     = "docker_configs"
	kubeConfigsServiceDir       = "kubeconfigs"
	werfConfigRendersServiceDir = "werf_config_renders"
	contextArchivesDir          = "context"

	commonPrefix           = "werf-"
	contextArchivePrefix   = commonPrefix + "context-"
	projectDirPrefix       = commonPrefix + "project-data-"
	dockerConfigDirPrefix  = commonPrefix + "docker-config-"
	kubeConfigDirPrefix    = commonPrefix + "kubeconfig-"
	werfConfigRenderPrefix = commonPrefix + "config-render-"
)

func getServiceTmpDir() string {
	return filepath.Join(werf.GetServiceDir(), "tmp")
}

func getCreatedTmpDirs() string {
	return filepath.Join(getServiceTmpDir(), "created")
}

func getReleasedTmpDirs() string {
	return filepath.Join(getServiceTmpDir(), "released")
}

func getContextTmpDir() string {
	return filepath.Join(getServiceTmpDir(), "context")
}

func TempFile(pattern string) (f *os.File, err error) {
	return os.CreateTemp(werf.GetTmpDir(), pattern)
}

func newTmpDir(prefix string) (string, error) {
	newDir, err := os.MkdirTemp(werf.GetTmpDir(), prefix)
	if err != nil {
		return "", err
	}

	return newDir, nil
}

func newTmpFile(prefix string) (string, error) {
	newFile, err := TempFile(prefix)
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
