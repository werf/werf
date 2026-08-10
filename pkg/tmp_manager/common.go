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
)

var (
	commonPrefix           = "werf-" + werf.Version + "-"
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

func TempFile(pattern string) (f *os.File, err error) {
	return os.CreateTemp(werf.GetTmpDir(), pattern)
}

// TempDir creates a temporary directory named after the common werf prefix, so that a
// directory leaked by an interrupted command is swept by `werf host purge`.
func TempDir(pattern string) (string, error) {
	return os.MkdirTemp(werf.GetTmpDir(), commonPrefix+pattern)
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
