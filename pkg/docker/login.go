package docker

import (
	"context"
	"fmt"

	configTypes "github.com/docker/cli/cli/config/types"
	registryTypes "github.com/docker/docker/api/types/registry"

	"github.com/werf/logboek"
)

func Login(ctx context.Context, username, password, repo string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}
	if repo == "" {
		return fmt.Errorf("registry address cannot be empty")
	}

	resp, err := apiCli(ctx).RegistryLogin(ctx, registryTypes.AuthConfig{
		Username:      username,
		Password:      password,
		ServerAddress: repo,
	})
	if err != nil {
		return fmt.Errorf("unable to authenticate into %q: %w", repo, err)
	}

	logboek.Context(ctx).Debug().LogF("Docker login successful: %s\n", resp.Status)

	authConfig := configTypes.AuthConfig{
		ServerAddress: repo,
		Username:      username,
		Password:      password,
	}

	if resp.IdentityToken != "" {
		authConfig.IdentityToken = resp.IdentityToken
	}

	if err := StoreCredentials(DockerConfigDir, authConfig); err != nil {
		return fmt.Errorf("unable to store credentials: %w", err)
	}

	return nil
}
