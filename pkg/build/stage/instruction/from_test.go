package instruction_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/build/stage"
	"github.com/werf/werf/v2/pkg/build/stage/instruction"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/dockerfile/frontend"
)

var _ = DescribeTable("FROM digest",
	func(ctx SpecContext, data *TestData) {
		digest, err := data.Stage.GetDependencies(ctx, data.Conveyor, data.ContainerBackend, nil, data.StageImage, data.BuildContext)
		Expect(err).To(Succeed())

		fmt.Printf("calculated digest: %s\n", digest)
		fmt.Printf("expected digest: %s\n", data.ExpectedDigest)

		Expect(digest).To(Equal(data.ExpectedDigest))
	},

	Entry("FROM with empty options", NewTestData(
		instruction.NewFrom("", "", "", nil, nil, &stage.BaseStageOptions{}),
		"8a54a5b1f4e60cfae40553461549416865ff5ee3531444285ddadb4eb8cb939d",
		TestDataOptions{},
	)),

	Entry("FROM with BaseImageReference option only ", NewTestData(
		instruction.NewFrom("test-image-reference", "", "", nil, nil, &stage.BaseStageOptions{}),
		"37478b120f37e3acfce89dc834fff808ed63e317e824316197bedbbbe513d0bf",
		TestDataOptions{},
	)),

	Entry("FROM with BaseImageRepoDigest option only", NewTestData(
		instruction.NewFrom("", "test-image-repo-digest", "", nil, nil, &stage.BaseStageOptions{}),
		"306fe2574cce78188720fc4341950df9f1fae8908854da6f7054f3c0c35aa132",
		TestDataOptions{},
	)),

	Entry("FROM with ImageCacheVersion option only", NewTestData(
		instruction.NewFrom("", "", "image-cache-version", nil, nil, &stage.BaseStageOptions{}),
		"9de9441477c5d35efbc350204fe9b71ad80116bd17c77972ac34232fb23e7c0b",
		TestDataOptions{},
	)),

	Entry("FROM with full options", NewTestData(
		instruction.NewFrom("test-image-reference", "test-image-repo-digest", "image-cache-version", nil, nil, &stage.BaseStageOptions{}),
		"05fc69b51fbf058a81c5822ba4c8d1aa952a75280a16654b03c02deb00a534bb",
		TestDataOptions{},
	)),
)

var _ = Describe("FROM dependency expansion", func() {
	newGiterminismManager := func() *stage.GiterminismManagerStub {
		return stage.NewGiterminismManagerStub(stage.NewLocalGitRepoStub("test"), stage.NewGiterminismInspectorStub())
	}

	deps := []*config.Dependency{
		{
			From: "ssr",
			Imports: []*config.DependencyImport{
				{Type: config.ImageNameImport, TargetBuildArg: "SSR_IMAGE"},
			},
		},
	}

	expanderFactory := frontend.NewShlexExpanderFactory('\\')

	It("should produce different digests when dependency image changes", func(ctx SpecContext) {
		conveyorV1 := stage.NewConveyorStub(
			newGiterminismManager(),
			map[string]string{"ssr": "registry.example.com/ssr:v1"}, nil,
		)
		fromV1 := instruction.NewFrom("${SSR_IMAGE}", "", "", deps, expanderFactory, &stage.BaseStageOptions{})
		Expect(fromV1.ExpandDependencies(ctx, conveyorV1, nil)).To(Succeed())
		Expect(fromV1.BaseImageReference).To(Equal("registry.example.com/ssr:v1"))
		digestV1, err := fromV1.GetDependencies(ctx, conveyorV1, nil, nil, nil, nil)
		Expect(err).To(Succeed())

		conveyorV2 := stage.NewConveyorStub(
			newGiterminismManager(),
			map[string]string{"ssr": "registry.example.com/ssr:v2"}, nil,
		)
		fromV2 := instruction.NewFrom("${SSR_IMAGE}", "", "", deps, expanderFactory, &stage.BaseStageOptions{})
		Expect(fromV2.ExpandDependencies(ctx, conveyorV2, nil)).To(Succeed())
		Expect(fromV2.BaseImageReference).To(Equal("registry.example.com/ssr:v2"))
		digestV2, err := fromV2.GetDependencies(ctx, conveyorV2, nil, nil, nil, nil)
		Expect(err).To(Succeed())

		Expect(digestV1).NotTo(Equal(digestV2))
	})

	It("should leave BaseImageReference unchanged when no expander factory provided", func(ctx SpecContext) {
		conveyor := stage.NewConveyorStub(
			newGiterminismManager(),
			map[string]string{"ssr": "registry.example.com/ssr:v1"}, nil,
		)
		from := instruction.NewFrom("${SSR_IMAGE}", "", "", deps, nil, &stage.BaseStageOptions{})
		Expect(from.ExpandDependencies(ctx, conveyor, nil)).To(Succeed())
		Expect(from.BaseImageReference).To(Equal("${SSR_IMAGE}"))
	})
})
