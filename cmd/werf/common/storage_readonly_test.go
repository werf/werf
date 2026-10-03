package common

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"

	"github.com/google/go-containerregistry/pkg/registry"
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/werf/werf/v3/pkg/docker_registry"
	"github.com/werf/werf/v3/pkg/storage"
	"github.com/werf/werf/v3/pkg/werf"
)

// writeRecordingRegistry fronts a real in-memory registry and records every
// request that could modify it, so a test can assert that an initialization
// path is read-only.
type writeRecordingRegistry struct {
	handler http.Handler

	mu     sync.Mutex
	writes []string
}

func (r *writeRecordingRegistry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		r.mu.Lock()
		r.writes = append(r.writes, req.Method+" "+req.URL.Path)
		r.mu.Unlock()
	}
	r.handler.ServeHTTP(w, req)
}

func (r *writeRecordingRegistry) recordedWrites() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.writes...)
}

func (r *writeRecordingRegistry) forgetWrites() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.writes = nil
}

// startRecordingRegistry returns the repo address of a fresh in-memory registry
// together with its write recorder.
func startRecordingRegistry(repo string) (string, *writeRecordingRegistry) {
	recorder := &writeRecordingRegistry{handler: registry.New()}
	server := httptest.NewServer(recorder)
	ginkgo.DeferCleanup(server.Close)
	return strings.TrimPrefix(server.URL, "http://") + "/" + repo, recorder
}

func storageManagerCmdData(args ...string) *CmdData {
	command := &cobra.Command{}
	data := &CmdData{}
	SetupSecondaryStagesStorageOptions(data, command)
	SetupCacheStagesStorageOptions(data, command)
	SetupRepoOptions(data, command, RepoDataOptions{OptionalRepo: true})
	SetupFinalRepo(data, command)
	SetupMetaRepo(data, command)
	SetupContainerRegistryMirror(data, command)
	SetupSynchronization(data, command)
	SetupCheckBuiltImages(data, command)
	gomega.Expect(command.ParseFlags(args)).To(gomega.Succeed())
	return data
}

func repoStagesStorageAt(ctx context.Context, address string) *storage.RepoStagesStorage {
	client, err := docker_registry.NewDockerRegistry(ctx, address, docker_registry.DefaultImplementationName, docker_registry.DockerRegistryOptions{InsecureRegistry: true})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return storage.NewRepoStagesStorage(&storage.NewRepoStagesStorageOptions{RepoAddress: address, DockerRegistry: client})
}

var _ = ginkgo.Describe("Storage manager initialization in check mode", func() {
	ginkgo.BeforeEach(func() {
		gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), ginkgo.GinkgoT().TempDir())).To(gomega.Succeed())
	})

	ginkgo.It("writes nothing into an empty repo and never contacts the sync server", func(ctx ginkgo.SpecContext) {
		address, recorder := startRecordingRegistry("test/repo")

		storageManager, err := NewStorageManager(ctx, &NewStorageManagerConfig{
			ProjectName: "project",
			CmdData: storageManagerCmdData(
				"--repo", address,
				"--insecure-registry",
				"--synchronization", "https://synchronization.invalid",
				"--check-built-images",
			),
		})

		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(storageManager.StorageLockManager).NotTo(gomega.BeNil())
		gomega.Expect(recorder.recordedWrites()).To(gomega.BeEmpty())
	})

	ginkgo.It("still validates the synchronization address", func(ctx ginkgo.SpecContext) {
		address, recorder := startRecordingRegistry("test/repo")

		_, err := NewStorageManager(ctx, &NewStorageManagerConfig{
			ProjectName: "project",
			CmdData: storageManagerCmdData(
				"--repo", address,
				"--insecure-registry",
				"--synchronization", "ftp://sync.example",
				"--check-built-images",
			),
		})

		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("unsupported synchronization address")))
		gomega.Expect(recorder.recordedWrites()).To(gomega.BeEmpty())
	})

	ginkgo.It("still rejects a repo whose meta-repo marker points elsewhere, without replanting it", func(ctx ginkgo.SpecContext) {
		address, recorder := startRecordingRegistry("test/repo")
		otherAddress, _ := startRecordingRegistry("test/meta")

		gomega.Expect(repoStagesStorageAt(ctx, address).PutMetaRepoMarker(ctx, "project", "registry.example/planted")).To(gomega.Succeed())
		recorder.forgetWrites()

		_, err := NewStorageManager(ctx, &NewStorageManagerConfig{
			ProjectName: "project",
			CmdData: storageManagerCmdData(
				"--repo", address,
				"--meta-repo", otherAddress,
				"--insecure-registry",
				"--synchronization", "https://synchronization.invalid",
				"--check-built-images",
			),
		})

		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("registry.example/planted")))
		gomega.Expect(recorder.recordedWrites()).To(gomega.BeEmpty())
	})

	ginkgo.It("registers the synchronization identity during ordinary initialization", func(ctx ginkgo.SpecContext) {
		address, recorder := startRecordingRegistry("test/repo")
		syncServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			defer ginkgo.GinkgoRecover()
			_, err := w.Write([]byte(`{"clientID":"test-client"}`))
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}))
		ginkgo.DeferCleanup(syncServer.Close)

		storageManager, err := NewStorageManager(ctx, &NewStorageManagerConfig{
			ProjectName: "project",
			CmdData: storageManagerCmdData(
				"--repo", address,
				"--insecure-registry",
				"--synchronization", syncServer.URL,
			),
		})

		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(storageManager.StorageLockManager).NotTo(gomega.BeNil())
		gomega.Expect(recorder.recordedWrites()).NotTo(gomega.BeEmpty())
	})
})
