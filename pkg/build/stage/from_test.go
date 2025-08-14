package stage

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/werf/werf/v2/test/mock"
)

var _ = Describe("FromStage", func() {
	DescribeTable("GetDependencies()",
		func(ctx SpecContext, data testDataFrom) {
			ctrl := gomock.NewController(GinkgoT())

			conveyor := NewConveyorStubForDependencies(NewGiterminismManagerStub(NewLocalGitRepoStub("9d8059842b6fde712c58315ca0ab4713d90761c0"), NewGiterminismInspectorStub()), make([]*TestDependency, 0))

			legacyImage := mock.NewMockLegacyImageInterface(ctrl)
			containerBackend := NewContainerBackendStub()

			prevImage := NewStageImage(containerBackend, "base-image", legacyImage)

			fromStage := &FromStage{
				fromImageOrArtifactImageName: data.FromImageOrArtifactImageName,
				baseImageRepoIdOrNone:        data.BaseImageRepoIdOrNone,
				fromCacheVersion:             data.FromCacheVersion,
				imageCacheVersion:            data.ImageCacheVersion,
				BaseStage:                    NewBaseStage(From, &BaseStageOptions{}),
			}

			if fromStage.fromImageOrArtifactImageName != "" {
				// do nothing
			} else {
				legacyImage.EXPECT().Name().Return(data.PrevImageImageName)
			}

			digest, err := fromStage.GetDependencies(ctx, conveyor, nil, prevImage, nil, nil)
			Expect(err).To(Succeed())

			Expect(digest).To(Equal(data.ExpectedDigest),
				fmt.Sprintf("\ncalculated digest: %s\nexpected digest: %s\n", digest, data.ExpectedDigest))
		},

		Entry("should calculate from stage digest without any param",
			testDataFrom{
				ExpectedDigest: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			}),

		Entry("should calculate from stage digest with imageCacheVersion param",
			testDataFrom{
				ImageCacheVersion: "image-cache-version",

				ExpectedDigest: "62cc7cbbeb4189a01f9071091675d14e56faffbb1cd910e7e26858546028ef8f",
			}),

		Entry("should calculate from stage digest with fromCacheVersion param",
			testDataFrom{
				FromCacheVersion: "from-cache-version",

				ExpectedDigest: "30a820396785223b2734a036e91697e727e16c01cd30fe64cbb04d81fbc6c1ae",
			}),

		Entry("should calculate from stage digest with baseImageRepoIdOrNone param",
			testDataFrom{
				BaseImageRepoIdOrNone: "base-image-repo-id-or-none",

				ExpectedDigest: "29e4de9b8f38c28e4fffb47f5a22f2c8ac76986cffd81133d5180586ebf85adf",
			}),

		Entry("should calculate from stage digest with fromImageOrArtifactImageName param",
			testDataFrom{
				FromImageOrArtifactImageName: "from-image-or-artifact-image-name",
				PrevImageImageName:           "prev-image-image-name",

				ExpectedDigest: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			}),
	)
})

type testDataFrom struct {
	FromImageOrArtifactImageName string
	BaseImageRepoIdOrNone        string
	FromCacheVersion             string
	ImageCacheVersion            string

	ImageContentDigest string
	PrevImageImageName string

	ExpectedDigest string
}
