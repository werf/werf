package common

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"

	cliconfig "github.com/docker/cli/cli/config"
	"github.com/google/go-containerregistry/pkg/registry"
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/werf/werf/v3/pkg/docker"
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

type writeRecordingRegistry struct {
	handler http.Handler

	mu     sync.Mutex
	writes []string
}

var _ http.Handler = (*writeRecordingRegistry)(nil)

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

func clearListEnv() {
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(key, "WERF_") {
			ginkgo.GinkgoT().Setenv(key, "")
			gomega.Expect(os.Unsetenv(key)).To(gomega.Succeed())
		}
	}
}

func registryMirrorContext() context.Context {
	oldDockerConfigDir, oldCLIConfigDir := docker.DockerConfigDir, cliconfig.Dir()
	ginkgo.DeferCleanup(func() {
		docker.DockerConfigDir = oldDockerConfigDir
		cliconfig.SetDir(oldCLIConfigDir)
	})
	for _, key := range []string{"DOCKER_HOST", "DOCKER_CONTEXT", "DOCKER_CONFIG", "DOCKER_CERT_PATH", "DOCKER_TLS", "DOCKER_TLS_VERIFY", "DOCKER_API_VERSION"} {
		ginkgo.GinkgoT().Setenv(key, "")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer ginkgo.GinkgoRecover()
		w.Header().Set("API-Version", "1.47")
		w.Header().Set("OSType", "linux")
		if strings.HasSuffix(r.URL.Path, "/info") {
			w.Header().Set("Content-Type", "application/json")
			_, err := io.WriteString(w, `{"RegistryConfig":{"Mirrors":["https://backend.example","https://bare.example"]}}`)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}
	}))
	ginkgo.DeferCleanup(server.Close)
	ginkgo.GinkgoT().Setenv("DOCKER_HOST", "tcp://"+server.Listener.Addr().String())
	gomega.Expect(docker.InitDockerConfig(docker.InitOptions{DockerConfigDir: ginkgo.GinkgoT().TempDir()})).To(gomega.Succeed())
	ctx, err := docker.NewContext(context.Background())
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return ctx
}
