package host_cleaning

import (
	"runtime"
	"slices"

	dockercontainer "github.com/moby/moby/api/types/container"
	dockerimage "github.com/moby/moby/api/types/image"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/werf/werf/v3/pkg/container_backend"
	"github.com/werf/werf/v3/pkg/image"
	"github.com/werf/werf/v3/pkg/stapel"
	"github.com/werf/werf/v3/pkg/werf"
	"github.com/werf/werf/v3/test/mock"
)

// initWerfHomeDirs points the werf home and tmp dirs at directories of the
// current spec, so that a purge running to the end touches nothing else.
func initWerfHomeDirs() {
	ginkgo.GinkgoT().Setenv("WERF_HOME", ginkgo.GinkgoT().TempDir())
	ginkgo.GinkgoT().Setenv("WERF_TMP_DIR", ginkgo.GinkgoT().TempDir())
	gomega.Expect(werf.Init("", "")).To(gomega.Succeed())
}

// purgeFixture is a host with the stapel containers of the current and of a
// previous werf release, each holding a volume that a leftover werf container
// created with --volumes-from keeps in use.
func purgeFixture() []dockercontainer.InspectResponse {
	return []dockercontainer.InspectResponse{
		stapelFixtureContainer("stapel-current", stapel.VERSION, "_linux_amd64", "vol-current"),
		stapelFixtureContainer("stapel-old", "0.6.2", "", "vol-old"),
		consumerFixtureContainer("build-container", image.StageContainerNamePrefix+"project-stage", "vol-current"),
		consumerFixtureContainer("import-server", image.ImportServerContainerNamePrefix+"project", "vol-old"),
	}
}

func newTestDockerServerBackend() *container_backend.DockerServerBackend {
	return container_backend.NewDockerServerBackend(mock.NewMockLocker(gomock.NewController(ginkgo.GinkgoT())))
}

func stapelFixtureContainer(id, version, platformSuffix, volumeName string) dockercontainer.InspectResponse {
	return dockercontainer.InspectResponse{
		ID:    id,
		Name:  "/" + image.AssemblingContainerNamePrefix + version + platformSuffix,
		Image: "sha256:" + id,
		Mounts: []dockercontainer.MountPoint{
			{Type: "volume", Name: volumeName, Destination: "/.werf/stapel"},
		},
		Config: &dockercontainer.Config{Image: stapel.IMAGE + ":" + version},
	}
}

// consumerFixtureContainer is a running werf container that keeps the stapel
// volume in use through --volumes-from, the way a leftover build or import
// container does.
func consumerFixtureContainer(id, name, volumeName string) dockercontainer.InspectResponse {
	return dockercontainer.InspectResponse{
		ID:    id,
		Name:  "/" + name,
		Image: "sha256:" + id,
		State: &dockercontainer.State{Running: true},
		Mounts: []dockercontainer.MountPoint{
			{Type: "volume", Name: volumeName, Destination: "/.werf/stapel"},
		},
		Config: &dockercontainer.Config{Image: "registry.example.com/project:latest"},
	}
}

func indexOfRequest(requests []string, request string) int {
	index := slices.Index(requests, request)
	gomega.ExpectWithOffset(1, index).NotTo(gomega.Equal(-1), "request %q was never served", request)
	return index
}

func purgeFixtureImages() map[string]dockerimage.InspectResponse {
	return map[string]dockerimage.InspectResponse{
		stapel.ImageName(): {ID: "sha256:stapel-cleanup", Os: "linux", Architecture: runtime.GOARCH},
	}
}
