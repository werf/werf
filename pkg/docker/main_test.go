package docker

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/docker/cli/cli/command"
	cliconfig "github.com/docker/cli/cli/config"
	contextdocker "github.com/docker/cli/cli/context/docker"
	"github.com/docker/cli/cli/context/store"
	dockercontainer "github.com/moby/moby/api/types/container"
	dockerimage "github.com/moby/moby/api/types/image"
	"github.com/moby/moby/client"
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
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

	ginkgo.It("does not switch the default socket to TCP when DOCKER_CERT_PATH is set", func() {
		configDir := ginkgo.GinkgoT().TempDir()
		ginkgo.GinkgoT().Setenv("DOCKER_CERT_PATH", configDir)
		gomega.Expect(InitDockerConfig(InitOptions{DockerConfigDir: configDir})).To(gomega.Succeed())
		c, err := newDockerCli(cliOptionsWithStreams(io.Discard, io.Discard))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		api := c.Client()
		ginkgo.DeferCleanup(api.Close)
		gomega.Expect(api.DaemonHost()).To(gomega.Equal(client.DefaultDockerHost))
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

	ginkgo.DescribeTable("preserves container creation API requirements", func(apiVersion string, startInterval time.Duration, expectedError string) {
		var creates atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer ginkgo.GinkgoRecover()
			if r.URL.Path == "/_ping" {
				w.Header().Set("API-Version", apiVersion)
				w.Header().Set("OSType", "linux")
				return
			}
			gomega.Expect(r.Method).To(gomega.Equal(http.MethodPost))
			gomega.Expect(r.URL.Path).To(gomega.Equal("/v" + apiVersion + "/containers/create"))
			gomega.Expect(r.URL.Query().Get("platform")).To(gomega.Equal("linux/amd64"))
			creates.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_, err := io.WriteString(w, `{"Id":"created"}`)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}))
		ginkgo.DeferCleanup(server.Close)
		ginkgo.GinkgoT().Setenv("DOCKER_HOST", "tcp://"+server.Listener.Addr().String())
		gomega.Expect(InitDockerConfig(InitOptions{DockerConfigDir: ginkgo.GinkgoT().TempDir()})).To(gomega.Succeed())
		ctx, err := NewContextWithStreams(context.Background(), io.Discard, io.Discard)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		ginkgo.DeferCleanup(apiCli(ctx).Close)
		config := &dockercontainer.Config{Image: "source"}
		if startInterval != 0 {
			config.Healthcheck = &dockercontainer.HealthConfig{StartInterval: startInterval}
		}
		id, err := ContainerCreate(ctx, config, &ocispec.Platform{OS: "linux", Architecture: "amd64"}, "target")
		if expectedError != "" {
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring(expectedError))
			gomega.Expect(creates.Load()).To(gomega.BeZero())
			return
		}
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(id).To(gomega.Equal("created"))
		gomega.Expect(creates.Load()).To(gomega.Equal(int32(1)))
	},
		ginkgo.Entry("rejects platform on API 1.40", "1.40", time.Duration(0), `"specify container image platform" requires API version 1.41`),
		ginkgo.Entry("allows platform on API 1.41", "1.41", time.Duration(0), ""),
		ginkgo.Entry("rejects start interval on API 1.43", "1.43", time.Second, `"specify health-check start interval" requires API version 1.44`),
		ginkgo.Entry("allows start interval on API 1.44", "1.44", time.Second, ""),
	)

	ginkgo.DescribeTable("preserves image load stream responses", func(response, expectedID, expectedError string) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer ginkgo.GinkgoRecover()
			switch r.URL.Path {
			case "/_ping":
				w.Header().Set("API-Version", "1.47")
				w.Header().Set("OSType", "linux")
			case "/v1.47/images/load":
				w.Header().Set("Content-Type", "application/json")
				_, err := io.WriteString(w, response)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			case "/v1.47/images/remote:latest/json":
				w.Header().Set("Content-Type", "application/json")
				_, err := io.WriteString(w, `{"Id":"sha256:from-inspect"}`)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			default:
				http.NotFound(w, r)
			}
		}))
		ginkgo.DeferCleanup(server.Close)
		ginkgo.GinkgoT().Setenv("DOCKER_HOST", "tcp://"+server.Listener.Addr().String())
		gomega.Expect(InitDockerConfig(InitOptions{DockerConfigDir: ginkgo.GinkgoT().TempDir()})).To(gomega.Succeed())
		ctx, err := NewContextWithStreams(context.Background(), io.Discard, io.Discard)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		ginkgo.DeferCleanup(apiCli(ctx).Close)

		id, err := CliLoadFromStream(ctx, strings.NewReader("archive"))
		if expectedError != "" {
			gomega.Expect(err).To(gomega.MatchError(expectedError))
			return
		}
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(id).To(gomega.Equal(expectedID))
	},
		ginkgo.Entry("image ID", `{"stream":"Loaded image ID: sha256:loaded\n"}`, "loaded", ""),
		ginkgo.Entry("image reference", `{"stream":"Loaded image: remote:latest\n"}`, "from-inspect", ""),
		ginkgo.Entry("structured error", `{"errorDetail":{"message":"load denied"}}`, "", "load failed: load denied"),
		ginkgo.Entry("legacy error", `{"error":"load denied"}`, "", "load failed: load denied"),
	)

	ginkgo.DescribeTable("uses the CLI endpoint for image operations", func(selection string, useTLS, global bool, apiVersion string) {
		server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer ginkgo.GinkgoRecover()
			if r.URL.Path == "/_ping" {
				w.Header().Set("API-Version", apiVersion)
				w.Header().Set("OSType", "linux")
				return
			}
			if r.URL.Path != "/v"+apiVersion+"/images/json" {
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
		case "host", "verified-host", "wrong-host":
			gomega.Expect(os.WriteFile(filepath.Join(configDir, "config.json"), []byte(`{"currentContext":"missing"}`), 0o600)).To(gomega.Succeed())
			ginkgo.GinkgoT().Setenv("DOCKER_HOST", host)
			if useTLS {
				key, err := x509.MarshalPKCS8PrivateKey(server.TLS.Certificates[0].PrivateKey)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(os.WriteFile(filepath.Join(configDir, "cert.pem"), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw}), 0o600)).To(gomega.Succeed())
				gomega.Expect(os.WriteFile(filepath.Join(configDir, "key.pem"), pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: key}), 0o600)).To(gomega.Succeed())
				ginkgo.GinkgoT().Setenv("DOCKER_CERT_PATH", configDir)
				if selection != "host" {
					ginkgo.GinkgoT().Setenv("DOCKER_TLS_VERIFY", "1")
					gomega.Expect(os.WriteFile(filepath.Join(configDir, "ca.pem"), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw}), 0o600)).To(gomega.Succeed())
				}
				if selection == "wrong-host" {
					gomega.Expect(server.Certificate().VerifyHostname("localhost")).To(gomega.HaveOccurred())
					_, port, err := net.SplitHostPort(server.Listener.Addr().String())
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
					ginkgo.GinkgoT().Setenv("DOCKER_HOST", "tcp://"+net.JoinHostPort("localhost", port))
				}
			}
		}

		ctx := context.Background()
		var api client.APIClient
		if global {
			gomega.Expect(Init(ctx, InitOptions{DockerConfigDir: configDir})).To(gomega.Succeed())
			api = defaultAPIClient
			gomega.Expect(api).To(gomega.BeIdenticalTo(defaultCLI.Client()))
		} else {
			gomega.Expect(InitDockerConfig(InitOptions{DockerConfigDir: configDir})).To(gomega.Succeed())
			var err error
			ctx, err = NewContextWithStreams(ctx, io.Discard, io.Discard)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			api = apiCli(ctx)
			gomega.Expect(api).To(gomega.BeIdenticalTo(cli(ctx).Client()))
		}
		ginkgo.DeferCleanup(api.Close)
		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		if apiVersion == "1.39" {
			_, err := api.Ping(ctx, client.PingOptions{NegotiateAPIVersion: true, ForceNegotiate: true})
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("minimum supported API version is 1.40"))
			return
		}
		images, err := Images(ctx, client.ImageListOptions{})
		if selection == "wrong-host" {
			gomega.Expect(err).To(gomega.HaveOccurred())
			var hostnameError x509.HostnameError
			gomega.Expect(errors.As(err, &hostnameError)).To(gomega.BeTrue())
			return
		}
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(images).To(gomega.HaveLen(1))
		gomega.Expect(images[0].ID).To(gomega.Equal("remote-image"))
		gomega.Expect(strings.TrimPrefix(api.DaemonHost(), "tcp://")).To(gomega.Equal(server.Listener.Addr().String()))
	},
		ginkgo.Entry("stored context in Init", "stored", false, true, "1.47"),
		ginkgo.Entry("stored context in NewContextWithStreams", "stored", false, false, "1.47"),
		ginkgo.Entry("DOCKER_CONTEXT overrides stored context in Init", "context", false, true, "1.47"),
		ginkgo.Entry("DOCKER_CONTEXT overrides stored context in NewContextWithStreams", "context", false, false, "1.47"),
		ginkgo.Entry("DOCKER_HOST overrides stored context in Init", "host", false, true, "1.47"),
		ginkgo.Entry("DOCKER_HOST overrides stored context in NewContextWithStreams", "host", false, false, "1.47"),
		ginkgo.Entry("legacy DOCKER_CERT_PATH without TLS_VERIFY in Init", "host", true, true, "1.47"),
		ginkgo.Entry("legacy DOCKER_CERT_PATH without TLS_VERIFY in NewContextWithStreams", "host", true, false, "1.47"),
		ginkgo.Entry("verified TLS context in Init", "stored", true, true, "1.47"),
		ginkgo.Entry("verified TLS context in NewContextWithStreams", "stored", true, false, "1.47"),
		ginkgo.Entry("DOCKER_TLS_VERIFY in Init", "verified-host", true, true, "1.47"),
		ginkgo.Entry("DOCKER_TLS_VERIFY in NewContextWithStreams", "verified-host", true, false, "1.47"),
		ginkgo.Entry("rejects TLS hostname mismatch in Init", "wrong-host", true, true, "1.47"),
		ginkgo.Entry("rejects TLS hostname mismatch in NewContextWithStreams", "wrong-host", true, false, "1.47"),
		ginkgo.Entry("minimum Docker API in Init", "stored", false, true, "1.40"),
		ginkgo.Entry("minimum Docker API in NewContextWithStreams", "stored", false, false, "1.40"),
		ginkgo.Entry("unsupported Docker API in Init", "stored", false, true, "1.39"),
		ginkgo.Entry("unsupported Docker API in NewContextWithStreams", "stored", false, false, "1.39"),
	)
})
