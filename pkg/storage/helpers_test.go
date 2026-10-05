package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/container_backend"
	"github.com/werf/werf/v3/pkg/docker_registry"
	registry_api "github.com/werf/werf/v3/pkg/docker_registry/api"
	"github.com/werf/werf/v3/pkg/image"
	"github.com/werf/werf/v3/pkg/opstats"
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
	name    string
	images  image.ImagesList
	err     error
	options container_backend.ImagesOptions
	onList  func(listing int)
	mu      sync.Mutex
	calls   int
}

func (backend *localImageListBackendStub) String() string {
	if backend.name == "" {
		return "docker-server-backend"
	}
	return backend.name
}

func (backend *localImageListBackendStub) Images(_ context.Context, options container_backend.ImagesOptions) (image.ImagesList, error) {
	backend.mu.Lock()
	defer backend.mu.Unlock()
	backend.calls++
	backend.options = options
	if backend.onList != nil {
		backend.onList(backend.calls)
	}
	return backend.images, backend.err
}

func (backend *localImageListBackendStub) callCount() int {
	backend.mu.Lock()
	defer backend.mu.Unlock()
	return backend.calls
}

// blockedCallTimeout bounds the wait for a call that must not be blocked by a listing in flight,
// so that a regression fails the spec instead of hanging it.
const blockedCallTimeout = "10s"

var _ context.Context = (*listingRegistrationContext)(nil)

type listingRegistrationContext struct {
	context.Context
	registered chan struct{}
}

// Done reports every wait for the cancellation channel. waitProjectListing settles the role of the
// caller under the flight mutex before it selects on ctx.Done, so a report proves this caller joined
// the listing that is in flight right now, which no absence of its result can prove.
func (ctx *listingRegistrationContext) Done() <-chan struct{} {
	ctx.registered <- struct{}{}
	return ctx.Context.Done()
}

func listingRegistrations(ctx context.Context) (context.Context, chan struct{}) {
	registered := make(chan struct{}, 8)
	return &listingRegistrationContext{Context: ctx, registered: registered}, registered
}

// blockListings makes each of the next count listings report its number on the returned channel and
// then hold until the matching release channel is closed.
func blockListings(backend *localImageListBackendStub, count int) (chan int, []chan struct{}) {
	started := make(chan int, count)
	releases := make([]chan struct{}, count)
	for i := range releases {
		releases[i] = make(chan struct{})
	}
	backend.onList = func(listing int) {
		started <- listing
		if listing <= len(releases) {
			<-releases[listing-1]
		}
	}
	DeferCleanup(func() {
		for _, release := range releases {
			closeIfOpen(release)
		}
	})
	return started, releases
}

// blockNextListing makes the next image listing report that it started and then hold until the
// returned release channel is closed.
func blockNextListing(backend *localImageListBackendStub) (chan struct{}, chan struct{}) {
	listing, release := make(chan struct{}), make(chan struct{})
	backend.onList = func(_ int) {
		closeIfOpen(listing)
		<-release
	}
	DeferCleanup(func() { closeIfOpen(release) })
	return listing, release
}

func closeIfOpen(ch chan struct{}) {
	select {
	case <-ch:
	default:
		close(ch)
	}
}

var (
	_ container_backend.ContainerBackend    = (*localPublishBackendStub)(nil)
	_ container_backend.NativeConfigMutator = (*localPublishBackendStub)(nil)
)

type localPublishBackendStub struct {
	*localImageListBackendStub
	tagImageErr error
	nativeErr   error
	tagErr      error
	tagged      []string
}

func newLocalPublishBackendStub(images image.ImagesList) *localPublishBackendStub {
	return &localPublishBackendStub{localImageListBackendStub: &localImageListBackendStub{images: images}}
}

func (backend *localPublishBackendStub) TagImageByName(_ context.Context, img container_backend.LegacyImageInterface) error {
	if backend.tagImageErr != nil {
		return backend.tagImageErr
	}
	backend.tagged = append(backend.tagged, img.Name())
	return nil
}

func (backend *localPublishBackendStub) MutateAndPushImageNative(_ context.Context, _, dest string, _ image.SpecConfig, _ string) error {
	if backend.nativeErr != nil {
		return backend.nativeErr
	}
	backend.tagged = append(backend.tagged, dest)
	return nil
}

func (backend *localPublishBackendStub) GetImageConfigFile(_ context.Context, _ string) (*v1.ConfigFile, error) {
	return &v1.ConfigFile{}, nil
}

func (backend *localPublishBackendStub) LoadImageFromStream(_ context.Context, input io.Reader) (string, error) {
	if _, err := io.Copy(io.Discard, input); err != nil {
		return "", err
	}
	return "sha256:mutated", nil
}

func (backend *localPublishBackendStub) Tag(_ context.Context, _, dest string, _ container_backend.TagOpts) error {
	if backend.tagErr != nil {
		return backend.tagErr
	}
	backend.tagged = append(backend.tagged, dest)
	return nil
}

var _ container_backend.LegacyImageInterface = (*localStageImageStub)(nil)

type localStageImageStub struct {
	container_backend.LegacyImageInterface
	name string
}

func (img *localStageImageStub) Name() string { return img.name }

func (img *localStageImageStub) GetTargetPlatform() string { return "" }

var _ docker_registry.Interface = (*brokenStageRegistry)(nil)

var _ container_backend.ContainerBackend = (*brokenStageBackend)(nil)

type brokenStageRegistry struct {
	*markerRegistry
	err error
}

func (r *brokenStageRegistry) GetRepoImage(_ context.Context, _ string) (*image.Info, error) {
	return nil, r.err
}

func (r *brokenStageRegistry) MutateAndPushImage(_ context.Context, _, _ string, _ ...registry_api.MutateOption) error {
	return r.err
}

type brokenStageBackend struct {
	container_backend.ContainerBackend
	err error
}

func (b *brokenStageBackend) PullImageFromRegistry(_ context.Context, _ container_backend.LegacyImageInterface) error {
	return b.err
}

func brokenCount(collector *opstats.Collector) int {
	for _, e := range collector.EventSummary() {
		if e.Event == opstats.EventStageBroken {
			return e.Count
		}
	}
	return 0
}

func collectingContext(ctx context.Context) (context.Context, *opstats.Collector) {
	collector := opstats.NewCollector()
	return opstats.NewContext(ctx, collector), collector
}
