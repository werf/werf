package cleaning

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/storage/manager"
	"github.com/werf/werf/v2/pkg/werf"
)

var _ = ginkgo.Describe("Purge meta-repo safeguard", func() {
	ginkgo.DescribeTable("dry run preserves the safeguard and metadata",
		func(withMetadata bool) {
			ctx := context.Background()
			gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), ginkgo.GinkgoT().TempDir())).To(gomega.Succeed())
			registry := &purgeMarkerRegistry{images: map[string]*image.Info{}}
			stages := storage.NewRepoStagesStorage(&storage.NewRepoStagesStorageOptions{RepoAddress: "example.com/stages", DockerRegistry: registry, SkipMetaCheck: true})
			meta := storage.NewRepoStagesStorage(&storage.NewRepoStagesStorageOptions{RepoAddress: "example.com/meta", DockerRegistry: registry, SkipMetaCheck: true})
			gomega.Expect(stages.PutMetaRepoMarker(ctx, "project", meta.Address())).To(gomega.Succeed())
			if withMetadata {
				gomega.Expect(meta.AddManagedImage(ctx, "project", "app")).To(gomega.Succeed())
			}
			beforeStages, err := registry.Tags(ctx, stages.Address())
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			beforeMeta, err := registry.Tags(ctx, meta.Address())
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			sm := &manager.StorageManager{ProjectName: "project", StagesStorage: stages, MetaStorage: meta}
			gomega.Expect(Purge(ctx, "project", sm, PurgeOptions{DryRun: true})).To(gomega.Succeed())
			afterStages, err := registry.Tags(ctx, stages.Address())
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			afterMeta, err := registry.Tags(ctx, meta.Address())
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(afterStages).To(gomega.ConsistOf(beforeStages))
			gomega.Expect(afterMeta).To(gomega.ConsistOf(beforeMeta))
		},
		ginkgo.Entry("empty meta repository", false),
		ginkgo.Entry("populated meta repository", true),
	)

	ginkgo.DescribeTable("retains the marker after registry deletion errors and removes it after a successful retry",
		func(record string) {
			ctx := context.Background()
			gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), ginkgo.GinkgoT().TempDir())).To(gomega.Succeed())
			registry := &purgeMarkerRegistry{images: map[string]*image.Info{}, failMetadataDelete: true}
			stages := storage.NewRepoStagesStorage(&storage.NewRepoStagesStorageOptions{RepoAddress: "example.com/stages", DockerRegistry: registry, SkipMetaCheck: true})
			meta := storage.NewRepoStagesStorage(&storage.NewRepoStagesStorageOptions{RepoAddress: "example.com/meta", DockerRegistry: registry, SkipMetaCheck: true})
			guarded, err := storage.SetupMetaRepoSafeguard(ctx, "project", stages, meta, false)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			switch record {
			case "managed":
				gomega.Expect(guarded.AddManagedImage(ctx, "project", "app")).To(gomega.Succeed())
			case "image":
				gomega.Expect(guarded.PutImageMetadata(ctx, "project", "app", "abcdef", "ff0011-1748001122334")).To(gomega.Succeed())
			case "custom tag":
				gomega.Expect(guarded.RegisterStageCustomTag(ctx, "project", newTestStageDesc("example.com/stages", image.NewStageID("ff0011", 1748001122334)), "release")).To(gomega.Succeed())
			}
			sm := &manager.StorageManager{ProjectName: "project", StagesStorage: stages, MetaStorage: guarded}
			gomega.Expect(Purge(ctx, "project", sm, PurgeOptions{})).To(gomega.MatchError(gomega.ContainSubstring("metadata remains in meta-repo")))
			gomega.Expect(registry.deletionFailures).To(gomega.BeNumerically(">", 0))
			_, found, err := stages.GetMetaRepoMarker(ctx, "project")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(found).To(gomega.BeTrue())
			registry.failMetadataDelete = false
			gomega.Expect(Purge(ctx, "project", sm, PurgeOptions{})).To(gomega.Succeed())
			_, found, err = stages.GetMetaRepoMarker(ctx, "project")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(found).To(gomega.BeFalse())
		},
		ginkgo.Entry("managed-image records", "managed"),
		ginkgo.Entry("image metadata", "image"),
		ginkgo.Entry("custom-tag metadata", "custom tag"),
	)
})
