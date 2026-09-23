package build

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/require"

	buildImage "github.com/werf/werf/v2/pkg/build/image"
	"github.com/werf/werf/v2/pkg/build/stage"
	"github.com/werf/werf/v2/pkg/config"
	imagePkg "github.com/werf/werf/v2/pkg/image"
)

type anchorStub struct {
	*stage.BaseStage
	selectFn func(context.Context, stage.Conveyor, imagePkg.StageDescSet) (*imagePkg.StageDesc, error)
}

func (s *anchorStub) SelectSuitableStageDesc(ctx context.Context, c stage.Conveyor, set imagePkg.StageDescSet) (*imagePkg.StageDesc, error) {
	if s.selectFn != nil {
		return s.selectFn(ctx, c, set)
	}
	return s.BaseStage.SelectSuitableStageDesc(ctx, c, set)
}

func TestAnchorSelector_HonoursParentTsToggle(t *testing.T) {
	phase := &BuildPhase{StagesIterator: NewStagesIterator(nil)}

	anchor := &anchorStub{BaseStage: stage.NewBaseStage(stage.ImageSpec, &stage.BaseStageOptions{})}
	anchor.SetContentAnchor(true)
	require.Equal(t, int64(0), phase.getPrevNonEmptyStageCreationTsForStage(anchor),
		"anchor stage must resolve with parentStageCreationTs=0")

	nonAnchor := &anchorStub{BaseStage: stage.NewBaseStage(stage.Setup, &stage.BaseStageOptions{})}
	require.Equal(t, int64(0), phase.getPrevNonEmptyStageCreationTsForStage(nonAnchor),
		"non-anchor branch resolves through the iterator (nil PrevNonEmptyStage → 0)")
}

var _ = Describe("Content anchor inputs", func() {
	It("defers anchors that need resolved dependency values", func() {
		img := newImage("app", buildImage.NoBaseImage, buildImage.ImageOptions{RequiresResolvedDependencyInputs: true})
		Expect(canCalculateAnchorDigest(img, nil)).To(BeFalse())
	})

	It("defers anchors until dependency anchors are available", func() {
		img := newImage("app", buildImage.NoBaseImage, buildImage.ImageOptions{})
		dep := newImage("base", buildImage.NoBaseImage, buildImage.ImageOptions{})
		Expect(canCalculateAnchorDigest(img, []*buildImage.Image{dep})).To(BeFalse())

		dep.SetAnchorDigest("anchor")
		Expect(canCalculateAnchorDigest(img, []*buildImage.Image{dep})).To(BeTrue())
	})

	It("folds dependency anchor digests into the consumer", func(ctx SpecContext) {
		img := newImage("app", buildImage.NoBaseImage, buildImage.ImageOptions{IsFinal: true})
		dep := newImage("base", buildImage.NoBaseImage, buildImage.ImageOptions{})

		dep.SetAnchorDigest("anchor-digest-v1")
		before, err := collectHolisticInputs(ctx, img, []*buildImage.Image{dep}, nil, nil, false)
		Expect(err).To(Succeed())

		dep.SetAnchorDigest("anchor-digest-v2")
		after, err := collectHolisticInputs(ctx, img, []*buildImage.Image{dep}, nil, nil, false)
		Expect(err).To(Succeed())
		Expect(after).NotTo(Equal(before))

		dep.SetAnchorDigest("")
		_, err = collectHolisticInputs(ctx, img, []*buildImage.Image{dep}, nil, nil, false)
		Expect(err).To(MatchError(ContainSubstring("no content-based digest")))
	})

	It("folds per-stage content dependencies into the consumer", func(ctx SpecContext) {
		img := newImage("app", buildImage.NoBaseImage, buildImage.ImageOptions{IsFinal: true})
		install := newContentDependenciesStub(stage.Install, "install-deps")
		setup := newContentDependenciesStub(stage.Setup, "setup-deps")
		img.SetStages([]stage.Interface{install, setup, newContentDependenciesStub(stage.GitCache, "")})

		inputs, err := collectHolisticInputs(ctx, img, nil, nil, nil, false)
		Expect(err).To(Succeed())
		Expect(inputs).To(ContainElements("install:install-deps", "setup:setup-deps"))
		Expect(inputs).NotTo(ContainElement(ContainSubstring("gitCache")))

		setup.deps = "setup-deps-changed"
		changed, err := collectHolisticInputs(ctx, img, nil, nil, nil, false)
		Expect(err).To(Succeed())
		Expect(changed).NotTo(Equal(inputs))

		setup.err = errors.New("content deps failure")
		_, err = collectHolisticInputs(ctx, img, nil, nil, nil, false)
		Expect(err).To(MatchError(ContainSubstring(`stage "setup" GetContentDependencies: content deps failure`)))
	})

	It("folds Dockerfile dependency import types into the consumer", func(ctx SpecContext) {
		newDockerfileImage := func(importType config.DependencyImportType) *buildImage.Image {
			return newImage("app", buildImage.NoBaseImage, buildImage.ImageOptions{
				IsFinal:           true,
				IsDockerfileImage: true,
				DockerfileImageConfig: &config.ImageFromDockerfile{
					Staged: true,
					Dependencies: []*config.Dependency{
						{From: "base", Imports: []*config.DependencyImport{{Type: importType, TargetBuildArg: "BASE_IMAGE"}}},
					},
				},
			})
		}

		dep := newImage("base", buildImage.NoBaseImage, buildImage.ImageOptions{})
		dep.SetAnchorDigest("anchor-digest")

		byImageName, err := collectHolisticInputs(ctx, newDockerfileImage(config.ImageNameImport), []*buildImage.Image{dep}, nil, nil, false)
		Expect(err).To(Succeed())
		byImageRepo, err := collectHolisticInputs(ctx, newDockerfileImage(config.ImageRepoImport), []*buildImage.Image{dep}, nil, nil, false)
		Expect(err).To(Succeed())
		Expect(byImageRepo).NotTo(Equal(byImageName))
	})

	It("folds the selected internal base role into the consumer", func(ctx SpecContext) {
		baseA := newImage("base-a", buildImage.NoBaseImage, buildImage.ImageOptions{})
		baseB := newImage("base-b", buildImage.NoBaseImage, buildImage.ImageOptions{})
		baseA.SetAnchorDigest("same-anchor")
		baseB.SetAnchorDigest("same-anchor")
		dependencies := []*buildImage.Image{baseA, baseB}

		fromA := newImage("app", buildImage.FromImage, buildImage.ImageOptions{BaseImageName: "base-a"})
		fromB := newImage("app", buildImage.FromImage, buildImage.ImageOptions{BaseImageName: "base-b"})

		inputsA, err := collectHolisticInputs(ctx, fromA, dependencies, nil, nil, false)
		Expect(err).To(Succeed())
		inputsB, err := collectHolisticInputs(ctx, fromB, dependencies, nil, nil, false)
		Expect(err).To(Succeed())
		Expect(inputsB).NotTo(Equal(inputsA))
	})

	DescribeTable("folds real dependency values into deferred anchors",
		func(ctx SpecContext, opts buildImage.ImageOptions) {
			img := newImage("app", buildImage.NoBaseImage, opts)
			dep := newImage("base", buildImage.NoBaseImage, buildImage.ImageOptions{})
			dep.SetAnchorDigest("same-anchor")

			inputsA, err := collectHolisticInputs(ctx, img, []*buildImage.Image{dep}, stage.NewConveyorStub(nil, map[string]string{"base": "repo-a:tag"}, map[string]string{"base": "digest"}), nil, true)
			Expect(err).To(Succeed())
			inputsB, err := collectHolisticInputs(ctx, img, []*buildImage.Image{dep}, stage.NewConveyorStub(nil, map[string]string{"base": "repo-b:tag"}, map[string]string{"base": "digest"}), nil, true)
			Expect(err).To(Succeed())
			Expect(inputsB).NotTo(Equal(inputsA))
		},
		Entry("Dockerfile build arg", buildImage.ImageOptions{
			IsDockerfileImage: true,
			DockerfileImageConfig: &config.ImageFromDockerfile{Dependencies: []*config.Dependency{
				{From: "base", Imports: []*config.DependencyImport{{Type: config.ImageNameImport, TargetBuildArg: "BASE_IMAGE"}}},
			}},
		}),
		Entry("Stapel environment", buildImage.ImageOptions{
			StapelImageConfig: &config.StapelImage{StapelImageBase: &config.StapelImageBase{Dependencies: []*config.Dependency{
				{From: "base", Imports: []*config.DependencyImport{{Type: config.ImageNameImport, TargetEnv: "BASE_IMAGE"}}},
			}}},
		}),
	)
})
