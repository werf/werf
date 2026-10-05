package manager

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/storage"
)

type lookupBackend struct {
	container_backend.ContainerBackend
	images image.ImagesList
	info   *image.Info
}

var _ container_backend.ContainerBackend = (*lookupBackend)(nil)

func (backend *lookupBackend) Images(_ context.Context, _ container_backend.ImagesOptions) (image.ImagesList, error) {
	return backend.images, nil
}

func (backend *lookupBackend) GetImageInfo(_ context.Context, _ string, _ container_backend.GetImageInfoOpts) (*image.Info, error) {
	return backend.info, nil
}

func newBlockedTagsStorageManager(ctx context.Context) (*StorageManager, <-chan struct{}, func()) {
	return newTagsStorageManagerBlockingListing(ctx, 1)
}

// Hold an empty pre-publication snapshot until release, while subsequent registry requests
// can observe a newly published image.
func newTagsStorageManagerBlockingListing(ctx context.Context, blockedListing int32) (*StorageManager, <-chan struct{}, func()) {
	listingStarted, release := make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	releaseListing := func() { releaseOnce.Do(func() { close(release) }) }
	var listings atomic.Int32
	backend := registry.New()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer ginkgo.GinkgoRecover()
		if strings.HasSuffix(request.URL.Path, "/tags/list") && listings.Add(1) == blockedListing {
			close(listingStarted)
			<-release
			_, err := fmt.Fprint(writer, `{"name":"project/werf","tags":[]}`)
			gomega.Expect(err).To(gomega.Succeed())
			return
		}
		backend.ServeHTTP(writer, request)
	}))
	ginkgo.DeferCleanup(server.Close)
	repoAddress := strings.TrimPrefix(server.URL, "http://") + "/project/werf"
	dockerRegistry, err := docker_registry.NewDockerRegistry(ctx, repoAddress, docker_registry.DefaultImplementationName, docker_registry.DockerRegistryOptions{InsecureRegistry: true})
	gomega.Expect(err).To(gomega.Succeed())
	return &StorageManager{
		ProjectName:   "test-project",
		StagesStorage: storage.NewRepoStagesStorage(&storage.NewRepoStagesStorageOptions{RepoAddress: repoAddress, DockerRegistry: dockerRegistry}),
	}, listingStarted, releaseListing
}
