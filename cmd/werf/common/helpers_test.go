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
	"github.com/werf/werf/v3/pkg/image"
	"github.com/werf/werf/v3/pkg/storage"
	"github.com/werf/werf/v3/pkg/storage/synchronization/lock_manager"
)

func synchronizationTestStorage(ctx context.Context) *storage.RepoStagesStorage {
	server := httptest.NewServer(registry.New())
	ginkgo.DeferCleanup(server.Close)
	address := strings.TrimPrefix(server.URL, "http://") + "/test/repo"
	client, err := docker_registry.NewDockerRegistry(ctx, address, docker_registry.DefaultImplementationName, docker_registry.DockerRegistryOptions{InsecureRegistry: true})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return storage.NewRepoStagesStorage(&storage.NewRepoStagesStorageOptions{RepoAddress: address, DockerRegistry: client})
}

type synchronizationTestLocker struct {
	acquireErr error
	releaseErr error
	acquired   int
	released   []lock_manager.LockHandle
	onAcquire  func()
}

var _ lock_manager.Interface = (*synchronizationTestLocker)(nil)

func (locker *synchronizationTestLocker) LockStage(_ context.Context, project, digest string) (lock_manager.LockHandle, error) {
	locker.acquired++
	if locker.onAcquire != nil {
		locker.onAcquire()
	}
	return lock_manager.LockHandle{ProjectName: project + "/" + digest}, locker.acquireErr
}

func (locker *synchronizationTestLocker) Unlock(_ context.Context, handle lock_manager.LockHandle) error {
	locker.released = append(locker.released, handle)
	return locker.releaseErr
}

type synchronizationErrorRegistry struct {
	docker_registry.Interface
	readErr  error
	writeErr error
}

var _ docker_registry.Interface = (*synchronizationErrorRegistry)(nil)

func (reg *synchronizationErrorRegistry) TryGetRepoImage(ctx context.Context, ref string) (*image.Info, error) {
	if reg.readErr != nil {
		return nil, reg.readErr
	}
	return reg.Interface.TryGetRepoImage(ctx, ref)
}

func (reg *synchronizationErrorRegistry) PushImage(ctx context.Context, ref string, opts *docker_registry.PushImageOptions) error {
	if reg.writeErr != nil {
		return reg.writeErr
	}
	return reg.Interface.PushImage(ctx, ref, opts)
}

type synchronizationDNSFailureTransport struct{ next http.RoundTripper }

var _ http.RoundTripper = (*synchronizationDNSFailureTransport)(nil)

func (transport *synchronizationDNSFailureTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.URL.Hostname() == "synchronization.werf.io" {
		return nil, &net.DNSError{Err: "no such host", Name: request.URL.Hostname(), IsNotFound: true}
	}
	return transport.next.RoundTrip(request)
}
