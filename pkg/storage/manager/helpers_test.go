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

	"github.com/werf/werf/v3/pkg/docker_registry"
	"github.com/werf/werf/v3/pkg/storage"
)

func newTagsListStorageManager(tagsListStatus int) (*StorageManager, *atomic.Int32) {
	var tagsListRequests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer ginkgo.GinkgoRecover()

		if request.URL.Path == "/v2/" {
			writer.WriteHeader(http.StatusOK)
			return
		}

		gomega.Expect(request.URL.Path).To(gomega.HaveSuffix("/tags/list"))
		tagsListRequests.Add(1)

		if tagsListStatus != http.StatusOK {
			writer.WriteHeader(tagsListStatus)
			return
		}

		writer.WriteHeader(http.StatusOK)
		_, err := fmt.Fprint(writer, `{"name":"project/werf","tags":[]}`)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}))
	ginkgo.DeferCleanup(server.Close)

	repoAddress := strings.TrimPrefix(server.URL, "http://") + "/project/werf"
	dockerRegistry, err := docker_registry.NewDockerRegistry(context.Background(), repoAddress, docker_registry.DefaultImplementationName, docker_registry.DockerRegistryOptions{InsecureRegistry: true})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return &StorageManager{
		ProjectName: "test-project",
		StagesStorage: storage.NewRepoStagesStorage(&storage.NewRepoStagesStorageOptions{
			RepoAddress:    repoAddress,
			DockerRegistry: dockerRegistry,
		}),
	}, &tagsListRequests
}

func newBlockedTagsStorageManager(ctx context.Context) (*StorageManager, <-chan struct{}, func()) {
	return newTagsStorageManagerBlockingListing(ctx, 1)
}

// newTagsStorageManagerBlockingListing serves an empty tags listing for every request but holds the
// blockedListing-th one until the returned release is called, so that a listing started before a
// publication can still be in flight when the publication is already visible to a new one.
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
