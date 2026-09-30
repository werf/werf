package docker

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/moby/moby/api/types/system"
	"github.com/moby/moby/client"
	"github.com/moby/moby/client/pkg/versions"
)

// daemonPingTimeout mirrors the init timeout of the docker cli (cli/command.defaultInitTimeout),
// with more headroom for remote daemons: without it a socket which accepts the connection but
// never answers hangs the command forever.
var daemonPingTimeout = 10 * time.Second

type daemonPing struct {
	apiVersion string
	err        error
	timedOut   bool
}

// checkedDaemons holds the API version of every daemon whose ping already succeeded, keyed by the
// api client. Only the version is cached: it does not change for the lifetime of the client, while
// the daemon being unreachable does — a daemon which starts up later must be pinged again.
var checkedDaemons sync.Map

type CheckConnectionOptions struct {
	AllowDaemonUnavailable bool
}

func CheckConnection(ctx context.Context, opts CheckConnectionOptions) error {
	api := apiCli(ctx)

	cached, ok := checkedDaemons.Load(api)
	if !ok {
		result := pingDaemon(ctx, api)
		switch {
		case result.timedOut:
			return fmt.Errorf("check Docker daemon API: %w", errDaemonDidNotAnswer())
		case result.err != nil:
			if opts.AllowDaemonUnavailable && ctx.Err() == nil && isDaemonUnavailableErr(result.err) {
				return nil
			}
			return fmt.Errorf("check Docker daemon API: %w", result.err)
		case result.apiVersion == "":
			return fmt.Errorf("Docker daemon did not report an API version; minimum supported API version is %s", client.MinAPIVersion)
		}
		checkedDaemons.Store(api, result.apiVersion)
		cached = result.apiVersion
	}

	if apiVersion := cached.(string); versions.LessThan(apiVersion, client.MinAPIVersion) {
		return fmt.Errorf("Docker daemon API version %s is unsupported: minimum supported API version is %s", apiVersion, client.MinAPIVersion)
	}

	return nil
}

func pingDaemon(ctx context.Context, api client.APIClient) daemonPing {
	var ping client.PingResult
	timedOut, err := callDaemon(ctx, func(callCtx context.Context) error {
		var err error
		ping, err = api.Ping(callCtx, client.PingOptions{NegotiateAPIVersion: true})
		return err
	})
	switch {
	case timedOut:
		return daemonPing{timedOut: true}
	case err != nil:
		return daemonPing{err: err}
	default:
		return daemonPing{apiVersion: ping.APIVersion}
	}
}

// callDaemon runs an Engine API call under the daemon deadline. It reports whether the call gave
// up on a daemon which never answered, as opposed to ctx itself being canceled.
func callDaemon(ctx context.Context, call func(ctx context.Context) error) (bool, error) {
	callCtx, cancel := context.WithTimeout(ctx, daemonPingTimeout)
	defer cancel()

	err := call(callCtx)
	if err != nil && ctx.Err() == nil && errors.Is(callCtx.Err(), context.DeadlineExceeded) {
		return true, err
	}

	return false, err
}

func errDaemonDidNotAnswer() error {
	return fmt.Errorf("Docker daemon did not answer within %s", daemonPingTimeout)
}

func Info(ctx context.Context) (system.Info, error) {
	api := apiCli(ctx)

	var result client.SystemInfoResult
	timedOut, err := callDaemon(ctx, func(callCtx context.Context) error {
		var err error
		result, err = api.Info(callCtx, client.InfoOptions{})
		return err
	})
	if timedOut {
		return system.Info{}, errDaemonDidNotAnswer()
	}
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
	var api client.APIClient
	switch {
	case IsContext(ctx):
		api = apiCli(ctx)
	case IsEnabled() && defaultAPIClient != nil:
		api = defaultAPIClient
	default:
		return nil, nil
	}

	var result client.SystemInfoResult
	timedOut, err := callDaemon(ctx, func(callCtx context.Context) error {
		var err error
		result, err = api.Info(callCtx, client.InfoOptions{})
		return err
	})
	if timedOut {
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
