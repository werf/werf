package storage

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/container_backend"
	"github.com/werf/werf/v3/pkg/docker_registry"
	"github.com/werf/werf/v3/pkg/image"
)

const tagCacheStageDigest = "2222222222222222222222222222222222222222222222222222222c"

var _ container_backend.ContainerBackend = (*pushStageBackendStub)(nil)

type pushStageBackendStub struct {
	container_backend.ContainerBackend

	pushErr error
}

func (b *pushStageBackendStub) Push(_ context.Context, _ string, _ container_backend.PushOpts) error {
	return b.pushErr
}

// newTagCacheRepoStagesStorage returns a stages storage backed by an in-memory registry, plus the
// number of tags listings it served.
func newTagCacheRepoStagesStorage(ctx context.Context, backend container_backend.ContainerBackend) (*RepoStagesStorage, *atomic.Int32) {
	var tagsListRequests atomic.Int32

	registryHandler := registry.New()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if strings.HasSuffix(request.URL.Path, "/tags/list") {
			tagsListRequests.Add(1)
		}
		registryHandler.ServeHTTP(writer, request)
	}))
	DeferCleanup(server.Close)

	repoAddress := strings.TrimPrefix(server.URL, "http://") + "/project/werf"
	dockerRegistry, err := docker_registry.NewDockerRegistry(ctx, repoAddress, docker_registry.DefaultImplementationName, docker_registry.DockerRegistryOptions{InsecureRegistry: true})
	Expect(err).NotTo(HaveOccurred())

	return NewRepoStagesStorage(&NewRepoStagesStorageOptions{
		RepoAddress:      repoAddress,
		DockerRegistry:   dockerRegistry,
		ContainerBackend: backend,
	}), &tagsListRequests
}

func cachedStageIDs(ctx context.Context, storage *RepoStagesStorage) []string {
	stageIDs, err := storage.GetStagesIDsByDigest(ctx, "", tagCacheStageDigest, 0, WithCache())
	Expect(err).NotTo(HaveOccurred())
	return stageStrings(stageIDs)
}

func pushRandomImage(reference string) {
	ref, err := name.ParseReference(reference, name.Insecure)
	Expect(err).NotTo(HaveOccurred())
	img, err := random.Image(1, 1)
	Expect(err).NotTo(HaveOccurred())
	Expect(remote.Write(ref, img)).To(Succeed())
}

var _ docker_registry.Interface = (*metadataPushRegistry)(nil)

type metadataPushRegistry struct {
	*pushImageRegistryStub
}

func (r *metadataPushRegistry) Tags(_ context.Context, _ string, _ ...docker_registry.Option) ([]string, error) {
	return nil, fmt.Errorf("tag listing must not be needed to publish metadata")
}

var _ docker_registry.Interface = (*stageLookupRegistry)(nil)

type stageLookupRegistry struct {
	*markerRegistry
	brokenImage *image.Info
}

func (r *stageLookupRegistry) GetRepoImage(ctx context.Context, reference string) (*image.Info, error) {
	info, err := r.markerRegistry.GetRepoImage(ctx, reference)
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, fmt.Errorf("%s: %s", transport.ManifestUnknownErrorCode, reference)
	}
	if info == r.brokenImage {
		return nil, fmt.Errorf("%s: %s", transport.BlobUnknownErrorCode, reference)
	}
	return info, nil
}

var _ container_backend.ContainerBackend = (*stageLookupBackend)(nil)

type stageLookupBackend struct {
	container_backend.ContainerBackend
	info *image.Info
}

func (b *stageLookupBackend) GetImageInfo(_ context.Context, _ string, _ container_backend.GetImageInfoOpts) (*image.Info, error) {
	return b.info, nil
}

func stageStrings(stages []image.StageID) []string {
	result := make([]string, 0, len(stages))
	for _, stage := range stages {
		result = append(result, stage.String())
	}
	return result
}

var _ container_backend.ContainerBackend = (*localImageListBackendStub)(nil)

type localImageListBackendStub struct {
	container_backend.ContainerBackend
	images  image.ImagesList
	err     error
	options container_backend.ImagesOptions
	mu      sync.Mutex
	calls   int
}

func (backend *localImageListBackendStub) Images(_ context.Context, options container_backend.ImagesOptions) (image.ImagesList, error) {
	backend.mu.Lock()
	defer backend.mu.Unlock()
	backend.calls++
	backend.options = options
	return backend.images, backend.err
}

var _ docker_registry.Interface = (*synchronizationCompetingRegistry)(nil)

type synchronizationCompetingRegistry struct{ docker_registry.Interface }

func (registry *synchronizationCompetingRegistry) PushImage(ctx context.Context, reference string, opts *docker_registry.PushImageOptions) error {
	replacement := &docker_registry.PushImageOptions{Labels: map[string]string{image.WerfLabel: opts.Labels[image.WerfLabel], synchronizationMarkerFingerprintLabel: strings.Repeat("b", 64)}}
	return registry.Interface.PushImage(ctx, reference, replacement)
}
