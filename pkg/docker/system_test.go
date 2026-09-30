package docker

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"time"

	"github.com/docker/cli/cli/command"
	cliconfig "github.com/docker/cli/cli/config"
	"github.com/moby/moby/client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CheckConnection", func() {
	var configDir string

	BeforeEach(func() {
		oldConfigDir, oldCliConfigDir := DockerConfigDir, cliconfig.Dir()
		oldTimeout := daemonPingTimeout
		DeferCleanup(func() {
			DockerConfigDir = oldConfigDir
			cliconfig.SetDir(oldCliConfigDir)
			daemonPingTimeout = oldTimeout
		})
		for _, key := range []string{"DOCKER_HOST", "DOCKER_CONTEXT", "DOCKER_CONFIG", "DOCKER_CERT_PATH", "DOCKER_TLS", "DOCKER_TLS_VERIFY", "DOCKER_API_VERSION", "WERF_DEBUG_DOCKER"} {
			GinkgoT().Setenv(key, "")
		}
		configDir = GinkgoT().TempDir()
	})

	// muteDaemonContext binds a context to a daemon which accepts connections and never answers.
	muteDaemonContext := func(configDir string) context.Context {
		daemonPingTimeout = 300 * time.Millisecond
		blocked := make(chan struct{})
		DeferCleanup(func() { close(blocked) })
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-blocked:
			case <-r.Context().Done():
			}
		}))
		DeferCleanup(server.Close)
		GinkgoT().Setenv("DOCKER_HOST", "tcp://"+server.Listener.Addr().String())
		Expect(InitDockerConfig(InitOptions{DockerConfigDir: configDir})).To(Succeed())
		ctx, err := NewContextWithStreams(context.Background(), io.Discard, io.Discard)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(apiCli(ctx).Close)
		return ctx
	}

	DescribeTable("does not wait forever for a daemon which never answers", func(allowDaemonUnavailable bool) {
		ctx := muteDaemonContext(configDir)

		started := time.Now()
		err := CheckConnection(ctx, CheckConnectionOptions{AllowDaemonUnavailable: allowDaemonUnavailable})
		Expect(time.Since(started)).To(BeNumerically("<", 5*time.Second))
		Expect(err).To(MatchError(ContainSubstring("daemon did not answer within 300ms")))
	},
		Entry("daemon-required command fails", false),
		Entry("an unresponsive daemon is not an unavailable one", true),
	)

	It("gives up on daemon info when the daemon never answers", func() {
		ctx := muteDaemonContext(configDir)

		started := time.Now()
		mirrors, err := GetRegistryMirrors(ctx)
		Expect(time.Since(started)).To(BeNumerically("<", 5*time.Second))
		Expect(err).NotTo(HaveOccurred())
		Expect(mirrors).To(BeNil())
	})

	It("keeps a daemon which answers the ping but not the info request usable", func() {
		daemonPingTimeout = 300 * time.Millisecond
		blocked := make(chan struct{})
		DeferCleanup(func() { close(blocked) })
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasSuffix(r.URL.Path, "/info") {
				w.Header().Set("API-Version", "1.47")
				w.Header().Set("OSType", "linux")
				return
			}
			select {
			case <-blocked:
			case <-r.Context().Done():
			}
		}))
		DeferCleanup(server.Close)
		GinkgoT().Setenv("DOCKER_HOST", "tcp://"+server.Listener.Addr().String())
		Expect(InitDockerConfig(InitOptions{DockerConfigDir: configDir})).To(Succeed())
		ctx, err := NewContextWithStreams(context.Background(), io.Discard, io.Discard)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(apiCli(ctx).Close)

		mirrors, err := GetRegistryMirrors(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(mirrors).To(BeNil())

		Expect(CheckConnection(ctx, CheckConnectionOptions{})).To(Succeed())
	})

	DescribeTable("re-pings a daemon which was not running yet at init", func(settings bool) {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		Expect(err).NotTo(HaveOccurred())
		addr := listener.Addr().String()
		Expect(listener.Close()).To(Succeed())

		GinkgoT().Setenv("DOCKER_HOST", "tcp://"+addr)
		Expect(InitDockerConfig(InitOptions{DockerConfigDir: configDir})).To(Succeed())
		ctx, err := NewContextWithStreams(context.Background(), io.Discard, io.Discard)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(apiCli(ctx).Close)

		if settings {
			mirrors, err := GetRegistryMirrors(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(mirrors).To(BeNil())
		} else {
			Expect(CheckConnection(ctx, CheckConnectionOptions{AllowDaemonUnavailable: true})).To(Succeed())
		}

		started, err := net.Listen("tcp", addr)
		Expect(err).NotTo(HaveOccurred())
		server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_, err := w.Write([]byte(`{"message":"daemon now requires authorization"}`))
			if err != nil {
				panic(err)
			}
		}))
		Expect(server.Listener.Close()).To(Succeed())
		server.Listener = started
		server.Start()
		DeferCleanup(server.Close)

		if settings {
			_, err := GetRegistryMirrors(ctx)
			Expect(err).To(MatchError(ContainSubstring("daemon now requires authorization")))
		} else {
			Expect(CheckConnection(ctx, CheckConnectionOptions{})).To(MatchError(ContainSubstring("daemon now requires authorization")))
		}
	}, Entry("connection check", false), Entry("optional settings", true))

	It("fails the daemon info request when the daemon never answers", func() {
		ctx := muteDaemonContext(configDir)

		started := time.Now()
		_, err := Info(ctx)
		Expect(time.Since(started)).To(BeNumerically("<", 5*time.Second))
		Expect(err).To(MatchError(ContainSubstring("Docker daemon did not answer within 300ms")))
	})

	It("pings the daemon once per client", func() {
		var pings atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pings.Add(1)
			w.Header().Set("API-Version", "1.47")
			w.Header().Set("OSType", "linux")
		}))
		DeferCleanup(server.Close)
		GinkgoT().Setenv("DOCKER_HOST", "tcp://"+server.Listener.Addr().String())
		Expect(InitDockerConfig(InitOptions{DockerConfigDir: configDir})).To(Succeed())
		ctx, err := NewContextWithStreams(context.Background(), io.Discard, io.Discard)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(apiCli(ctx).Close)

		Expect(CheckConnection(ctx, CheckConnectionOptions{})).To(Succeed())
		afterFirstCheck := pings.Load()
		Expect(afterFirstCheck).NotTo(BeZero())
		Expect(CheckConnection(ctx, CheckConnectionOptions{})).To(Succeed())
		Expect(pings.Load()).To(Equal(afterFirstCheck))
	})
})

var _ = Describe("docker system", func() {
	Describe("cliWithCustomOptions", func() {
		It("uses a separate Docker CLI", func() {
			ctx, err := NewContext(context.Background())
			Expect(err).NotTo(HaveOccurred())

			originalCli := cli(ctx)
			var customCli command.Cli
			err = cliWithCustomOptions(ctx, nil, func(c command.Cli) error {
				customCli = c
				return nil
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(customCli).NotTo(Equal(originalCli))
			Expect(cli(ctx)).To(Equal(originalCli))
		})
	})

	Describe("isDaemonUnavailableErr", func() {
		It("should return false for nil error", func() {
			Expect(isDaemonUnavailableErr(nil)).To(BeFalse())
		})

		It("does not treat untyped connection refused text as daemon unavailability", func() {
			err := errors.New("connect: connection refused")
			Expect(isDaemonUnavailableErr(err)).To(BeFalse())
		})

		It("does not treat untyped no such file text as daemon unavailability", func() {
			err := errors.New("connect: no such file or directory")
			Expect(isDaemonUnavailableErr(err)).To(BeFalse())
		})

		It("does not treat untyped cannot connect text as daemon unavailability", func() {
			err := errors.New("Cannot connect to the Docker daemon")
			Expect(isDaemonUnavailableErr(err)).To(BeFalse())
		})

		It("should return false for other errors", func() {
			err := errors.New("some other error")
			Expect(isDaemonUnavailableErr(err)).To(BeFalse())
		})
	})
})

var _ = Describe("optional daemon settings", func() {
	DescribeTable("propagates daemon errors", func(path string, status int, body string) {
		ctx := daemonSettingsContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			w.Header().Set("API-Version", "1.40")
			w.Header().Set("Content-Type", "application/json")
			if strings.HasSuffix(r.URL.Path, path) {
				w.WriteHeader(status)
				_, err := w.Write([]byte(body))
				Expect(err).NotTo(HaveOccurred())
				return
			}
			_, err := w.Write([]byte(`{"RegistryConfig":{"Mirrors":["https://mirror.example"]}}`))
			Expect(err).NotTo(HaveOccurred())
		}))
		_, err := GetRegistryMirrors(ctx)
		Expect(err).To(HaveOccurred())
	},
		Entry("info authentication", "/info", http.StatusUnauthorized, `{"message":"authentication required"}`),
		Entry("info forbidden", "/info", http.StatusForbidden, `{"message":"access denied"}`),
		Entry("info unexpected server failure", "/info", http.StatusInternalServerError, `{"message":"Cannot connect to the Docker daemon"}`),
		Entry("info malformed response", "/info", http.StatusOK, `{"RegistryConfig":`),
		Entry("ping authentication", "/_ping", http.StatusUnauthorized, `{"message":"authentication required"}`),
		Entry("ping unexpected server failure", "/_ping", http.StatusInternalServerError, `{"message":"daemon broken"}`),
	)

	It("does not swallow caller cancellation", func() {
		ctx := daemonSettingsContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("API-Version", "1.40")
		}))
		ctx, cancel := context.WithCancel(ctx)
		cancel()
		_, err := GetRegistryMirrors(ctx)
		Expect(errors.Is(err, context.Canceled)).To(BeTrue(), "error: %v", err)
	})

	DescribeTable("preserves caller cancellation during a request", func(path string) {
		var cancel context.CancelFunc
		ctx := daemonSettingsContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("API-Version", "1.40")
			if strings.HasSuffix(r.URL.Path, path) {
				cancel()
				<-r.Context().Done()
			}
		}))
		ctx, cancel = context.WithCancel(ctx)
		defer cancel()
		_, err := GetRegistryMirrors(ctx)
		Expect(errors.Is(err, context.Canceled)).To(BeTrue(), "error: %v", err)
	}, Entry("ping", "/_ping"), Entry("info", "/info"))

	It("preserves an expired caller deadline", func() {
		ctx := daemonSettingsContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		ctx, cancel := context.WithDeadline(ctx, time.Now().Add(-time.Second))
		defer cancel()
		_, err := GetRegistryMirrors(ctx)
		Expect(errors.Is(err, context.DeadlineExceeded)).To(BeTrue(), "error: %v", err)
	})

	It("rejects malformed advertised versions", func() {
		ctx := daemonSettingsContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("API-Version", "1.invalid")
		}))
		_, err := GetRegistryMirrors(ctx)
		Expect(err).To(MatchError(ContainSubstring("invalid Docker daemon API version")))
	})

	It("propagates TLS failures", func() {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		DeferCleanup(server.Close)
		api, err := client.New(client.WithHost(server.URL), client.WithScheme("https"))
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(api.Close)
		ctx := context.WithValue(context.Background(), ctxDockerCliKey, true)
		ctx = context.WithValue(ctx, ctxAPIClientKey, api)
		_, err = GetRegistryMirrors(ctx)
		Expect(err).To(MatchError(ContainSubstring("certificate")))
	})
})
