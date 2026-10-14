package import_server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/docker"
)

func importServerDockerContext(ctx context.Context, status int, body string) context.Context {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer ginkgo.GinkgoRecover()
		if r.URL.Path == "/_ping" {
			w.Header().Set("API-Version", "1.47")
			w.Header().Set("OSType", "linux")
			return
		}
		gomega.Expect(r.URL.Path).To(gomega.ContainSubstring("/containers/owned-import-server"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, err := io.WriteString(w, body)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}))
	ginkgo.DeferCleanup(server.Close)
	for _, key := range []string{"DOCKER_CONTEXT", "DOCKER_TLS", "DOCKER_TLS_VERIFY", "DOCKER_CERT_PATH", "DOCKER_API_v2"} {
		ginkgo.GinkgoT().Setenv(key, "")
	}
	ginkgo.GinkgoT().Setenv("DOCKER_CONFIG", ginkgo.GinkgoT().TempDir())
	ginkgo.GinkgoT().Setenv("DOCKER_HOST", "tcp://"+server.Listener.Addr().String())
	result, err := docker.NewContext(ctx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return result
}
