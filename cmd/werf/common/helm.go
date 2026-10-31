package common

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/werf/logboek"
	nelmcommon "github.com/werf/nelm/v2/pkg/common"
	"github.com/werf/nelm/v2/pkg/helm/pkg/registry"
	"github.com/werf/werf/v3/pkg/docker"
)

func NewHelmRegistryClient(ctx context.Context, dockerConfig string, insecureHelmDependencies bool) (*registry.Client, error) {
	client, err := NewHelmRegistryClientWithoutInit(ctx)
	if err != nil {
		return nil, err
	}

	InitHelmRegistryClient(client, dockerConfig, insecureHelmDependencies)

	return client, nil
}

func NewHelmRegistryClientWithoutInit(ctx context.Context) (*registry.Client, error) {
	return registry.NewClient(
		registry.ClientOptDebug(logboek.Context(ctx).Debug().IsAccepted()),
		registry.ClientOptWriter(logboek.Context(ctx).OutStream()),
	)
}

func InitHelmRegistryClient(registryClient *registry.Client, dockerConfig string, insecureHelmDependencies bool) {
	registry.ClientOptCredentialsFile(docker.GetDockerConfigCredentialsFile(dockerConfig))(registryClient)
	if insecureHelmDependencies {
		registry.ClientOptPlainHTTP()
	}
}

// NewBundlesRegistryClient returns the client of the container registry the bundles are pushed
// into and pulled from, connected as GetBundlesRegistryConnectionOptions says: over plain HTTP if
// insecure or, as Docker does, on a loopback host, where verifying TLS certificates doesn't apply.
func NewBundlesRegistryClient(ctx context.Context, commonCmdData *CmdData) (*registry.Client, error) {
	clientOpts := []registry.ClientOption{
		registry.ClientOptDebug(*commonCmdData.LogDebug),
		registry.ClientOptEnableCache(true),
		registry.ClientOptWriter(logboek.Context(ctx).OutStream()),
		registry.ClientOptCredentialsFile(docker.GetDockerConfigCredentialsFile(*commonCmdData.DockerConfig)),
	}

	connectionOpts := GetBundlesRegistryConnectionOptions(commonCmdData)
	clientOpts = append(clientOpts, registry.ClientOptPlainHTTPHost(func(host string) bool {
		return connectionOpts.ChartRepoInsecure || isLoopbackHost(host)
	}))

	if connectionOpts.ChartRepoSkipTLSVerify {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

		clientOpts = append(clientOpts, registry.ClientOptHTTPClient(&http.Client{Transport: transport}))
	}

	client, err := registry.NewClient(clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("create registry client: %w", err)
	}

	return client, nil
}

// GetBundlesRegistryConnectionOptions returns the options to connect to the container registries of
// the bundles and their images with. Nelm accesses the bundle registries over plain HTTP if
// insecure, and other registries, e.g. of the images of a bundle copy, over HTTPS, skipping TLS
// verification if asked.
func GetBundlesRegistryConnectionOptions(commonCmdData *CmdData) nelmcommon.ChartRepoConnectionOptions {
	return nelmcommon.ChartRepoConnectionOptions{
		ChartRepoInsecure:      *commonCmdData.InsecureRegistry,
		ChartRepoSkipTLSVerify: *commonCmdData.SkipTlsVerifyRegistry,
	}
}

// isLoopbackHost returns whether the registry host (host or host:port) is a loopback host.
func isLoopbackHost(host string) bool {
	if hostname, _, err := net.SplitHostPort(host); err == nil {
		host = hostname
	}

	if host == "localhost" {
		return true
	}

	ip := net.ParseIP(strings.Trim(host, "[]"))

	return ip != nil && ip.IsLoopback()
}
