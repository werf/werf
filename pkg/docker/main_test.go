package docker

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/cli/cli/command"
	cliconfig "github.com/docker/cli/cli/config"
	contextdocker "github.com/docker/cli/cli/context/docker"
	"github.com/docker/cli/cli/context/store"
	dockerimage "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Docker API connection", func() {
	ginkgo.BeforeEach(func() {
		oldCLI, oldAPI := defaultCLI, defaultAPIClient
		oldConfigDir, oldCliConfigDir := DockerConfigDir, cliconfig.Dir()
		oldDefaultPlatform, oldRuntimePlatform := defaultPlatform, runtimePlatform
		oldDebug, oldLiveOutput := isDebug, liveCliOutputEnabled
		ginkgo.DeferCleanup(func() {
			defaultCLI, defaultAPIClient = oldCLI, oldAPI
			DockerConfigDir = oldConfigDir
			cliconfig.SetDir(oldCliConfigDir)
			defaultPlatform, runtimePlatform = oldDefaultPlatform, oldRuntimePlatform
			isDebug, liveCliOutputEnabled = oldDebug, oldLiveOutput
		})
		for _, key := range []string{"DOCKER_HOST", "DOCKER_CONTEXT", "DOCKER_CONFIG", "DOCKER_CERT_PATH", "DOCKER_TLS", "DOCKER_TLS_VERIFY", "DOCKER_API_VERSION", "DOCKER_DEFAULT_PLATFORM", "WERF_DEBUG_DOCKER"} {
			ginkgo.GinkgoT().Setenv(key, "")
		}
	})

	ginkgo.DescribeTable("returns an error for a missing context", func(global bool) {
		configDir := ginkgo.GinkgoT().TempDir()
		gomega.Expect(os.WriteFile(filepath.Join(configDir, "config.json"), []byte(`{"currentContext":"missing"}`), 0o600)).To(gomega.Succeed())
		var err error
		if global {
			err = Init(context.Background(), InitOptions{DockerConfigDir: configDir})
		} else {
			gomega.Expect(InitDockerConfig(InitOptions{DockerConfigDir: configDir})).To(gomega.Succeed())
			_, err = NewContextWithStreams(context.Background(), io.Discard, io.Discard)
		}
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring(`context "missing"`))
	},
		ginkgo.Entry("Init", true),
		ginkgo.Entry("NewContextWithStreams", false),
	)

	ginkgo.DescribeTable("uses the CLI endpoint for image operations", func(selection string, useTLS, global bool) {
		server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer ginkgo.GinkgoRecover()
			if r.URL.Path == "/_ping" {
				w.Header().Set("API-Version", "1.47")
				w.Header().Set("OSType", "linux")
				return
			}
			if r.URL.Path != "/v1.47/images/json" {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			gomega.Expect(json.NewEncoder(w).Encode([]dockerimage.Summary{{ID: "remote-image"}})).To(gomega.Succeed())
		}))
		if useTLS {
			server.StartTLS()
		} else {
			server.Start()
		}
		ginkgo.DeferCleanup(server.Close)

		configDir := ginkgo.GinkgoT().TempDir()
		host := "tcp://" + server.Listener.Addr().String()
		contextStore := store.New(filepath.Join(configDir, "contexts"), command.DefaultContextStoreConfig())
		gomega.Expect(contextStore.CreateOrUpdate(store.Metadata{
			Name: "remote",
			Endpoints: map[string]any{
				"docker": contextdocker.EndpointMeta{Host: host},
			},
		})).To(gomega.Succeed())
		if useTLS {
			gomega.Expect(contextStore.ResetTLSMaterial("remote", &store.ContextTLSData{
				Endpoints: map[string]store.EndpointTLSData{
					"docker": {Files: map[string][]byte{
						"ca.pem": pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw}),
					}},
				},
			})).To(gomega.Succeed())
		}

		switch selection {
		case "stored":
			gomega.Expect(os.WriteFile(filepath.Join(configDir, "config.json"), []byte(`{"currentContext":"remote"}`), 0o600)).To(gomega.Succeed())
		case "context":
			ginkgo.GinkgoT().Setenv("DOCKER_CONTEXT", "remote")
			gomega.Expect(os.WriteFile(filepath.Join(configDir, "config.json"), []byte(`{"currentContext":"missing"}`), 0o600)).To(gomega.Succeed())
		case "host":
			gomega.Expect(os.WriteFile(filepath.Join(configDir, "config.json"), []byte(`{"currentContext":"missing"}`), 0o600)).To(gomega.Succeed())
			ginkgo.GinkgoT().Setenv("DOCKER_HOST", host)
			if useTLS {
				key, err := x509.MarshalPKCS8PrivateKey(server.TLS.Certificates[0].PrivateKey)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(os.WriteFile(filepath.Join(configDir, "cert.pem"), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw}), 0o600)).To(gomega.Succeed())
				gomega.Expect(os.WriteFile(filepath.Join(configDir, "key.pem"), pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: key}), 0o600)).To(gomega.Succeed())
				ginkgo.GinkgoT().Setenv("DOCKER_CERT_PATH", configDir)
			}
		}

		ctx := context.Background()
		var api client.APIClient
		if global {
			gomega.Expect(Init(ctx, InitOptions{DockerConfigDir: configDir})).To(gomega.Succeed())
			api = defaultAPIClient
			ginkgo.DeferCleanup(defaultCLI.Client().Close)
		} else {
			gomega.Expect(InitDockerConfig(InitOptions{DockerConfigDir: configDir})).To(gomega.Succeed())
			var err error
			ctx, err = NewContextWithStreams(ctx, io.Discard, io.Discard)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			api = apiCli(ctx)
			ginkgo.DeferCleanup(cli(ctx).Client().Close)
		}
		ginkgo.DeferCleanup(api.Close)
		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		images, err := Images(ctx, dockerimage.ListOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(images).To(gomega.HaveLen(1))
		gomega.Expect(images[0].ID).To(gomega.Equal("remote-image"))
		gomega.Expect(strings.TrimPrefix(api.DaemonHost(), "tcp://")).To(gomega.Equal(server.Listener.Addr().String()))
	},
		ginkgo.Entry("stored context in Init", "stored", false, true),
		ginkgo.Entry("stored context in NewContextWithStreams", "stored", false, false),
		ginkgo.Entry("DOCKER_CONTEXT overrides stored context in Init", "context", false, true),
		ginkgo.Entry("DOCKER_CONTEXT overrides stored context in NewContextWithStreams", "context", false, false),
		ginkgo.Entry("DOCKER_HOST overrides stored context in Init", "host", false, true),
		ginkgo.Entry("DOCKER_HOST overrides stored context in NewContextWithStreams", "host", false, false),
		ginkgo.Entry("legacy DOCKER_CERT_PATH without TLS_VERIFY in Init", "host", true, true),
		ginkgo.Entry("legacy DOCKER_CERT_PATH without TLS_VERIFY in NewContextWithStreams", "host", true, false),
		ginkgo.Entry("verified TLS context in Init", "stored", true, true),
		ginkgo.Entry("verified TLS context in NewContextWithStreams", "stored", true, false),
	)
})
