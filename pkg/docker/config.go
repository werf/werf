package docker

import (
	"fmt"
	"path/filepath"

	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/types"
	"github.com/docker/docker/pkg/homedir"
)

func GetDockerConfigCredentialsFile(configDir string) string {
	if configDir == "" {
		return filepath.Join(homedir.Get(), ".docker", config.ConfigFileName)
	} else {
		return filepath.Join(configDir, config.ConfigFileName)
	}
}

// StoreCredentials
//
// Inspired with https://github.com/google/go-containerregistry/blob/v0.20.3/cmd/crane/cmd/auth.go#L242
func StoreCredentials(configDir string, authConf types.AuthConfig) error {
	conf, err := config.Load(configDir)
	if err != nil {
		return fmt.Errorf("unable to load %s: %w", GetDockerConfigCredentialsFile(configDir), err)
	}
	creds := conf.GetCredentialsStore(authConf.ServerAddress)
	err = creds.Store(authConf)
	if err != nil {
		return fmt.Errorf("unable to store credentials: %w", err)
	}
	return nil
}

// EraseCredentials
//
// Inspired with https://github.com/google/go-containerregistry/blob/v0.20.3/cmd/crane/cmd/auth.go#L279
func EraseCredentials(configDir string, authConf types.AuthConfig) error {
	conf, err := config.Load(configDir)
	if err != nil {
		return fmt.Errorf("unable to load %s: %w", GetDockerConfigCredentialsFile(configDir), err)
	}
	creds := conf.GetCredentialsStore(authConf.ServerAddress)
	err = creds.Erase(authConf.ServerAddress)
	if err != nil {
		return fmt.Errorf("unable to erase credentials: %w", err)
	}
	return nil
}
