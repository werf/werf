package tmp_manager

import (
	"os"
	"path/filepath"
	"time"

	"github.com/werf/werf/v3/pkg/werf"
)

const (
	projectsServiceDir          = "projects"
	dockerConfigsServiceDir     = "docker_configs"
	kubeConfigsServiceDir       = "kubeconfigs"
	werfConfigRendersServiceDir = "werf_config_renders"
	contextArchivesDir          = "context"
	contextPinsServiceDir       = "context_pins"

	// contextPinMaxAge is the age past which an unregistered context pin dir is considered orphaned:
	// a build that takes longer than that loses its pin.
	contextPinMaxAge = time.Hour * 24
)

var (
	commonPrefix           = "werf-" + werf.Version + "-"
	contextArchivePrefix   = commonPrefix + "context-"
	projectDirPrefix       = commonPrefix + "project-data-"
	dockerConfigDirPrefix  = commonPrefix + "docker-config-"
	kubeConfigDirPrefix    = commonPrefix + "kubeconfig-"
	werfConfigRenderPrefix = commonPrefix + "config-render-"
	contextPinDirPrefix    = commonPrefix + "context-pin-"
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
