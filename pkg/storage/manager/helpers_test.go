package manager

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"

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
