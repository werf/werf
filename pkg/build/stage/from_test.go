package stage

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/container_backend"
	imagePkg "github.com/werf/werf/v2/pkg/image"
)

var _ = Describe("FromStage", func() {
	DescribeTable("GetDependencies()",
		func(ctx SpecContext, data testDataFrom) {
			conveyor := NewConveyorStubForDependencies(NewGiterminismManagerStub(NewLocalGitRepoStub("9d8059842b6fde712c58315ca0ab4713d90761c0"), NewGiterminismInspectorStub()), make([]*TestDependency, 0))

			fromStage := &FromStage{
				fromImageName:         data.FromImageName,
				fromExternal:          data.FromExternal,
				baseImageRepoIdOrNone: data.BaseImageRepoIdOrNone,
				fromCacheVersion:      data.FromCacheVersion,
				imageCacheVersion:     data.ImageCacheVersion,
				fromScratch:           data.FromScratch,
				BaseStage:             NewBaseStage(From, &BaseStageOptions{}),
			}

			digest, err := fromStage.GetDependencies(ctx, conveyor, nil, nil, nil, nil)
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

				ExpectedDigest: "52a335b26821a21ae8a47eb3a36a1ab5f388dd44dd1ab369c0578c173d88a752",
			}),

		Entry("should calculate from stage digest with fromCacheVersion param",
			testDataFrom{
				FromCacheVersion: "from-cache-version",

				ExpectedDigest: "e6bb62452421c2e2ca95e9c5c88abf9aa5ccef94073de4aa4c84ba442f5da449",
			}),

		Entry("should calculate from stage digest with baseImageRepoIdOrNone param",
			testDataFrom{
				BaseImageRepoIdOrNone: "base-image-repo-id-or-none",

				ExpectedDigest: "3d805b3dc9b36b26bb95a7723399b2767548d080e31af2148f69130f70a0b66f",
			}),

		Entry("should calculate from stage digest with fromImageName param",
			testDataFrom{
				FromImageName: "from-image-or-artifact-image-name",

				ExpectedDigest: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			}),

		// The expected digest is what the pre-fix code produced when the build phase passed the
		// external base as prevImage, so this entry pins stage digests across the change.
		Entry("should calculate from stage digest for external base image",
			testDataFrom{
				FromImageName: externalBaseRef,
				FromExternal:  true,

				ExpectedDigest: externalBaseDigest,
			}),

		Entry("should calculate from stage digest for scratch base image",
			testDataFrom{
				FromScratch: true,

				ExpectedDigest: "5a9cb6b54ea56d52a69891af8c21afb73d3841611bbabb5d7c61312d81e6e041",
			}),
	)

	Describe("external base image identity", func() {
		newExternalFromStage := func(ref string) *FromStage {
			imageBaseConfig := &config.StapelImageBase{From: ref}
			imageBaseConfig.SetFromExternal()
			return GenerateFromStage(imageBaseConfig, "", "", &BaseStageOptions{})
		}

		newConveyor := func() Conveyor {
			return NewConveyorStubForDependencies(NewGiterminismManagerStub(NewLocalGitRepoStub("9d8059842b6fde712c58315ca0ab4713d90761c0"), NewGiterminismInspectorStub()), make([]*TestDependency, 0))
		}

		It("gives the content digest the same value as the stage digest", func(ctx SpecContext) {
			digest, err := newExternalFromStage(externalBaseRef).GetContentDependencies(ctx, newConveyor(), nil)
			Expect(err).To(Succeed())

			Expect(digest).To(Equal(externalBaseDigest),
				"content-anchor input must carry the external base the same way the stage digest does")
		})

		It("gives different content digests to from stages with different external bases", func(ctx SpecContext) {
			conveyor := newConveyor()

			digestOne, err := newExternalFromStage(externalBaseRef).GetContentDependencies(ctx, conveyor, nil)
			Expect(err).To(Succeed())

			digestTwo, err := newExternalFromStage("registry.example.com/factory@sha256:2222222222222222222222222222222222222222222222222222222222222222").GetContentDependencies(ctx, conveyor, nil)
			Expect(err).To(Succeed())

			Expect(digestOne).NotTo(Equal(digestTwo),
				"content-anchor input must not match a from stage built from a different external base image")
		})
	})

	Describe("scratch semantics", func() {
		It("marks scratch from stage as mutable and not buildable", func() {
			stage := GenerateFromStage(&config.StapelImageBase{From: "scratch"}, "", "", &BaseStageOptions{})

			Expect(stage.IsBuildable()).To(BeFalse())
			Expect(stage.IsMutable()).To(BeTrue())
			Expect(stage.HasPrevStage()).To(BeFalse())
		})

		It("keeps regular from stage buildable and not mutable", func() {
			stage := GenerateFromStage(&config.StapelImageBase{From: "alpine"}, "", "", &BaseStageOptions{})

			Expect(stage.IsBuildable()).To(BeTrue())
			Expect(stage.IsMutable()).To(BeFalse())
			Expect(stage.HasPrevStage()).To(BeFalse())
		})

		It("preserves existing commit label when head commit is empty", func(ctx SpecContext) {
			stage := GenerateFromStage(&config.StapelImageBase{From: "scratch"}, "", "", &BaseStageOptions{})
			stageImage := NewStageImage(NewContainerBackendStub(), "", newLegacyImageForFromScratchTests("scratch-stage"))
			stageImage.Image.SetBuildServiceLabels(map[string]string{
				imagePkg.WerfProjectRepoCommitLabel: "valid-commit",
			})

			conveyor := NewConveyorStubForDependencies(NewGiterminismManagerStub(NewLocalGitRepoStub(""), NewGiterminismInspectorStub()), nil)

			Expect(stage.PrepareImage(ctx, conveyor, NewContainerBackendStub(), nil, stageImage, nil)).To(Succeed())
			Expect(stageImage.Image.GetBuildServiceLabels()).To(Equal(map[string]string{
				imagePkg.WerfProjectRepoCommitLabel: "valid-commit",
			}))
		})
	})

	Describe("MutateImage()", func() {
		It("returns a clear error when storage does not support scratch manifest creation", func(ctx SpecContext) {
			stage := &FromStage{fromScratch: true, BaseStage: NewBaseStage(From, &BaseStageOptions{})}
			stageImage := NewStageImage(NewContainerBackendStub(), "scratch-stage", newLegacyImageForFromScratchTests("scratch-stage"))
			stageImage.Image.SetBuildServiceLabels(map[string]string{"werf-stage-content-digest": "digest"})

			err := stage.MutateImage(ctx, scratchManifestUnsupportedStorageStub{}, nil, stageImage)
			Expect(err).To(MatchError("scratch from stage storage does not support manifest creation"))
		})

		It("creates manifest first and then mutates the same stage image reference", func(ctx SpecContext) {
			stage := &FromStage{fromScratch: true, BaseStage: NewBaseStage(From, &BaseStageOptions{TargetPlatform: "linux/amd64"})}
			stageImage := NewStageImage(NewContainerBackendStub(), "scratch-stage", newLegacyImageForFromScratchTests("scratch-stage"))
			stageImage.Image.SetBuildServiceLabels(map[string]string{
				"werf":                              "project",
				"werf-stage-content-digest":         "digest",
				imagePkg.WerfProjectRepoCommitLabel: "commit",
			})

			storage := &scratchManifestStorageStub{}
			err := stage.MutateImage(ctx, storage, nil, stageImage)
			Expect(err).NotTo(HaveOccurred())

			Expect(storage.postManifestRef).To(Equal("scratch-stage"))
			Expect(storage.postManifestOpts.Labels).To(ContainElements(
				"werf=project",
				"werf-stage-content-digest=digest",
				fmt.Sprintf("%s=%s", imagePkg.WerfProjectRepoCommitLabel, "commit"),
			))
			Expect(storage.postManifestOpts.TargetPlatform).To(Equal("linux/amd64"))
			Expect(storage.mutateSrc).To(Equal("scratch-stage"))
			Expect(storage.mutateDest).To(Equal("scratch-stage"))
			Expect(storage.mutateConfig.Labels).To(Equal(map[string]string{
				"werf":                              "project",
				"werf-stage-content-digest":         "digest",
				imagePkg.WerfProjectRepoCommitLabel: "commit",
			}))
			Expect(storage.mutateStageImage).To(BeIdenticalTo(stageImage.Image))
		})
	})
})

const (
	externalBaseRef    = "registry.example.com/factory@sha256:1111111111111111111111111111111111111111111111111111111111111111"
	externalBaseDigest = "7148bcbd6a0899be6c6185bbc660dc857bcc9b55866c78189401dca5f9588a3e"
)

type testDataFrom struct {
	FromImageName         string
	FromExternal          bool
	BaseImageRepoIdOrNone string
	FromCacheVersion      string
	ImageCacheVersion     string
	FromScratch           bool

	ImageContentDigest string

	ExpectedDigest string
}

type scratchManifestUnsupportedStorageStub struct{}

func (scratchManifestUnsupportedStorageStub) MutateAndPushImage(context.Context, string, string, imagePkg.SpecConfig, container_backend.LegacyImageInterface) error {
	return nil
}

type scratchManifestStorageStub struct {
	postManifestRef  string
	postManifestOpts container_backend.PostManifestOpts
	mutateSrc        string
	mutateDest       string
	mutateConfig     imagePkg.SpecConfig
	mutateStageImage container_backend.LegacyImageInterface
}

func (s *scratchManifestStorageStub) PostManifest(_ context.Context, ref string, opts container_backend.PostManifestOpts) error {
	s.postManifestRef = ref
	s.postManifestOpts = opts
	return nil
}

func (s *scratchManifestStorageStub) MutateAndPushImage(_ context.Context, src, dest string, newConfig imagePkg.SpecConfig, stageImage container_backend.LegacyImageInterface) error {
	s.mutateSrc = src
	s.mutateDest = dest
	s.mutateConfig = newConfig
	s.mutateStageImage = stageImage
	return nil
}

type legacyImageForFromScratchTests struct {
	name               string
	buildServiceLabels map[string]string
	builtID            string
}

func newLegacyImageForFromScratchTests(name string) *legacyImageForFromScratchTests {
	return &legacyImageForFromScratchTests{name: name}
}

func (i *legacyImageForFromScratchTests) Name() string               { return i.name }
func (i *legacyImageForFromScratchTests) SetName(name string)        { i.name = name }
func (i *legacyImageForFromScratchTests) GetTargetPlatform() string  { return "" }
func (i *legacyImageForFromScratchTests) Pull(context.Context) error { panic("unexpected call") }
func (i *legacyImageForFromScratchTests) Push(context.Context) error { panic("unexpected call") }
func (i *legacyImageForFromScratchTests) SetBuildServiceLabels(labels map[string]string) {
	i.buildServiceLabels = labels
}

func (i *legacyImageForFromScratchTests) GetBuildServiceLabels() map[string]string {
	return i.buildServiceLabels
}

func (i *legacyImageForFromScratchTests) Container() container_backend.LegacyContainer {
	panic("unexpected call")
}

func (i *legacyImageForFromScratchTests) BuilderContainer() container_backend.LegacyBuilderContainer {
	panic("unexpected call")
}

func (i *legacyImageForFromScratchTests) SetCommitChangeOptions(container_backend.LegacyCommitChangeOptions) {
	panic("unexpected call")
}

func (i *legacyImageForFromScratchTests) Build(context.Context, container_backend.BuildOptions) error {
	panic("unexpected call")
}
func (i *legacyImageForFromScratchTests) SetBuiltID(builtID string)         { i.builtID = builtID }
func (i *legacyImageForFromScratchTests) BuiltID() string                   { return i.builtID }
func (i *legacyImageForFromScratchTests) Introspect(context.Context) error  { panic("unexpected call") }
func (i *legacyImageForFromScratchTests) SetInfo(*imagePkg.Info)            { panic("unexpected call") }
func (i *legacyImageForFromScratchTests) IsExistsLocally() bool             { panic("unexpected call") }
func (i *legacyImageForFromScratchTests) SetStageDesc(*imagePkg.StageDesc)  { panic("unexpected call") }
func (i *legacyImageForFromScratchTests) GetStageDesc() *imagePkg.StageDesc { panic("unexpected call") }
func (i *legacyImageForFromScratchTests) GetFinalStageDesc() *imagePkg.StageDesc {
	panic("unexpected call")
}

func (i *legacyImageForFromScratchTests) SetFinalStageDesc(*imagePkg.StageDesc) {
	panic("unexpected call")
}

func (i *legacyImageForFromScratchTests) GetCopy() container_backend.LegacyImageInterface {
	panic("unexpected call")
}
