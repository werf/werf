package common

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/google/go-containerregistry/pkg/registry"
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/docker_registry"
	"github.com/werf/werf/v3/pkg/storage"
)

func synchronizationTestStorage(ctx context.Context) *storage.RepoStagesStorage {
	server := httptest.NewServer(registry.New())
	ginkgo.DeferCleanup(server.Close)
	address := strings.TrimPrefix(server.URL, "http://") + "/test/repo"
	client, err := docker_registry.NewDockerRegistry(ctx, address, docker_registry.DefaultImplementationName, docker_registry.DockerRegistryOptions{InsecureRegistry: true})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return storage.NewRepoStagesStorage(&storage.NewRepoStagesStorageOptions{RepoAddress: address, DockerRegistry: client})
}

type synchronizationDNSFailureTransport struct{ next http.RoundTripper }

var _ http.RoundTripper = (*synchronizationDNSFailureTransport)(nil)

func (transport *synchronizationDNSFailureTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.URL.Hostname() == "synchronization.werf.io" {
		return nil, &net.DNSError{Err: "no such host", Name: request.URL.Hostname(), IsNotFound: true}
	}
	return transport.next.RoundTrip(request)
}
