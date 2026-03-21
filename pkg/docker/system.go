package docker

import (
	"context"
	"net"
	"strings"

	"github.com/docker/docker/api/types"
	dockerclient "github.com/docker/docker/client"
)

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
			cidrStr := (*net.IPNet)(cidr).String()
			if !seen[cidrStr] {
				seen[cidrStr] = true
				result = append(result, cidrStr)
			}
		}

		return result, nil
	}

	return nil, nil
}
