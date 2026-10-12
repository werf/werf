package manager

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/werf"
)

var _ = ginkgo.Describe("meta storage routing", func() {
	ginkgo.It("routes metadata to meta storage and aliases to stages", func() {
		gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), "")).To(gomega.Succeed())

		stages := &routingFakeStorage{}
		meta := &routingFakeStorage{}

		sm := &StorageManager{
			ProjectName:   "proj",
			StagesStorage: stages,
			MetaStorage:   meta,
		}
		ctx := context.Background()

		gomega.Expect(sm.ForEachRmImageMetadata(ctx, "proj", "img1", map[string][]string{
			"stage-id-1": {"commit-a"},
		}, func(_ context.Context, _, _ string, err error) error { return err })).To(gomega.Succeed())

		gomega.Expect(sm.ForEachRmManagedImage(ctx, "proj", []string{"mng1", "mng2"},
			func(_ context.Context, _ string, err error) error { return err })).To(gomega.Succeed())

		gomega.Expect(sm.ForEachDeleteStageCustomTag(ctx, []string{"tag1"},
			func(_ context.Context, _ string, err error) error { return err })).To(gomega.Succeed())

		gomega.Expect(sm.ForEachGetStageCustomTagMetadata(ctx, []string{"id1"},
			func(_ context.Context, _ string, _ *storage.CustomTagMetadata, err error) error { return err })).To(gomega.Succeed())

		gomega.Expect(meta.rmImageMetadata).To(gomega.Equal([]string{"img1/commit-a/stage-id-1"}))
		gomega.Expect(stages.rmImageMetadata).To(gomega.BeEmpty())

		gomega.Expect(meta.rmManagedImage).To(gomega.ConsistOf([]string{"mng1", "mng2"}))
		gomega.Expect(stages.rmManagedImage).To(gomega.BeEmpty())

		gomega.Expect(stages.deleteCustomTag).To(gomega.Equal([]string{"tag1"}))
		gomega.Expect(meta.deleteCustomTag).To(gomega.BeEmpty())

		gomega.Expect(meta.unregisterCustomTag).To(gomega.Equal([]string{"tag1"}))
		gomega.Expect(stages.unregisterCustomTag).To(gomega.BeEmpty())

		gomega.Expect(meta.getCustomTagMeta).To(gomega.Equal([]string{"id1"}))
		gomega.Expect(stages.getCustomTagMeta).To(gomega.BeEmpty())
	})

	ginkgo.It("preserves default stages routing", func() {
		gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), "")).To(gomega.Succeed())

		stages := &routingFakeStorage{}
		sm := &StorageManager{
			ProjectName:   "proj",
			StagesStorage: stages,
		}

		gomega.Expect(sm.GetMetaStorage(context.Background())).To(gomega.BeIdenticalTo(stages))

		ctx := context.Background()
		gomega.Expect(sm.ForEachRmManagedImage(ctx, "proj", []string{"mng1"},
			func(_ context.Context, _ string, err error) error { return err })).To(gomega.Succeed())

		gomega.Expect(stages.rmManagedImage).To(gomega.Equal([]string{"mng1"}))
	})
})
