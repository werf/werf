package instruction_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/build/stage"
	"github.com/werf/werf/v2/pkg/build/stage/instruction"
)

var _ = DescribeTable("FROM digest",
	func(data *TestData) {
		ctx := context.Background()

		digest, err := data.Stage.GetDependencies(ctx, data.Conveyor, data.ContainerBackend, nil, data.StageImage, data.BuildContext)
		Expect(err).To(Succeed())

		fmt.Printf("calculated digest: %s\n", digest)
		fmt.Printf("expected digest: %s\n", data.ExpectedDigest)

		Expect(digest).To(Equal(data.ExpectedDigest))
	},

	Entry("FROM with empty options", NewTestData(
		instruction.NewFrom(
			"",
			"",
			&stage.BaseStageOptions{},
		),
		"8a54a5b1f4e60cfae40553461549416865ff5ee3531444285ddadb4eb8cb939d",
		TestDataOptions{},
	)),

	Entry("FROM with BaseImageReference option only ", NewTestData(
		instruction.NewFrom(
			"test-image-reference",
			"",
			&stage.BaseStageOptions{},
		),
		"37478b120f37e3acfce89dc834fff808ed63e317e824316197bedbbbe513d0bf",
		TestDataOptions{},
	)),

	Entry("FROM with BaseImageRepoDigest option only", NewTestData(
		instruction.NewFrom(
			"",
			"test-image-repo-digest",
			&stage.BaseStageOptions{},
		),
		"306fe2574cce78188720fc4341950df9f1fae8908854da6f7054f3c0c35aa132",
		TestDataOptions{},
	)),

	Entry("FROM with ImageCacheVersion option only", NewTestData(
		instruction.NewFrom(
			"",
			"",
			&stage.BaseStageOptions{
				ImageCacheVersion: "image-cache-version",
			},
		),
		"9de9441477c5d35efbc350204fe9b71ad80116bd17c77972ac34232fb23e7c0b",
		TestDataOptions{},
	)),

	Entry("FROM with full options", NewTestData(
		instruction.NewFrom(
			"test-image-reference",
			"test-image-repo-digest",
			&stage.BaseStageOptions{
				ImageCacheVersion: "image-cache-version",
			},
		),
		"05fc69b51fbf058a81c5822ba4c8d1aa952a75280a16654b03c02deb00a534bb",
		TestDataOptions{},
	)),
)
