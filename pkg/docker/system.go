package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/docker/docker/api/types"
	dockerclient "github.com/docker/docker/client"
)

type daemonConfig struct {
	RegistryMirrors    []string `json:"registry-mirrors"`
	InsecureRegistries []string `json:"insecure-registries"`
}

func getDaemonConfigPaths() []string {
	var paths []string

	if home := os.Getenv("HOME"); home != "" {
		paths = append(paths, filepath.Join(home, ".docker", "daemon.json"))
	}
	if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
		paths = append(paths, filepath.Join(userProfile, ".docker", "daemon.json"))
	}

	switch runtime.GOOS {
	case "linux":
		paths = append(paths, "/etc/docker/daemon.json")
	case "windows":
		if programData := os.Getenv("ProgramData"); programData != "" {
			paths = append(paths, filepath.Join(programData, "docker", "config", "daemon.json"))
		}
	}

	return paths
}

// readDaemonConfigFromFile reads daemon.json directly from filesystem.
// Used as a fallback when Docker daemon is not running.
func readDaemonConfigFromFile() (*daemonConfig, error) {
	for _, path := range getDaemonConfigPaths() {
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) || os.IsPermission(err) {
				// File doesn't exist or not readable, try next path
				continue
			}
			return nil, fmt.Errorf("unable to read docker config %q: %w", path, err)
		}

		var cfg daemonConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("unable to parse docker config %q: %w", path, err)
		}

		return &cfg, nil
	}

	return nil, nil
}

func Info(ctx context.Context) (types.Info, error) {
	return apiCli(ctx).Info(ctx)
}

func isDaemonUnavailableErr(err error) bool {
	if err == nil {
		return false
	}
	if dockerclient.IsErrConnectionFailed(err) {
		return true
	}
	msg := err.Error()
	for _, substr := range []string{
		"Cannot connect to the Docker daemon",
		"connect: no such file or directory",
		"connect: connection refused",
		"dial unix",
	} {
		if strings.Contains(msg, substr) {
			return true
		}
	}
	return false
}

func getDaemonInfo(ctx context.Context) (*types.Info, error) {
	var info types.Info
	var err error

	if IsContext(ctx) {
		info, err = apiCli(ctx).Info(ctx)
	} else if IsEnabled() && defaultCLI != nil {
		info, err = defaultCLI.Client().Info(ctx)
	} else {
		return nil, nil
	}

	if err != nil {
		if isDaemonUnavailableErr(err) {
			return nil, nil
		}
		return nil, err
	}

	return &info, nil
}

// TryGetDaemonInfo attempts to get Docker daemon info without requiring prior Init.
// Returns (nil, nil) when daemon is unavailable, allowing graceful fallback.
func TryGetDaemonInfo(ctx context.Context) (*types.Info, error) {
	if IsEnabled() || IsContext(ctx) {
		return getDaemonInfo(ctx)
	}

	c, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		// Cannot even construct the client – treat as unavailable.
		return nil, nil
	}
	defer c.Close()

	info, err := c.Info(ctx)
	if err != nil {
		if isDaemonUnavailableErr(err) {
			return nil, nil
		}
		return nil, err
	}
	return &info, nil
}

func registryInfoFromDaemon(ctx context.Context) (*types.Info, error) {
	if IsEnabled() || IsContext(ctx) {
		return getDaemonInfo(ctx)
	}
	return TryGetDaemonInfo(ctx)
}

// GetRegistryMirrors returns registry mirrors from Docker daemon API.
// Falls back to reading daemon.json if daemon is unavailable.
func GetRegistryMirrors(ctx context.Context) ([]string, error) {
	info, err := registryInfoFromDaemon(ctx)
	if err != nil {
		return nil, err
	}

	if info != nil && info.RegistryConfig != nil {
		return info.RegistryConfig.Mirrors, nil
	}

	cfg, err := readDaemonConfigFromFile()
	if err != nil {
		return nil, err
	}
	if cfg != nil {
		return cfg.RegistryMirrors, nil
	}

	return nil, nil
}

// GetInsecureRegistries returns insecure registries from Docker daemon API.
// Falls back to reading daemon.json if daemon is unavailable.
func GetInsecureRegistries(ctx context.Context) ([]string, error) {
	info, err := registryInfoFromDaemon(ctx)
	if err != nil {
		return nil, err
	}

	if info != nil && info.RegistryConfig != nil {
		var result []string
		seen := make(map[string]bool)

		for host, indexInfo := range info.RegistryConfig.IndexConfigs {
			if !indexInfo.Secure {
				if !seen[host] {
					seen[host] = true
					result = append(result, host)
				}
			}
		}

		for _, cidr := range info.RegistryConfig.InsecureRegistryCIDRs {
			cidrStr := (*net.IPNet)(cidr).String()
			if !seen[cidrStr] {
				seen[cidrStr] = true
				result = append(result, cidrStr)
			}
		}

		return result, nil
	}

	cfg, err := readDaemonConfigFromFile()
	if err != nil {
		return nil, err
	}
	if cfg != nil {
		return cfg.InsecureRegistries, nil
	}

	return nil, nil
}
