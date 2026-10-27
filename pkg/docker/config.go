package docker

import (
	"fmt"
	"path/filepath"

	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/credentials"
	"github.com/docker/cli/cli/config/types"
	"github.com/docker/docker/pkg/homedir"
	"github.com/google/go-containerregistry/pkg/authn"
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
	originalAddress := authConf.ServerAddress
	authConf.ServerAddress = credentialsServerAddress(authConf.ServerAddress)
	conf, err := config.Load(configDir)
	if err != nil {
		return fmt.Errorf("unable to load %s: %w", GetDockerConfigCredentialsFile(configDir), err)
	}
	if authConf.ServerAddress == authn.DefaultAuthKey && conf.CredentialHelpers[authn.DefaultAuthKey] == "" {
		for _, address := range []string{originalAddress, "index.docker.io", "docker.io"} {
			if helper := conf.CredentialHelpers[address]; helper != "" {
				conf.CredentialHelpers[authn.DefaultAuthKey] = helper
				if err := conf.Save(); err != nil {
					return fmt.Errorf("unable to save Docker Hub credential helper: %w", err)
				}
				break
			}
		}
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
	serverAddress := credentialsServerAddress(authConf.ServerAddress)
	addresses := []string{serverAddress}
	if serverAddress == authn.DefaultAuthKey {
		addresses = append(addresses, "index.docker.io", "docker.io")
	}
	for _, address := range addresses {
		creds := conf.GetCredentialsStore(address)
		if serverAddress == authn.DefaultAuthKey {
			// Docker CLI normalizes Hub aliases; select the legacy stores by their original keys.
			helper := conf.CredentialHelpers[address]
			if helper == "" {
				helper = conf.CredentialsStore
			}
			creds = credentials.NewFileStore(conf)
			if helper != "" {
				creds = credentials.NewNativeStore(conf, helper)
			}
			stored, err := creds.Get(address)
			if err != nil {
				return fmt.Errorf("unable to get credentials for %q: %w", address, err)
			}
			stored.ServerAddress = ""
			if stored == (types.AuthConfig{}) {
				creds = credentials.NewFileStore(conf)
			}
		}
		if err := creds.Erase(address); err != nil {
			return fmt.Errorf("unable to erase credentials: %w", err)
		}
	}
	return nil
}

func credentialsServerAddress(address string) string {
	switch address {
	case "docker.io", "index.docker.io":
		return authn.DefaultAuthKey
	default:
		return address
	}
}
