package docker

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/opstats"
)

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

	ctx, err := NewContext(context.Background())
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

func daemonSettingsContext(handler http.Handler) context.Context {
	server := httptest.NewServer(handler)
	ginkgo.DeferCleanup(server.Close)
	return cliDaemonContext(server.Listener.Addr().String())
}
