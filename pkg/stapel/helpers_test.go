package stapel

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"

	dockercontainer "github.com/moby/moby/api/types/container"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/docker"
)

func stapelVolumeMount(volumeName string) dockercontainer.MountPoint {
	return dockercontainer.MountPoint{
		Type:        "volume",
		Name:        volumeName,
		Destination: containerVolumeDestination,
	}
}

// fakeContainer is an inspect response as the daemon would report it for a
// container created from imageRef under name. The reported image is the id the
// daemon resolves the reference to, which differs from the creation reference
// of Config as soon as the tag is moved to another image.
func fakeContainer(id, name, imageRef string, mounts ...dockercontainer.MountPoint) dockercontainer.InspectResponse {
	return dockercontainer.InspectResponse{
		ID:     id,
		Name:   "/" + name,
		Image:  "sha256:" + id,
		Mounts: mounts,
		Config: &dockercontainer.Config{Image: imageRef},
	}
}

func stapelDaemonContext(handler http.Handler) context.Context {
	server := httptest.NewServer(handler)
	ginkgo.DeferCleanup(server.Close)
	for _, key := range []string{"DOCKER_CONTEXT", "DOCKER_CONFIG", "DOCKER_TLS", "DOCKER_TLS_VERIFY", "DOCKER_CERT_PATH", "DOCKER_API_VERSION"} {
		ginkgo.GinkgoT().Setenv(key, "")
	}
	ginkgo.GinkgoT().Setenv("DOCKER_HOST", "tcp://"+server.Listener.Addr().String())
	ctx, err := docker.NewContextWithStreams(context.Background(), io.Discard, io.Discard)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return ctx
}
