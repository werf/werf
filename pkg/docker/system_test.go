package docker

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"syscall"
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

	DescribeTable("rejects malformed advertised versions", func(version string) {
		ctx := daemonSettingsContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("API-Version", version)
		}))
		_, err := GetRegistryMirrors(ctx)
		Expect(err).To(MatchError(ContainSubstring("invalid Docker daemon API version")))
	},
		Entry("non-numeric minor", "1.invalid"),
		Entry("signed minor", "1.+39"),
		Entry("signed major", "+1.39"),
		Entry("negative major", "-0.39"),
		Entry("negative minor", "1.-39"),
		Entry("hexadecimal major", "0x1.39"),
		Entry("underscored minor", "1.3_9"),
		Entry("no minor", "1"),
		Entry("empty minor", "1."),
	)

	DescribeTable("uses well-formed advertised versions", func(version string, expectedMirrors []string) {
		ctx := daemonSettingsContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			w.Header().Set("API-Version", version)
			if r.Method == http.MethodHead {
				return
			}
			_, err := w.Write([]byte(`{"RegistryConfig":{"Mirrors":["https://mirror.example"]}}`))
			Expect(err).NotTo(HaveOccurred())
		}))
		mirrors, err := GetRegistryMirrors(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(mirrors).To(Equal(expectedMirrors))
	},
		Entry("an old API is skipped", "1.39", []string(nil)),
		Entry("a supported API is used", "1.40", []string{"https://mirror.example"}),
	)

	DescribeTable("does not swallow caller cancellation once the daemon API version is cached", func(version string) {
		ctx := daemonSettingsContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			w.Header().Set("API-Version", version)
			if r.Method == http.MethodHead {
				return
			}
			_, err := w.Write([]byte(`{"RegistryConfig":{"Mirrors":["https://mirror.example"]}}`))
			Expect(err).NotTo(HaveOccurred())
		}))
		_, err := GetRegistryMirrors(ctx)
		Expect(err).NotTo(HaveOccurred())

		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()
		_, err = GetRegistryMirrors(canceledCtx)
		Expect(errors.Is(err, context.Canceled)).To(BeTrue(), "error: %v", err)
	},
		Entry("an old API", "1.39"),
		Entry("a supported API", "1.40"),
	)

	// wsaeConnRefused is Winsock's WSAECONNREFUSED, which a Windows client gets from a Docker
	// TCP endpoint whose port refuses connections:
	// https://learn.microsoft.com/en-us/windows/win32/winsock/windows-sockets-error-codes-2
	const wsaeConnRefusedFixture = syscall.Errno(10061)

	DescribeTable("classifies transport failures", func(transportErr error, daemonUnavailable bool) {
		ctx := daemonTransportContext(transportErr)
		mirrors, err := GetRegistryMirrors(ctx)
		if !daemonUnavailable {
			Expect(err).To(HaveOccurred())
			return
		}
		Expect(err).NotTo(HaveOccurred())
		Expect(mirrors).To(BeNil())
	},
		Entry("a refused TCP connection", &net.OpError{Op: "dial", Net: "tcp", Err: os.NewSyscallError("connect", syscall.ECONNREFUSED)}, true),
		Entry("a refused TCP connection on Windows", &net.OpError{Op: "dial", Net: "tcp", Err: os.NewSyscallError("connectex", wsaeConnRefusedFixture)}, true),
		Entry("an unreachable TCP network", &net.OpError{Op: "dial", Net: "tcp", Err: os.NewSyscallError("connect", syscall.ENETUNREACH)}, true),
		Entry("an unroutable TCP host", &net.OpError{Op: "dial", Net: "tcp", Err: os.NewSyscallError("connect", syscall.EHOSTUNREACH)}, true),
		Entry("a down TCP network", &net.OpError{Op: "dial", Net: "tcp", Err: os.NewSyscallError("connect", syscall.ENETDOWN)}, true),
		Entry("a down TCP host", &net.OpError{Op: "dial", Net: "tcp", Err: os.NewSyscallError("connect", syscall.EHOSTDOWN)}, true),
		Entry("an unreachable TCP network on Windows", &net.OpError{Op: "dial", Net: "tcp", Err: os.NewSyscallError("connectex", syscall.Errno(10051))}, true),
		Entry("an unroutable TCP host on Windows", &net.OpError{Op: "dial", Net: "tcp", Err: os.NewSyscallError("connectex", syscall.Errno(10065))}, true),
		Entry("a down TCP network on Windows", &net.OpError{Op: "dial", Net: "tcp", Err: os.NewSyscallError("connectex", syscall.Errno(10050))}, true),
		Entry("a down TCP host on Windows", &net.OpError{Op: "dial", Net: "tcp", Err: os.NewSyscallError("connectex", syscall.Errno(10064))}, true),
		Entry("a missing Unix socket", &net.OpError{Op: "dial", Net: "unix", Err: os.NewSyscallError("connect", syscall.ENOENT)}, true),
		Entry("an unresolvable daemon host", &net.OpError{Op: "dial", Net: "tcp", Err: &net.DNSError{Err: "no such host", Name: "docker.example", IsNotFound: true}}, true),
		Entry("a Unix socket which denies access", &net.OpError{Op: "dial", Net: "unix", Err: os.NewSyscallError("connect", syscall.EACCES)}, false),
		Entry("a denied SSH authentication", errors.New("command [ssh docker system dial-stdio] has exited with exit status 255: stderr=Permission denied (publickey)."), false),
	)

	DescribeTable("classifies SSH connection helper failures", func(diagnostic string, daemonUnavailable bool) {
		mirrors, err := GetRegistryMirrors(sshDaemonContext(diagnostic))
		if !daemonUnavailable {
			Expect(err).To(HaveOccurred())
			return
		}
		Expect(err).NotTo(HaveOccurred())
		Expect(mirrors).To(BeNil())
	},
		Entry("a refused SSH connection", "ssh: connect to host docker.example port 22: Connection refused", true),
		Entry("an unreachable SSH network", "ssh: connect to host docker.example port 22: Network is unreachable", true),
		Entry("a down SSH network", "ssh: connect to host docker.example port 22: Network is down", true),
		Entry("a down SSH host", "ssh: connect to host docker.example port 22: Host is down", true),
		Entry("an unroutable SSH host", "ssh: connect to host docker.example port 22: No route to host", true),
		Entry("a timed out SSH connection", "ssh: connect to host docker.example port 22: Operation timed out", true),
		Entry("an unresolvable SSH host", "ssh: Could not resolve hostname docker.example: nodename nor servname provided, or not known", true),
		Entry("a denied SSH authentication", "docker.example: Permission denied (publickey).", false),
		Entry("a denied SSH socket connection", "ssh: connect to host docker.example port 22: Permission denied", false),
		Entry("a remote host without the docker CLI", "bash: line 1: docker: command not found", false),
	)

	It("propagates a plaintext request to a TLS daemon", func() {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		DeferCleanup(server.Close)
		api, err := client.New(client.WithHost("tcp://" + server.Listener.Addr().String()))
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(api.Close)
		ctx := context.WithValue(context.Background(), ctxDockerCliKey, true)
		ctx = context.WithValue(ctx, ctxAPIClientKey, api)
		_, err = GetRegistryMirrors(ctx)
		Expect(err).To(MatchError(ContainSubstring("Client sent an HTTP request to an HTTPS server")))
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
