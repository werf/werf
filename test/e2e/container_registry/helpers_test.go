package e2e_container_registry_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/werf/werf/v3/pkg/docker_registry"
)

func dockerHubToken(ctx context.Context, client *http.Client, options docker_registry.DockerRegistryOptions) (string, error) {
	body, err := json.Marshal(map[string]string{"username": options.DockerHubUsername, "password": options.DockerHubPassword})
	if err != nil {
		return "", fmt.Errorf("encode Docker Hub credentials: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://hub.docker.com/v2/users/login/", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create Docker Hub login request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("log in to Docker Hub: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("log in to Docker Hub: HTTP %d", response.StatusCode)
	}
	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode Docker Hub login response: %w", err)
	}
	if result.Token == "" {
		return "", fmt.Errorf("log in to Docker Hub: missing token")
	}
	return result.Token, nil
}

func dockerHubRepositoryExists(ctx context.Context, client *http.Client, endpoint, token string) (bool, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return false, fmt.Errorf("create Docker Hub repository request: %w", err)
	}
	request.Header.Set("Authorization", "JWT "+token)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Cache-Control", "no-cache")
	response, err := client.Do(request)
	if err != nil {
		return false, fmt.Errorf("get Docker Hub repository: %w", err)
	}
	defer response.Body.Close()
	switch response.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("get Docker Hub repository: HTTP %d", response.StatusCode)
	}
}
