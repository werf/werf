package manager

import (
	"encoding/json"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/logging"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/werf"
)

var _ = ginkgo.DescribeTable("Stage lookup with an invalid config size",
	func(ctx ginkgo.SpecContext, configSize int64) {
		ctxWithLogger := logging.WithLogger(ctx)
		gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), "")).To(gomega.Succeed())
		previousCache := image.CommonManifestCache
		image.CommonManifestCache = image.NewManifestCache(ginkgo.GinkgoT().TempDir())
		ginkgo.DeferCleanup(func() { image.CommonManifestCache = previousCache })
		manager := newManifestStorageManager(ctxWithLogger)
		brokenID := image.NewStageID(missingStageDigest, 100)
		brokenName := manager.StagesStorage.ConstructStageImageName(manager.ProjectName, brokenID.Digest, brokenID.CreationTs)
		brokenRef, err := name.NewTag(brokenName, name.Insecure)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(remote.Write(brokenRef, empty.Image, remote.WithContext(ctx))).To(gomega.Succeed())
		manifest, err := empty.Image.Manifest()
		gomega.Expect(err).To(gomega.Succeed())
		manifest = manifest.DeepCopy()
		manifest.Config.Size = configSize
		manifestJSON, err := json.Marshal(manifest)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(remote.Put(brokenRef, &remote.Descriptor{Manifest: manifestJSON}, remote.WithContext(ctx))).To(gomega.Succeed())

		repo := manager.StagesStorage.(*storage.RepoStagesStorage)
		_, err = repo.DockerRegistry.GetRepoImage(ctxWithLogger, brokenName)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(docker_registry.IsBrokenImageError(err)).To(gomega.BeTrue(), "error: %v", err)
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("invalid size %d", configSize))
		_, err = repo.GetStageDesc(ctxWithLogger, manager.ProjectName, *brokenID)
		gomega.Expect(err).To(gomega.MatchError(storage.ErrBrokenImage))

		stages, err := manager.GetStageDescSetByDigest(ctxWithLogger, "stage", brokenID.Digest, 0)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(stages.IsEmpty()).To(gomega.BeTrue())

		healthyName := repo.ConstructStageImageName(manager.ProjectName, brokenID.Digest, 200)
		healthyRef, err := name.NewTag(healthyName, name.Insecure)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(remote.Write(healthyRef, empty.Image, remote.WithContext(ctx))).To(gomega.Succeed())
		stages, err = manager.GetStageDescSetByDigest(ctxWithLogger, "stage", brokenID.Digest, 0)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(stages.Cardinality()).To(gomega.Equal(1))
		gomega.Expect(stages.ToSlice()[0].Info.Name).To(gomega.Equal(healthyName))
	},
	ginkgo.Entry("zero size", int64(0)),
	ginkgo.Entry("negative size", int64(-1)),
)
