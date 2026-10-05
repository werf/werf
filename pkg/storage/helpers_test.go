package storage

import (
	"bytes"
	"context"
	"io"
	"sync"

	"github.com/onsi/ginkgo/v2"

	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/image"
)

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
	onList  func(listing int)
	mu      sync.Mutex
	calls   int
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
	ginkgo.DeferCleanup(func() {
		for _, release := range releases {
			closeIfOpen(release)
		}
	})
	return started, releases
}

func blockNextListing(backend *localImageListBackendStub) (chan struct{}, chan struct{}) {
	listing, release := make(chan struct{}), make(chan struct{})
	backend.onList = func(_ int) {
		closeIfOpen(listing)
		<-release
	}
	ginkgo.DeferCleanup(func() { closeIfOpen(release) })
	return listing, release
}

func closeIfOpen(ch chan struct{}) {
	select {
	case <-ch:
	default:
		close(ch)
	}
}

var _ container_backend.ContainerBackend = (*localPublishBackendStub)(nil)

type localPublishBackendStub struct {
	*localImageListBackendStub
	tagImageErr error
}

func newLocalPublishBackendStub(images image.ImagesList) *localPublishBackendStub {
	return &localPublishBackendStub{localImageListBackendStub: &localImageListBackendStub{images: images}}
}

func (backend *localPublishBackendStub) TagImageByName(_ context.Context, _ container_backend.LegacyImageInterface) error {
	return backend.tagImageErr
}

func (backend *localPublishBackendStub) SaveImageToStream(_ context.Context, _ string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(newTinyDockerSaveTar())), nil
}

func (backend *localPublishBackendStub) LoadImageFromStream(_ context.Context, input io.Reader) (string, error) {
	if _, err := io.Copy(io.Discard, input); err != nil {
		return "", err
	}
	return "sha256:mutated", nil
}

var _ container_backend.LegacyImageInterface = (*localStageImageStub)(nil)

type localStageImageStub struct {
	container_backend.LegacyImageInterface
	name    string
	builtID string
}

func (img *localStageImageStub) Name() string { return img.name }

func (img *localStageImageStub) GetTargetPlatform() string { return "" }

func (img *localStageImageStub) SetBuiltID(builtID string) { img.builtID = builtID }
