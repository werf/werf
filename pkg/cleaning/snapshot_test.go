package cleaning

import (
	"context"
	"errors"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/cleaning/stage_manager"
	"github.com/werf/werf/v3/pkg/cleanup_report"
	"github.com/werf/werf/v3/pkg/config"
)

var _ = ginkgo.Describe("Cleanup snapshot publication order", func() {
	ginkgo.DescribeTable("retains a stage published while final tags are being listed",
		func(splitMeta, dryRun bool) {
			ctx := context.Background()
			storageManager := newSnapshotStorageManager(splitMeta)
			cleanup := &cleanupManager{
				stageManager: stage_manager.NewManager(), StorageManager: storageManager,
				ProjectName: "project", ImageNameList: []string{"configured", "shared"}, DryRun: dryRun,
			}
			gomega.Expect(cleanup.init(ctx)).To(gomega.Succeed())
			gomega.Expect(cleanup.cleanupFinalStages(ctx)).To(gomega.Succeed())
			gomega.Expect(storageManager.deletedFinalStages).To(gomega.BeEmpty())
			gomega.Expect(cleanup.ImageNameList).To(gomega.Equal([]string{"configured", "published", "shared"}))
			gomega.Expect(storageManager.events[:3]).To(gomega.Equal([]string{"final", "managed", "primary"}))
			gomega.Expect(storageManager.metadataNames).To(gomega.Equal(cleanup.ImageNameList))
		},
		ginkgo.Entry("shared metadata repository", false, false),
		ginkgo.Entry("separate metadata repository", true, false),
		ginkgo.Entry("shared metadata dry-run", false, true),
		ginkgo.Entry("separate metadata dry-run", true, true),
	)

	ginkgo.It("still selects a genuine orphan for deletion", func() {
		ctx := context.Background()
		storageManager := newSnapshotStorageManager(false)
		storageManager.publish = false
		cleanup := &cleanupManager{stageManager: stage_manager.NewManager(), StorageManager: storageManager}
		gomega.Expect(cleanup.init(ctx)).To(gomega.Succeed())
		gomega.Expect(cleanup.cleanupFinalStages(ctx)).To(gomega.Succeed())
		gomega.Expect(storageManager.deletedFinalStages).To(gomega.HaveLen(1))
	})

	ginkgo.It("does not fetch managed names or primary stages after a final-list failure", func() {
		storageManager := newSnapshotStorageManager(false)
		storageManager.finalErr = errors.New("final unavailable")
		cleanup := &cleanupManager{stageManager: stage_manager.NewManager(), StorageManager: storageManager}
		gomega.Expect(cleanup.init(context.Background())).To(gomega.MatchError("final unavailable"))
		gomega.Expect(storageManager.events).To(gomega.Equal([]string{"final"}))
	})

	ginkgo.It("discovers published images with cleanup disabled without accessing final storage", func() {
		storageManager := newSnapshotStorageManager(false)
		storageManager.finalErr = errors.New("must not access final")
		cleanup := &cleanupManager{
			stageManager: stage_manager.NewManager(), StorageManager: storageManager,
			ImageNameList: []string{"configured"}, DryRun: true, ConfigMetaCleanup: config.MetaCleanup{DisableCleanup: true},
			report: cleanup_report.NewReport(context.Background(), "cleanup", true, "primary", cleanup_report.NewReportOptions{}),
		}
		gomega.Expect(cleanup.run(context.Background())).To(gomega.Succeed())
		gomega.Expect(cleanup.report.Deleted).To(gomega.ContainElement(cleanup_report.Item{Type: cleanup_report.ItemTypeManagedImage, ImageName: "published"}))
		gomega.Expect(storageManager.events).NotTo(gomega.ContainElement("final"))
	})
})
