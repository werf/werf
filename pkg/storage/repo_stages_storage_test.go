package storage

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/image"
)

type pushImageRegistryStub struct {
	docker_registry.Interface

	pushedRef  string
	pushedOpts *docker_registry.PushImageOptions
}

func (r *pushImageRegistryStub) PushImage(_ context.Context, reference string, opts *docker_registry.PushImageOptions) error {
	r.pushedRef = reference
	r.pushedOpts = opts
	return nil
}

var _ = Describe("RepoStagesStorage", func() {
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

	It("rejects unsupported target platform option", func(ctx SpecContext) {
		storage := &RepoStagesStorage{DockerRegistry: &pushImageRegistryStub{}}

		err := storage.PostManifest(ctx, "registry.example/project:tag", container_backend.PostManifestOpts{
			CommonOpts: container_backend.CommonOpts{TargetPlatform: "linux/amd64"},
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unsupported target platform"))
	})

	It("rejects unsupported manifests option", func(ctx SpecContext) {
		storage := &RepoStagesStorage{DockerRegistry: &pushImageRegistryStub{}}

		err := storage.PostManifest(ctx, "registry.example/project:tag", container_backend.PostManifestOpts{
			Manifests: []*image.Info{{Name: "registry.example/project:tag"}},
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unsupported manifests option"))
	})
})
