package cleaning

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/cleaning/stage_manager"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/storage/manager"
)

var _ manager.StorageManagerInterface = (*fakeStorageManager)(nil)

func newCleanupManagerForImportsMetadataTest(ctx context.Context, protectedStageDescs ...*image.StageDesc) *cleanupManager {
	ginkgo.GinkgoHelper()
	sm := newFakeStorageManager()
	sm.stageDescSet = image.NewStageDescSet(protectedStageDescs...)

	stageManager := stage_manager.NewManager()
	gomega.Expect(stageManager.InitStageDescSet(ctx, sm)).To(gomega.Succeed())
	for _, stageDesc := range protectedStageDescs {
		stageManager.MarkStageDescAsProtected(stageDesc, stage_manager.ProtectionReasonImportSource, false)
	}

	return &cleanupManager{
		stageManager:   stageManager,
		StorageManager: sm,
		ProjectName:    "myproject",
		report:         newTestReport(),
	}
}

func newTestStageDesc(repository string, stageID *image.StageID) *image.StageDesc {
	return &image.StageDesc{
		StageID: stageID,
		Info: &image.Info{
			Repository: repository,
			Tag:        stageID.String(),
			Name:       repository + ":" + stageID.String(),
		},
	}
}

func (f *fakeStorageManager) ForEachDeleteFinalStage(ctx context.Context, _ manager.ForEachDeleteStageOptions, stages image.StageDescSet, onDelete func(context.Context, *image.StageDesc, error) error) error {
	for stage := range stages.Iter() {
		f.deletedFinalStages = append(f.deletedFinalStages, stage)
		if err := onDelete(ctx, stage, nil); err != nil {
			return err
		}
	}
	return nil
}
