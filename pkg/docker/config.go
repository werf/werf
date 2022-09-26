package docker

import (
	"path/filepath"

	"github.com/docker/cli/cli/config"
	"github.com/docker/docker/pkg/homedir"
)

func GetDockerConfigCredentialsFile(configDir string) string {
	if configDir == "" {
		return filepath.Join(homedir.Get(), ".docker", config.ConfigFileName)
	} else {
		return filepath.Join(configDir, config.ConfigFileName)
	}
}
