package container_backend

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/docker"
	"github.com/werf/werf/v3/pkg/opstats"
)

var _ BuildContextArchiver = (*stubBuildContextArchive)(nil)

type stubBuildContextArchive struct {
	BuildContextArchiver
	dir string
	err error
}

func (a *stubBuildContextArchive) ExtractOrGetExtractedDir(_ context.Context) (string, error) {
	if a.err != nil {
		return "", a.err
	}
	return a.dir, nil
}

// dockerDaemonContext points the docker client at a fake daemon serving handler and returns
// a context carrying both that client and a fresh operation collector.
func dockerDaemonContext(handler http.Handler) (context.Context, *opstats.Collector) {
	server := httptest.NewServer(handler)
	ginkgo.DeferCleanup(server.Close)
	for _, key := range []string{"DOCKER_CONTEXT", "DOCKER_CONFIG", "DOCKER_TLS", "DOCKER_TLS_VERIFY", "DOCKER_CERT_PATH", "DOCKER_API_VERSION"} {
		ginkgo.GinkgoT().Setenv(key, "")
	}
	ginkgo.GinkgoT().Setenv("DOCKER_HOST", "tcp://"+server.Listener.Addr().String())

	ctx, err := docker.NewContextWithStreams(context.Background(), io.Discard, io.Discard)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	collector := opstats.NewCollector()
	return opstats.NewContext(ctx, collector), collector
}

func daemonHandler(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/_ping") {
			w.Header().Set("API-Version", "1.47")
			w.Header().Set("OSType", "linux")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, err := io.WriteString(w, body)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
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
