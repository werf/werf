package stage

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/config"
)

var _ = Describe("Content dependencies of cross-image references", func() {
	newImport := func(from string) *config.Import {
		return &config.Import{
			From:   from,
			Export: &config.Export{ExportBase: &config.ExportBase{Add: "/app", To: "/app"}},
		}
	}

	newConveyor := func(baseStageID string) *ConveyorStub {
		return NewConveyorStub(
			NewGiterminismManagerStub(NewLocalGitRepoStub("9d8059842b6fde712c58315ca0ab4713d90761c0"), NewGiterminismInspectorStub()),
			map[string]string{"base": baseStageID},
			map[string]string{"base": baseStageID + "-repo-digest"},
		)
	}

	It("keeps the from stage content digest stable when the base image is rebuilt", func(ctx SpecContext) {
		stg := &FromStage{fromImageName: "base", BaseStage: NewBaseStage(From, &BaseStageOptions{})}

		before, err := stg.GetContentDependencies(ctx, newConveyor("digest-1"), nil)
		Expect(err).To(Succeed())
		after, err := stg.GetContentDependencies(ctx, newConveyor("digest-2"), nil)
		Expect(err).To(Succeed())
		Expect(after).To(Equal(before))

		stageDigestBefore, err := stg.GetDependencies(ctx, newConveyor("digest-1"), nil, nil, nil, nil)
		Expect(err).To(Succeed())
		stageDigestAfter, err := stg.GetDependencies(ctx, newConveyor("digest-2"), nil, nil, nil, nil)
		Expect(err).To(Succeed())
		Expect(stageDigestAfter).NotTo(Equal(stageDigestBefore), "stage digests still follow the rebuilt base image")

		other := &FromStage{fromImageName: "other-base", BaseStage: NewBaseStage(From, &BaseStageOptions{})}
		otherDigest, err := other.GetContentDependencies(ctx, newConveyor("digest-1"), nil)
		Expect(err).To(Succeed())
		Expect(otherDigest).NotTo(Equal(before), "a different base image is a different content")
	})

	It("keeps the dependencies stage content digest stable when the source images are rebuilt", func(ctx SpecContext) {
		stg := newDependenciesStage(
			[]*config.Import{newImport("base")},
			[]*config.Dependency{{From: "base", Imports: []*config.DependencyImport{{Type: config.ImageNameImport, TargetEnv: "BASE_IMAGE"}}}},
			DependenciesAfterInstall,
			&BaseStageOptions{},
		)

		before, err := stg.GetContentDependencies(ctx, newConveyor("digest-1"), nil)
		Expect(err).To(Succeed())
		after, err := stg.GetContentDependencies(ctx, newConveyor("digest-2"), nil)
		Expect(err).To(Succeed())
		Expect(after).To(Equal(before))

		stageDigestBefore, err := stg.GetDependencies(ctx, newConveyor("digest-1"), nil, nil, nil, nil)
		Expect(err).To(Succeed())
		stageDigestAfter, err := stg.GetDependencies(ctx, newConveyor("digest-2"), nil, nil, nil, nil)
		Expect(err).To(Succeed())
		Expect(stageDigestAfter).NotTo(Equal(stageDigestBefore), "stage digests still follow the rebuilt source images")

		otherSource := newDependenciesStage(
			[]*config.Import{newImport("other-base")},
			[]*config.Dependency{{From: "base", Imports: []*config.DependencyImport{{Type: config.ImageNameImport, TargetEnv: "BASE_IMAGE"}}}},
			DependenciesAfterInstall,
			&BaseStageOptions{},
		)
		otherDigest, err := otherSource.GetContentDependencies(ctx, newConveyor("digest-1"), nil)
		Expect(err).To(Succeed())
		Expect(otherDigest).NotTo(Equal(before), "a different source image is a different content")
	})

	It("resolves dependency build args to placeholders distinct per image and import type", func() {
		resolved := ResolveDependenciesArgsForContent([]*config.Dependency{
			{From: "base", Imports: []*config.DependencyImport{
				{Type: config.ImageNameImport, TargetBuildArg: "BASE_IMAGE"},
				{Type: config.ImageTagImport, TargetBuildArg: "BASE_TAG"},
			}},
			{From: "other-base", Imports: []*config.DependencyImport{
				{Type: config.ImageNameImport, TargetBuildArg: "OTHER_IMAGE"},
			}},
		})

		Expect(resolved).To(HaveLen(3))
		Expect(resolved["BASE_IMAGE"]).NotTo(Equal(resolved["BASE_TAG"]))
		Expect(resolved["BASE_IMAGE"]).NotTo(Equal(resolved["OTHER_IMAGE"]))
	})
})
