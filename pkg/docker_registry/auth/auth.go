package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v5"
	apiregistry "github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/registry"
)

type Options struct {
	Hostname  string
	Username  string
	Password  string
	Insecure  bool
	UserAgent string
}

// Auth
//
// Inspired with https://github.com/oras-project/oras-go/blob/v1/pkg/auth/docker/login.go#L52
func Auth(ctx context.Context, options Options) (string, error) {
	registryOptions := registry.ServiceOptions{}

	if options.Insecure {
		registryOptions.InsecureRegistries = []string{options.Hostname}
	}

	remote, err := registry.NewService(registryOptions)
	if err != nil {
		return "", fmt.Errorf("unable to initiatiate registry service: %w", err)
	}

	authConfig := &apiregistry.AuthConfig{
		Username:      options.Username,
		ServerAddress: options.Hostname,
	}

	if options.Username == "" {
		authConfig.IdentityToken = options.Password
	} else {
		authConfig.Password = options.Password
	}

	operation := func() (string, error) {
		_, token, err := remote.Auth(ctx, authConfig, options.UserAgent)
		if err != nil {
			if strings.Contains(err.Error(), "failed with status: 429") {
				return "", err
			}
			// Do not retry on other errors.
			return "", backoff.Permanent(err)
		}
		return token, nil
	}

	token, err := backoff.Retry(ctx, operation,
		backoff.WithBackOff(backoff.NewExponentialBackOff()),
		backoff.WithMaxElapsedTime(5*time.Minute)) // Maximum time for all retries.
	if err != nil {
		return "", err
	}

	return token, nil
}
