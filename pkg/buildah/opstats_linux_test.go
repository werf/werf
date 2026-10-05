package buildah

import (
	"context"
	"errors"
	"testing/iotest"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/opstats"
)

var _ = ginkgo.Describe("NativeBuildah operation stats", func() {
	// Every call below fails while resolving its arguments, before the containers storage is
	// touched, so the observation is the only thing being exercised here.
	const invalidPlatform = "bad platform"

	ginkgo.DescribeTable("records exactly one operation for a failed call",
		func(expectedOp opstats.Operation, call func(ctx context.Context, b *NativeBuildah) error) {
			collector := opstats.NewCollector()
			ctx := opstats.NewContext(context.Background(), collector)

			gomega.Expect(call(ctx, &NativeBuildah{})).To(gomega.HaveOccurred())

			summary := collector.Summary()
			gomega.Expect(summary).To(gomega.HaveLen(1))
			gomega.Expect(summary[0].Operation).To(gomega.Equal(expectedOp))
			gomega.Expect(summary[0].Count).To(gomega.Equal(1))
		},
		ginkgo.Entry("BuildFromDockerfile", opstats.Operation("buildah: image build"), func(ctx context.Context, b *NativeBuildah) error {
			_, err := b.BuildFromDockerfile(ctx, "Dockerfile", BuildFromDockerfileOpts{CommonOpts: CommonOpts{TargetPlatform: invalidPlatform}})
			return err
		}),
		ginkgo.Entry("FromCommand", opstats.Operation("buildah: container create"), func(ctx context.Context, b *NativeBuildah) error {
			_, err := b.FromCommand(ctx, "container", "image", FromCommandOpts{TargetPlatform: invalidPlatform})
			return err
		}),
		ginkgo.Entry("Pull", opstats.Operation("buildah: image pull"), func(ctx context.Context, b *NativeBuildah) error {
			_, err := b.Pull(ctx, "image", PullOpts{TargetPlatform: invalidPlatform})
			return err
		}),
		ginkgo.Entry("Tag", opstats.Operation("buildah: image tag"), func(ctx context.Context, b *NativeBuildah) error {
			return b.Tag(ctx, "image", "newImage", TagOpts{TargetPlatform: invalidPlatform})
		}),
		ginkgo.Entry("Rmi", opstats.Operation("buildah: image remove"), func(ctx context.Context, b *NativeBuildah) error {
			return b.Rmi(ctx, "image", RmiOpts{CommonOpts: CommonOpts{TargetPlatform: invalidPlatform}})
		}),
		ginkgo.Entry("PruneImages", opstats.Operation("buildah: image prune"), func(ctx context.Context, b *NativeBuildah) error {
			_, err := b.PruneImages(ctx, PruneImagesOptions{CommonOpts: CommonOpts{TargetPlatform: invalidPlatform}})
			return err
		}),
		ginkgo.Entry("Images", opstats.Operation("buildah: image list"), func(ctx context.Context, b *NativeBuildah) error {
			_, err := b.Images(ctx, ImagesOptions{CommitOpts: CommitOpts{CommonOpts: CommonOpts{TargetPlatform: invalidPlatform}}})
			return err
		}),
		ginkgo.Entry("LoadImageFromStream", opstats.Operation("buildah: image load"), func(ctx context.Context, b *NativeBuildah) error {
			_, err := b.LoadImageFromStream(ctx, iotest.ErrReader(errors.New("read failed")))
			return err
		}),
	)
})
