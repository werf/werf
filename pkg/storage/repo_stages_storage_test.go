package storage

import (
	"context"
	"errors"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker_registry"
	registry_api "github.com/werf/werf/v2/pkg/docker_registry/api"
	"github.com/werf/werf/v2/pkg/image"
)

type pushImageRegistryStub struct {
	docker_registry.Interface

	pushedRef  string
	pushedOpts *docker_registry.PushImageOptions
	mutatedCF  *v1.ConfigFile
}

func (r *pushImageRegistryStub) PushImage(_ context.Context, reference string, opts *docker_registry.PushImageOptions) error {
	r.pushedRef = reference
	r.pushedOpts = opts
	return nil
}

func (r *pushImageRegistryStub) MutateAndPushImage(ctx context.Context, _, destinationReference string, opts ...registry_api.MutateOption) error {
	destRef, err := name.ParseReference(destinationReference)
	if err != nil {
		return err
	}

	mutatedImageOrIndex, _, err := registry_api.MutateImageOrIndex(ctx, registry_api.MutateImageOrIndexOpts{
		ImageOrIndex:      empty.Image,
		Dest:              destRef,
		IsDestRefByDigest: false,
		MutateOptions:     opts,
	})
	if err != nil {
		return err
	}

	mutatedImage, ok := mutatedImageOrIndex.(v1.Image)
	if !ok {
		return nil
	}

	r.mutatedCF, err = mutatedImage.ConfigFile()
	if err != nil {
		return err
	}

	return nil
}

var _ = Describe("RepoStagesStorage", func() {
	It("publishes image metadata without listing registry tags", func(ctx SpecContext) {
		registry := &metadataPushRegistry{pushImageRegistryStub: &pushImageRegistryStub{}}
		storage := &RepoStagesStorage{RepoAddress: "registry.example/project", DockerRegistry: registry}

		Expect(storage.PutImageMetadata(ctx, "project", "app", "commit", "stage")).To(Succeed())
		Expect(registry.pushedRef).To(Equal(makeRepoImageMetadataName(storage.RepoAddress, "app", "commit", "stage")))
		Expect(registry.pushedOpts.Labels).To(HaveKeyWithValue(image.WerfLabel, "project"))
	})

	DescribeTable("managed image name encoding", func(imageName, encodedName string) {
		Expect(slugImageName(imageName)).To(Equal(encodedName))
		Expect(unslugImageName(encodedName)).To(Equal(imageName))
	},
		Entry("plus signs", "libstdc++", "libstdc__plus____plus__"),
	)

	DescribeTable("custom tag check against the content-based tag",
		func(ctx SpecContext, customTagInfo *image.Info, registryErr error, expectedErr string) {
			registry := newFakeRegistry()
			const customTagRef = "registry.example/project:v1.0.0"
			registry.tryGetInfo[customTagRef] = customTagInfo
			if registryErr != nil {
				registry.tryGetErr[customTagRef] = registryErr
			}
			storage := &RepoStagesStorage{RepoAddress: "registry.example/project", DockerRegistry: registry}
			stageDesc := &image.StageDesc{
				StageID: image.NewStageID("deadbeef", 1700000000),
				Info:    &image.Info{ID: "sha256:content"},
			}

			err := storage.CheckStageCustomTag(ctx, stageDesc, "v1.0.0")
			if expectedErr == "" {
				Expect(err).NotTo(HaveOccurred())
				return
			}
			Expect(err).To(MatchError(ContainSubstring(expectedErr)))
		},
		Entry("same image ID passes", &image.Info{ID: "sha256:content"}, nil, ""),
		Entry("stale custom tag is rejected", &image.Info{ID: "sha256:stale"}, nil,
			`custom tag "v1.0.0" image must be the same as associated content-based tag "deadbeef-1700000000" image`),
		Entry("missing custom tag is rejected", nil, nil, `custom tag "v1.0.0" not found`),
		Entry("registry failure propagates", nil, errors.New("registry unavailable"), "registry unavailable"),
	)

	It("pushes a manifest-only image to the registry in PostManifest", func(ctx SpecContext) {
		registry := &pushImageRegistryStub{}
		storage := &RepoStagesStorage{DockerRegistry: registry}

		err := storage.PostManifest(ctx, "registry.example/project:tag", container_backend.PostManifestOpts{
			Labels: []string{"werf=project", "werf-stage-content-digest=digest"},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(registry.pushedRef).To(Equal("registry.example/project:tag"))
		Expect(registry.pushedOpts).NotTo(BeNil())
		Expect(registry.pushedOpts.Labels).To(Equal(map[string]string{
			"werf":                      "project",
			"werf-stage-content-digest": "digest",
		}))
	})

	It("fails on malformed manifest label specs", func(ctx SpecContext) {
		storage := &RepoStagesStorage{DockerRegistry: &pushImageRegistryStub{}}

		err := storage.PostManifest(ctx, "registry.example/project:tag", container_backend.PostManifestOpts{
			Labels: []string{"broken-label"},
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("expected KEY=VALUE"))
	})

	It("accepts target platform option for manifest creation", func(ctx SpecContext) {
		registry := &pushImageRegistryStub{}
		storage := &RepoStagesStorage{DockerRegistry: registry}

		err := storage.PostManifest(ctx, "registry.example/project:tag", container_backend.PostManifestOpts{
			CommonOpts: container_backend.CommonOpts{TargetPlatform: "linux/amd64"},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(registry.pushedRef).To(Equal("registry.example/project:tag"))
	})

	It("rejects unsupported manifests option", func(ctx SpecContext) {
		storage := &RepoStagesStorage{DockerRegistry: &pushImageRegistryStub{}}

		err := storage.PostManifest(ctx, "registry.example/project:tag", container_backend.PostManifestOpts{
			Manifests: []*image.Info{{Name: "registry.example/project:tag"}},
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unsupported manifests option"))
	})

	It("preserves target platform in repo-backed mutation", func(ctx SpecContext) {
		registry := &pushImageRegistryStub{}
		storage := &RepoStagesStorage{DockerRegistry: registry}
		stageImage := container_backend.NewLegacyStageImage(nil, "registry.example/project:tag", nil, "linux/amd64")

		err := storage.MutateAndPushImage(ctx, "registry.example/project:src", "registry.example/project:dest", image.SpecConfig{}, stageImage)
		Expect(err).NotTo(HaveOccurred())
		Expect(registry.mutatedCF).NotTo(BeNil())
		Expect(registry.mutatedCF.OS).To(Equal("linux"))
		Expect(registry.mutatedCF.Architecture).To(Equal("amd64"))
		Expect(registry.mutatedCF.Variant).To(BeEmpty())
	})
})
