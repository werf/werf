package container_backend

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"github.com/samber/lo"

	"github.com/werf/werf/v3/pkg/docker"
	"github.com/werf/werf/v3/pkg/image"
)

var _ = ginkgo.DescribeTable("Docker layer compression exporter attributes", func(format string, level *int, expected map[string]string) {
	gomega.Expect(compressionExportAttrs(image.LayerCompression{Format: format, Level: level})).To(gomega.Equal(expected))
},
	ginkgo.Entry("no options keep the daemon default", "", nil, nil),
	ginkgo.Entry("explicit gzip is forced", "gzip", nil, map[string]string{"compression": "gzip", "force-compression": "true"}),
	ginkgo.Entry("gzip level", "gzip", lo.ToPtr(1), map[string]string{"compression": "gzip", "compression-level": "1", "force-compression": "true"}),
	ginkgo.Entry("level without format is gzip", "", lo.ToPtr(0), map[string]string{"compression": "gzip", "compression-level": "0", "force-compression": "true"}),
	ginkgo.Entry("zstd with level", "zstd", lo.ToPtr(19), map[string]string{"compression": "zstd", "compression-level": "19", "force-compression": "true", "oci-mediatypes": "true"}),
	ginkgo.Entry("none", "none", nil, map[string]string{"compression": "uncompressed", "force-compression": "true", "oci-mediatypes": "true"}),
)

var _ = ginkgo.DescribeTable("Docker layer compression requires the containerd image store", func(driver string, driverStatus [][2]string, expectedError string) {
	var infoRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer ginkgo.GinkgoRecover()
		if r.URL.Path == "/_ping" {
			w.Header().Set("API-Version", "1.47")
			w.Header().Set("OSType", "linux")
			return
		}
		gomega.Expect(r.URL.Path).To(gomega.Equal("/v1.47/info"))
		infoRequests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		gomega.Expect(json.NewEncoder(w).Encode(map[string]any{"Driver": driver, "DriverStatus": driverStatus})).To(gomega.Succeed())
	}))
	ginkgo.DeferCleanup(server.Close)
	for _, key := range []string{"DOCKER_CONTEXT", "DOCKER_TLS", "DOCKER_TLS_VERIFY", "DOCKER_CERT_PATH", "DOCKER_API_VERSION"} {
		ginkgo.GinkgoT().Setenv(key, "")
	}
	ginkgo.GinkgoT().Setenv("DOCKER_HOST", "tcp://"+server.Listener.Addr().String())
	ctx, err := docker.NewContextWithStreams(context.Background(), io.Discard, io.Discard)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	backend := NewDockerServerBackend(nil, DockerServerBackendOptions{CompressionFormat: "zstd"})
	for range 2 {
		err = backend.checkLayerCompressionSupported(ctx)
		if expectedError == "" {
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		} else {
			gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring(expectedError)))
		}
	}
	gomega.Expect(infoRequests.Load()).To(gomega.Equal(int32(1)))
},
	ginkgo.Entry("containerd image store", "overlayfs", [][2]string{{"driver-type", "io.containerd.snapshotter.v1"}}, ""),
	ginkgo.Entry("classic image store", "overlay2", [][2]string{{"Backing Filesystem", "extfs"}, {"Supports d_type", "true"}}, `require the containerd image store with the Docker container backend, but the Docker daemon uses the "overlay2" storage driver`),
)
