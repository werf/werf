package stapel

import (
	"context"
	"net/http"

	dockercontainer "github.com/moby/moby/api/types/container"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/test/pkg/fakedockerd"
)

var _ = ginkgo.Describe("stapel purge", func() {
	ginkgo.DescribeTable("removes own containers with their stapel volumes only",
		func(containers []dockercontainer.InspectResponse, expectedContainers, expectedVolumes []string) {
			ctx := context.Background()
			daemon := &fakedockerd.Daemon{Containers: containers}
			dockerCtx := fakedockerd.NewContext(ctx, daemon)

			gomega.Expect(Purge(dockerCtx)).To(gomega.Succeed())

			gomega.Expect(daemon.RemovedContainers(ctx)).To(gomega.ConsistOf(expectedContainers))
			gomega.Expect(daemon.RemovedVolumes(ctx)).To(gomega.ConsistOf(expectedVolumes))
		},
		ginkgo.Entry("container without a platform suffix",
			[]dockercontainer.InspectResponse{
				fakeContainer("legacy", containerName(getVersion(), ""), ImageName(), stapelVolumeMount("vol-legacy")),
			},
			[]string{"legacy"},
			[]string{"vol-legacy"},
		),
		ginkgo.Entry("containers of every platform, including the variants",
			[]dockercontainer.InspectResponse{
				fakeContainer("amd64", containerName(getVersion(), "linux/amd64"), ImageName(), stapelVolumeMount("vol-amd64")),
				fakeContainer("arm64", containerName(getVersion(), "linux/arm64"), ImageName(), stapelVolumeMount("vol-arm64")),
				fakeContainer("arm64v8", containerName(getVersion(), "linux/arm64/v8"), ImageName(), stapelVolumeMount("vol-arm64v8")),
				fakeContainer("armv7", containerName(getVersion(), "linux/arm/v7"), ImageName(), stapelVolumeMount("vol-armv7")),
			},
			[]string{"amd64", "arm64", "arm64v8", "armv7"},
			[]string{"vol-amd64", "vol-arm64", "vol-arm64v8", "vol-armv7"},
		),
		ginkgo.Entry("containers of the stapel versions left behind by previous werf releases",
			[]dockercontainer.InspectResponse{
				fakeContainer("old", containerName("0.6.2", ""), getImage()+":0.6.2", stapelVolumeMount("vol-old")),
				fakeContainer("old-platform", containerName("0.7.0", "linux/amd64"), getImage()+":0.7.0", stapelVolumeMount("vol-old-platform")),
			},
			[]string{"old", "old-platform"},
			[]string{"vol-old", "vol-old-platform"},
		),
		ginkgo.Entry("container with an additional non-stapel volume",
			[]dockercontainer.InspectResponse{
				fakeContainer("extra", containerName(getVersion(), "linux/amd64"), ImageName(),
					stapelVolumeMount("vol-stapel"),
					dockercontainer.MountPoint{Type: "volume", Name: "vol-userdata", Destination: "/data"},
				),
			},
			[]string{"extra"},
			[]string{"vol-stapel"},
		),
		ginkgo.Entry("container whose name does not belong to the version of its image",
			[]dockercontainer.InspectResponse{
				fakeContainer("name-of-another-version", containerName("9.9.9", ""), ImageName(), stapelVolumeMount("vol-another-version")),
				fakeContainer("not-a-platform-suffix", containerName(getVersion(), "")+"_not-a-platform", ImageName(), stapelVolumeMount("vol-not-a-platform")),
			},
			nil,
			nil,
		),
		ginkgo.Entry("container created from another image",
			[]dockercontainer.InspectResponse{
				fakeContainer("other-image", containerName(getVersion(), "linux/amd64"), "registry.example.com/other:latest", stapelVolumeMount("vol-other-image")),
				fakeContainer("lookalike-image", containerName(getVersion(), "linux/amd64"), getImage()+"-fork:"+getVersion(), stapelVolumeMount("vol-lookalike-image")),
				fakeContainer("untagged-image", containerName(getVersion(), "linux/amd64"), getImage()+":", stapelVolumeMount("vol-untagged-image")),
			},
			nil,
			nil,
		),
		ginkgo.Entry("container without an actual volume at the stapel mount point",
			[]dockercontainer.InspectResponse{
				fakeContainer("no-mounts", containerName(getVersion(), "linux/amd64"), ImageName()),
				fakeContainer("wrong-destination", containerName(getVersion(), "linux/arm64"), ImageName(),
					dockercontainer.MountPoint{Type: "volume", Name: "vol-elsewhere", Destination: "/elsewhere"},
				),
				fakeContainer("bind-mount", containerName(getVersion(), "linux/arm/v7"), ImageName(),
					dockercontainer.MountPoint{Type: "bind", Name: "vol-bind", Destination: containerVolumeDestination},
				),
			},
			nil,
			nil,
		),
	)

	ginkgo.It("removes a container of a custom stapel version whose name contains underscores", func() {
		ginkgo.GinkgoT().Setenv("WERF_STAPEL_IMAGE_VERSION", "0_7_2-dev")

		ctx := context.Background()
		daemon := &fakedockerd.Daemon{
			Containers: []dockercontainer.InspectResponse{
				fakeContainer("custom", containerName(getVersion(), "linux/amd64"), ImageName(), stapelVolumeMount("vol-custom")),
			},
		}
		dockerCtx := fakedockerd.NewContext(ctx, daemon)

		gomega.Expect(Purge(dockerCtx)).To(gomega.Succeed())

		gomega.Expect(daemon.RemovedContainers(ctx)).To(gomega.ConsistOf("custom"))
		gomega.Expect(daemon.RemovedVolumes(ctx)).To(gomega.ConsistOf("vol-custom"))
	})

	ginkgo.It("ignores a container without a config", func() {
		ctx := context.Background()
		daemon := &fakedockerd.Daemon{
			Containers: []dockercontainer.InspectResponse{
				{
					ID:     "no-config",
					Name:   "/" + containerName(getVersion(), ""),
					Mounts: []dockercontainer.MountPoint{stapelVolumeMount("vol-no-config")},
				},
			},
		}
		dockerCtx := fakedockerd.NewContext(ctx, daemon)

		gomega.Expect(Purge(dockerCtx)).To(gomega.Succeed())

		gomega.Expect(daemon.RemovedContainers(ctx)).To(gomega.BeEmpty())
		gomega.Expect(daemon.RemovedVolumes(ctx)).To(gomega.BeEmpty())
	})

	ginkgo.It("keeps the volumes of a container that could not be removed", func() {
		ctx := context.Background()
		daemon := &fakedockerd.Daemon{
			Containers: []dockercontainer.InspectResponse{
				fakeContainer("running", containerName(getVersion(), ""), ImageName(), stapelVolumeMount("vol-running")),
			},
			ContainerRemoveStatus: map[string]int{"running": http.StatusConflict},
		}
		dockerCtx := fakedockerd.NewContext(ctx, daemon)

		gomega.Expect(Purge(dockerCtx)).To(gomega.MatchError(gomega.ContainSubstring("remove container running")))

		gomega.Expect(daemon.RemovedVolumes(ctx)).To(gomega.BeEmpty())
	})

	ginkgo.It("reports a stapel volume still held by a container werf does not own", func() {
		ctx := context.Background()
		daemon := &fakedockerd.Daemon{
			Containers: []dockercontainer.InspectResponse{
				fakeContainer("stapel", containerName(getVersion(), ""), ImageName(), stapelVolumeMount("vol-shared")),
				fakeContainer("foreign", "user-container", "registry.example.com/other:latest", stapelVolumeMount("vol-shared")),
			},
		}
		dockerCtx := fakedockerd.NewContext(ctx, daemon)

		gomega.Expect(Purge(dockerCtx)).To(gomega.MatchError(gomega.ContainSubstring("remove volume vol-shared")))

		gomega.Expect(daemon.RemovedContainers(ctx)).To(gomega.ConsistOf("stapel"))
		gomega.Expect(daemon.RemovedVolumes(ctx)).To(gomega.BeEmpty())
	})

	ginkgo.DescribeTable("tolerates objects that disappeared concurrently",
		func(daemon *fakedockerd.Daemon) {
			ctx := context.Background()
			daemon.Containers = []dockercontainer.InspectResponse{
				fakeContainer("gone", containerName(getVersion(), ""), ImageName(), stapelVolumeMount("vol-gone")),
			}
			dockerCtx := fakedockerd.NewContext(ctx, daemon)

			gomega.Expect(Purge(dockerCtx)).To(gomega.Succeed())
		},
		ginkgo.Entry("container gone before the inspect",
			&fakedockerd.Daemon{ContainerInspectStatus: map[string]int{"gone": http.StatusNotFound}},
		),
		ginkgo.Entry("container gone before the removal",
			&fakedockerd.Daemon{ContainerRemoveStatus: map[string]int{"gone": http.StatusNotFound}},
		),
		ginkgo.Entry("volume gone before the removal",
			&fakedockerd.Daemon{VolumeRemoveStatus: map[string]int{"vol-gone": http.StatusNotFound}},
		),
	)
})
