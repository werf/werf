package host_cleaning

import (
	"context"

	dockercontainer "github.com/moby/moby/api/types/container"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/test/pkg/fakedockerd"
)

var _ = ginkgo.Describe("HostPurge", func() {
	ginkgo.It("removes the stapel volumes of every version once nothing holds them in use", func() {
		ctx := context.Background()
		initWerfHomeDirs()
		daemon := &fakedockerd.Daemon{Containers: purgeFixture(), Images: purgeFixtureImages()}
		dockerCtx := fakedockerd.NewContext(ctx, daemon)

		gomega.Expect(HostPurge(dockerCtx, newTestDockerServerBackend(), HostPurgeOptions{})).To(gomega.Succeed())

		gomega.Expect(daemon.RemovedContainers(ctx)).To(gomega.ConsistOf("stapel-current", "stapel-old", "build-container", "import-server"))
		gomega.Expect(daemon.RemovedVolumes(ctx)).To(gomega.ConsistOf("vol-current", "vol-old"))

		requests := daemon.Requests(ctx)
		gomega.Expect(requests).To(gomega.ContainElement(gomega.HavePrefix("POST /containers/create")))
		gomega.Expect(daemon.Images).To(gomega.BeEmpty())
		gomega.Expect(indexOfRequest(requests, "DELETE /containers/build-container?force=1")).
			To(gomega.BeNumerically("<", indexOfRequest(requests, "DELETE /volumes/vol-current")))
		gomega.Expect(indexOfRequest(requests, "DELETE /containers/import-server?force=1")).
			To(gomega.BeNumerically("<", indexOfRequest(requests, "DELETE /volumes/vol-old")))
	})

	ginkgo.It("does not remove volumes through container deletion", func() {
		ctx := context.Background()
		initWerfHomeDirs()
		containers := purgeFixture()
		containers[0].Mounts = append(containers[0].Mounts, dockercontainer.MountPoint{
			Type: "volume", Name: "vol-userdata", Destination: "/data",
		})
		containers[2].Mounts = append(containers[2].Mounts, dockercontainer.MountPoint{
			Type: "volume", Name: "vol-buildcache", Destination: "/cache",
		})
		daemon := &fakedockerd.Daemon{
			Containers:       containers,
			Images:           purgeFixtureImages(),
			AnonymousVolumes: []string{"vol-userdata", "vol-buildcache"},
		}
		dockerCtx := fakedockerd.NewContext(ctx, daemon)

		gomega.Expect(HostPurge(dockerCtx, newTestDockerServerBackend(), HostPurgeOptions{})).To(gomega.Succeed())

		gomega.Expect(daemon.RemovedVolumes(ctx)).To(gomega.ConsistOf("vol-current", "vol-old"))
	})

	ginkgo.It("mutates nothing in dry run mode", func() {
		ctx := context.Background()
		initWerfHomeDirs()
		daemon := &fakedockerd.Daemon{Containers: purgeFixture(), Images: purgeFixtureImages()}
		dockerCtx := fakedockerd.NewContext(ctx, daemon)

		gomega.Expect(HostPurge(dockerCtx, newTestDockerServerBackend(), HostPurgeOptions{DryRun: true})).To(gomega.Succeed())

		gomega.Expect(daemon.RemovedContainers(ctx)).To(gomega.BeEmpty())
		gomega.Expect(daemon.RemovedVolumes(ctx)).To(gomega.BeEmpty())
	})
})
