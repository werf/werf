package docker

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/docker/cli/cli/connhelper/commandconn"
	"github.com/moby/moby/client"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/opstats"
)

type failingTransport struct {
	err error
}

var _ http.RoundTripper = (*failingTransport)(nil)

func (t *failingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, t.err
}

func daemonTransportContext(transportErr error) context.Context {
	api, err := client.New(client.WithHost("http://127.0.0.1:1"), client.WithHTTPClient(&http.Client{Transport: &failingTransport{err: transportErr}}))
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	ginkgo.DeferCleanup(api.Close)
	ctx := context.WithValue(context.Background(), ctxDockerCliKey, true)
	return context.WithValue(ctx, ctxAPIClientKey, api)
}

func sshDaemonContext(diagnostic string) context.Context {
	if runtime.GOOS == "windows" {
		ginkgo.Skip("SSH connection-helper fixture requires a POSIX shell")
	}
	diagnosticPath := filepath.Join(ginkgo.GinkgoT().TempDir(), "stderr")
	gomega.Expect(os.WriteFile(diagnosticPath, []byte(diagnostic+"\n"), 0o600)).To(gomega.Succeed())
	transport := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return commandconn.New(ctx, "/bin/sh", "-c", `read request; cat "$0" >&2; exit 255`, diagnosticPath)
	}}
	api, err := client.New(client.WithHost("http://docker.example"), client.WithHTTPClient(&http.Client{Transport: transport}))
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	ginkgo.DeferCleanup(api.Close)
	ctx := context.WithValue(context.Background(), ctxDockerCliKey, true)
	return context.WithValue(ctx, ctxAPIClientKey, api)
}

func daemonSettingsContext(handler http.Handler) context.Context {
	server := httptest.NewServer(handler)
	ginkgo.DeferCleanup(server.Close)
	api, err := client.New(client.WithHost(server.URL))
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	ginkgo.DeferCleanup(api.Close)
	ctx := context.WithValue(context.Background(), ctxDockerCliKey, true)
	return context.WithValue(ctx, ctxAPIClientKey, api)
}

// newDelayedCreateServer returns the address of a daemon which stalls the container create
// request and then rejects it, so that a run which never reaches container start still takes
// at least delay.
func newDelayedCreateServer(delay time.Duration) string {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/_ping"):
			w.Header().Set("API-Version", "1.47")
			w.Header().Set("OSType", "linux")
			return
		case strings.HasSuffix(r.URL.Path, "/containers/create"):
			select {
			case <-time.After(delay):
			case <-r.Context().Done():
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		if _, err := io.WriteString(w, `{"message":"No such image: img"}`); err != nil {
			panic(err)
		}
	}))
	ginkgo.DeferCleanup(server.Close)
	return server.Listener.Addr().String()
}

// newMissingImageDaemonServer returns the address of a daemon rejecting every request as
// not found, an error the cli retry helper treats as permanent, so a command reaching it
// fails after a single attempt.
func newMissingImageDaemonServer() string {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/_ping") {
			w.Header().Set("API-Version", "1.47")
			w.Header().Set("OSType", "linux")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		if _, err := io.WriteString(w, `{"message":"No such image: img"}`); err != nil {
			panic(err)
		}
	}))
	ginkgo.DeferCleanup(server.Close)
	return server.Listener.Addr().String()
}

// cliDaemonContext points the docker CLI at addr and returns a context carrying a CLI bound
// to it, so that commands executed through the CLI (pull, push, run) reach the fake daemon
// instead of a real one.
func cliDaemonContext(addr string) context.Context {
	for _, key := range []string{"DOCKER_CONTEXT", "DOCKER_CONFIG", "DOCKER_CERT_PATH", "DOCKER_TLS", "DOCKER_TLS_VERIFY", "DOCKER_API_VERSION", "DOCKER_CONTENT_TRUST", "WERF_DEBUG_DOCKER"} {
		ginkgo.GinkgoT().Setenv(key, "")
	}
	ginkgo.GinkgoT().Setenv("DOCKER_HOST", "tcp://"+addr)
	gomega.Expect(InitDockerConfig(InitOptions{DockerConfigDir: ginkgo.GinkgoT().TempDir()})).To(gomega.Succeed())

	ctx, err := NewContextWithStreams(context.Background(), io.Discard, io.Discard)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	ginkgo.DeferCleanup(apiCli(ctx).Close)
	return ctx
}

func observedDaemonContext(handler http.HandlerFunc) (context.Context, *opstats.Collector) {
	collector := opstats.NewCollector()
	return opstats.NewContext(daemonSettingsContext(handler), collector), collector
}

func respondJSON(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/_ping") {
			w.Header().Set("API-Version", "1.47")
			w.Header().Set("OSType", "linux")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(w, body); err != nil {
			panic(err)
		}
	}
}

func respondError(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/_ping") {
		w.Header().Set("API-Version", "1.47")
		w.Header().Set("OSType", "linux")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	if _, err := io.WriteString(w, `{"message":"daemon failure"}`); err != nil {
		panic(err)
	}
}

func operationCount(collector *opstats.Collector, op opstats.Operation) int {
	for _, summary := range collector.Summary() {
		if summary.Operation == op {
			return summary.Count
		}
	}
	return 0
}
