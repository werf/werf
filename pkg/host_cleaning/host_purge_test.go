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
		daemon := &fakedockerd.Daemon{Containers: purgeFixture()}
		dockerCtx := fakedockerd.NewContext(ctx, daemon)

		gomega.Expect(HostPurge(dockerCtx, newTestDockerServerBackend(), HostPurgeOptions{})).To(gomega.Succeed())

		gomega.Expect(daemon.RemovedContainers(ctx)).To(gomega.ConsistOf("stapel-current", "stapel-old", "build-container", "import-server"))
		gomega.Expect(daemon.RemovedVolumes(ctx)).To(gomega.ConsistOf("vol-current", "vol-old"))

		requests := daemon.Requests(ctx)
		gomega.Expect(indexOfRequest(requests, "DELETE /containers/build-container?force=1")).
			To(gomega.BeNumerically("<", indexOfRequest(requests, "DELETE /volumes/vol-current")))
		gomega.Expect(indexOfRequest(requests, "DELETE /containers/import-server?force=1")).
			To(gomega.BeNumerically("<", indexOfRequest(requests, "DELETE /volumes/vol-old")))
	})

	ginkgo.It("keeps the volumes werf does not own", func() {
		ctx := context.Background()
		initWerfHomeDirs()
		containers := purgeFixture()
		containers[0].Mounts = append(containers[0].Mounts, dockercontainer.MountPoint{
			Type: "volume", Name: "vol-userdata", Destination: "/data",
		})
		daemon := &fakedockerd.Daemon{Containers: containers}
		dockerCtx := fakedockerd.NewContext(ctx, daemon)

		gomega.Expect(HostPurge(dockerCtx, newTestDockerServerBackend(), HostPurgeOptions{})).To(gomega.Succeed())

		gomega.Expect(daemon.RemovedVolumes(ctx)).NotTo(gomega.ContainElement("vol-userdata"))
	})

	ginkgo.It("mutates nothing in dry run mode", func() {
		ctx := context.Background()
		initWerfHomeDirs()
		daemon := &fakedockerd.Daemon{Containers: purgeFixture()}
		dockerCtx := fakedockerd.NewContext(ctx, daemon)

		gomega.Expect(HostPurge(dockerCtx, newTestDockerServerBackend(), HostPurgeOptions{DryRun: true})).To(gomega.Succeed())

		gomega.Expect(daemon.RemovedContainers(ctx)).To(gomega.BeEmpty())
		gomega.Expect(daemon.RemovedVolumes(ctx)).To(gomega.BeEmpty())
	})
})
