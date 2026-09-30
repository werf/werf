package container_backend

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/docker"
)

var _ = ginkgo.DescribeTable("Docker removal errors", func(message string, expected error) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer ginkgo.GinkgoRecover()
		if r.URL.Path == "/_ping" {
			w.Header().Set("API-Version", "1.47")
			w.Header().Set("OSType", "linux")
			return
		}
		gomega.Expect(r.Method).To(gomega.Equal(http.MethodDelete))
		gomega.Expect(r.URL.Path).To(gomega.Equal("/v1.47/containers/target"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		gomega.Expect(json.NewEncoder(w).Encode(map[string]string{"message": message})).To(gomega.Succeed())
	}))
	ginkgo.DeferCleanup(server.Close)
	for _, key := range []string{"DOCKER_CONTEXT", "DOCKER_TLS", "DOCKER_TLS_VERIFY", "DOCKER_CERT_PATH", "DOCKER_API_VERSION"} {
		ginkgo.GinkgoT().Setenv(key, "")
	}
	ginkgo.GinkgoT().Setenv("DOCKER_HOST", "tcp://"+server.Listener.Addr().String())
	ctx, err := docker.NewContextWithStreams(context.Background(), io.Discard, io.Discard)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	backend := &DockerServerBackend{}
	err = backend.Rm(ctx, "target", RmOpts{})
	gomega.Expect(errors.Is(err, expected)).To(gomega.BeTrue(), "%v", err)
},
	ginkgo.Entry("paused", `cannot remove container "target": container is paused and must be unpaused first`, ErrCannotRemovePausedContainer),
	ginkgo.Entry("running", `cannot remove container "target": container is running: stop the container before removing or force remove`, ErrCannotRemoveRunningContainer),
)

var _ = ginkgo.DescribeTable("Docker exposed ports", func(input map[string]struct{}, expectedJSON, expectedError string) {
	ports, err := toPortSet(input)
	if expectedError != "" {
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring(expectedError))
		return
	}
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	data, err := json.Marshal(ports)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(string(data)).To(gomega.MatchJSON(expectedJSON))
},
	ginkgo.Entry("empty", map[string]struct{}{}, `null`, ""),
	ginkgo.Entry("TCP and UDP", map[string]struct{}{"80/tcp": {}, "53/udp": {}}, `{"80/tcp":{},"53/udp":{}}`, ""),
	ginkgo.Entry("default protocol", map[string]struct{}{"443": {}}, `{"443/tcp":{}}`, ""),
	ginkgo.Entry("invalid port", map[string]struct{}{"invalid/tcp": {}}, "", `parse port "invalid/tcp"`),
)
