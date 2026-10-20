package build

import (
	"net/http/httptest"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/build/stage"
	"github.com/werf/werf/v3/pkg/config"
	"github.com/werf/werf/v3/pkg/docker_registry"
	imagePkg "github.com/werf/werf/v3/pkg/image"
	"github.com/werf/werf/v3/pkg/storage"
	"github.com/werf/werf/v3/pkg/storage/manager"
	"github.com/werf/werf/v3/pkg/werf"
)

var _ = ginkgo.Describe("Require built images", func() {
	ginkgo.It("reuses a published stage despite a stale empty stages snapshot", func(ctx ginkgo.SpecContext) {
		server, _ := newPublicationLockServer()
		primary := &publicationStorage{desc: &imagePkg.StageDesc{
			StageID: imagePkg.NewStageID("shared-digest", 100),
			Info:    &imagePkg.Info{Name: "repo:shared-digest-100", Labels: map[string]string{imagePkg.WerfStageContentDigestLabel: "published-content"}},
		}}
		storageManager := &publicationStorageManager{primary: primary}
		phase, img, stg := newPublicationPhase(ctx, storageManager, server.URL, false, 10)
		phase.Conveyor.StorageManager = &requireLookupStorageManager{storageManager}
		phase.RequireBuiltImages = true
		reported := &reportedStage{buildableStage: &buildableStage{publicationStage: stg.(*publicationStage)}}

		gomega.Expect(phase.onImageStage(ctx, img, reported)).To(gomega.Succeed())
		gomega.Expect(reported.GetStageImage().Image.GetStageDesc()).To(gomega.Equal(primary.desc))
		gomega.Expect(primary.writes).To(gomega.BeZero())
		gomega.Expect(reported.GetStageImage().Builder.(*stageBuilderStub).builds).To(gomega.BeZero())
	})

	ginkgo.It("reuses a secondary stage published after the cached empty snapshot", func(ctx ginkgo.SpecContext) {
		gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), "")).To(gomega.Succeed())
		oldCache := imagePkg.CommonManifestCache
		imagePkg.CommonManifestCache = imagePkg.NewManifestCache(ginkgo.GinkgoT().TempDir())
		ginkgo.DeferCleanup(func() { imagePkg.CommonManifestCache = oldCache })
		registryServer := httptest.NewServer(registry.New())
		ginkgo.DeferCleanup(registryServer.Close)
		address := strings.TrimPrefix(registryServer.URL, "http://") + "/project"
		client, err := docker_registry.NewDockerRegistry(ctx, address, docker_registry.DefaultImplementationName, docker_registry.DockerRegistryOptions{InsecureRegistry: true})
		gomega.Expect(err).To(gomega.Succeed())
		secondary := storage.NewRepoStagesStorage(&storage.NewRepoStagesStorageOptions{RepoAddress: address, DockerRegistry: client})
		primary := &publicationStorage{}
		publicationManager := &publicationStorageManager{primary: primary}
		storageManager := &secondaryLookupStorageManager{publicationStorageManager: publicationManager, lookup: &manager.StorageManager{ProjectName: "publication-project", StagesStorage: secondary}, secondary: secondary}
		server, _ := newPublicationLockServer()
		phase, img, stg := newPublicationPhase(ctx, publicationManager, server.URL, false, 10)
		phase.Conveyor.StorageManager = storageManager
		phase.RequireBuiltImages = true
		reported := &reportedStage{buildableStage: &buildableStage{publicationStage: stg.(*publicationStage)}}
		_, cleanup, err := phase.calculateStage(ctx, img, reported)
		gomega.Expect(err).To(gomega.Succeed())
		if cleanup != nil {
			cleanup()
		}
		stageID := *imagePkg.NewStageID(reported.GetDigest(), 1611836746968)
		cached, err := storageManager.GetStageDescSetByDigestFromStagesStorageCached(ctx, "stage", stageID.Digest, 10, secondary)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(cached.IsEmpty()).To(gomega.BeTrue())
		containerImage, err := mutate.ConfigFile(empty.Image, &v1.ConfigFile{OS: "linux", Architecture: "amd64", Config: v1.Config{Labels: map[string]string{imagePkg.WerfStageContentDigestLabel: "published-content", imagePkg.WerfLabel: "publication-project"}}})
		gomega.Expect(err).To(gomega.Succeed())
		tag, err := name.NewTag(secondary.ConstructStageImageName("publication-project", stageID.Digest, stageID.CreationTs))
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(remote.Write(tag, containerImage, remote.WithContext(ctx))).To(gomega.Succeed())
		cached, err = storageManager.GetStageDescSetByDigestFromStagesStorageCached(ctx, "stage", stageID.Digest, 10, secondary)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(cached.IsEmpty()).To(gomega.BeTrue())

		gomega.Expect(phase.onImageStage(ctx, img, reported)).To(gomega.Succeed())
		gomega.Expect(primary.writes).To(gomega.Equal(1))
		gomega.Expect(primary.desc.StageID).To(gomega.Equal(&stageID))
		gomega.Expect(reported.GetContentDigest()).To(gomega.Equal("published-content"))
		gomega.Expect(reported.GetStageImage().Builder.(*stageBuilderStub).builds).To(gomega.BeZero())
	})

	ginkgo.DescribeTable("does not build a missing stage when built images are required",
		func(ctx ginkgo.SpecContext, require, dockerfile bool) {
			server, _ := newPublicationLockServer()
			primary := &publicationStorage{}
			phase, img, stg := newPublicationPhase(ctx, &publicationStorageManager{primary: primary}, server.URL, false, 10)
			phase.RequireBuiltImages = require
			phase.Conveyor.ContainerBackend = &publicationContainerBackend{}
			img.IsDockerfileImage = dockerfile
			img.DockerfileImageConfig = &config.ImageFromDockerfile{}
			reported := &reportedStage{buildableStage: &buildableStage{publicationStage: stg.(*publicationStage)}}
			img.SetStages([]stage.Interface{reported})

			err := phase.onImageStage(ctx, img, reported)
			if require {
				gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("build images first without --require-built-images")))
				gomega.Expect(primary.writes).To(gomega.BeZero())
				gomega.Expect(reported.GetStageImage().Builder).To(gomega.BeNil())
			} else {
				gomega.Expect(err).To(gomega.Succeed())
				gomega.Expect(primary.writes).To(gomega.Equal(1))
				gomega.Expect(reported.GetStageImage().Builder.(*stageBuilderStub).builds).To(gomega.Equal(1))
			}
		},
		ginkgo.Entry("require fails before invoking the Dockerfile builder", true, true),
		ginkgo.Entry("require fails before invoking the Stapel builder", true, false),
		ginkgo.Entry("ordinary Dockerfile build invokes the builder and publishes the stage", false, true),
		ginkgo.Entry("ordinary Stapel build invokes the builder and publishes the stage", false, false),
	)
})
