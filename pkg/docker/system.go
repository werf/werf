package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/moby/moby/api/types/system"
	"github.com/moby/moby/client"
	"github.com/moby/moby/client/pkg/versions"
)

type CheckConnectionOptions struct {
	AllowDaemonUnavailable bool
}

func CheckConnection(ctx context.Context, opts CheckConnectionOptions) error {
	ping, err := apiCli(ctx).Ping(ctx, client.PingOptions{NegotiateAPIVersion: true})
	if err != nil {
		if opts.AllowDaemonUnavailable && ctx.Err() == nil && isDaemonUnavailableErr(err) {
			return nil
		}
		return fmt.Errorf("check Docker daemon API: %w", err)
	}
	if ping.APIVersion == "" {
		return fmt.Errorf("Docker daemon did not report an API version; minimum supported API version is %s", client.MinAPIVersion)
	}
	if versions.LessThan(ping.APIVersion, client.MinAPIVersion) {
		return fmt.Errorf("Docker daemon API version %s is unsupported: minimum supported API version is %s", ping.APIVersion, client.MinAPIVersion)
	}
	return nil
}

func Info(ctx context.Context) (system.Info, error) {
	result, err := apiCli(ctx).Info(ctx, client.InfoOptions{})
	if err != nil {
		return system.Info{}, err
	}

	return result.Info, nil
}

func isDaemonUnavailableErr(err error) bool {
	if err == nil {
		return false
	}
	if client.IsErrConnectionFailed(err) {
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

func getDaemonInfo(ctx context.Context) (*system.Info, error) {
	var result client.SystemInfoResult
	var err error

	if IsContext(ctx) {
		result, err = apiCli(ctx).Info(ctx, client.InfoOptions{})
	} else if IsEnabled() && defaultAPIClient != nil {
		result, err = defaultAPIClient.Info(ctx, client.InfoOptions{})
	} else {
		return nil, nil
	}

	if err != nil {
		if isDaemonUnavailableErr(err) {
			return nil, nil
		}
		return nil, err
	}

	return &result.Info, nil
}

func GetRegistryMirrors(ctx context.Context) ([]string, error) {
	info, err := getDaemonInfo(ctx)
	if err != nil {
		return nil, err
	}

	if info != nil && info.RegistryConfig != nil {
		return info.RegistryConfig.Mirrors, nil
	}

	return nil, nil
}

func GetInsecureRegistries(ctx context.Context) ([]string, error) {
	info, err := getDaemonInfo(ctx)
	if err != nil {
		return nil, err
	}

	if info != nil && info.RegistryConfig != nil {
		var result []string
		seen := make(map[string]bool)

		for host, indexInfo := range info.RegistryConfig.IndexConfigs {
			if !indexInfo.Secure && !seen[host] {
				seen[host] = true
				result = append(result, host)
			}
		}

		for _, cidr := range info.RegistryConfig.InsecureRegistryCIDRs {
			cidrStr := cidr.String()
			if !seen[cidrStr] {
				seen[cidrStr] = true
				result = append(result, cidrStr)
			}
		}

		return result, nil
	}

	return nil, nil
}
