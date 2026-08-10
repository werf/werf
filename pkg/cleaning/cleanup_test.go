package cleaning

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/werf/werf/v2/pkg/cleaning/stage_manager"
	"github.com/werf/werf/v2/pkg/cleanup_report"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/storage/manager"
)

type fakePrimaryStagesStorage struct {
	storage.PrimaryStagesStorage

	mu sync.Mutex

	rejectedStageIDs []image.StageID
	rejectedErr      error

	deleteImageErrs  map[string]error
	deleteRecordErrs map[string]error
	deleteTagErrs    map[string]error

	deletedImages  []image.StageID
	deletedRecords []image.StageID
	deletedTags    []string
}

func (f *fakePrimaryStagesStorage) GetRejectedStageIDs(_ context.Context, _ ...storage.Option) ([]image.StageID, error) {
	return f.rejectedStageIDs, f.rejectedErr
}

func (f *fakePrimaryStagesStorage) DeleteRejectedStageImage(_ context.Context, stageID image.StageID, _ storage.DeleteImageOptions) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deletedImages = append(f.deletedImages, stageID)
	return f.deleteImageErrs[stageID.String()]
}

func (f *fakePrimaryStagesStorage) DeleteRejectedStageRecord(_ context.Context, stageID image.StageID, _ storage.DeleteImageOptions) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deletedRecords = append(f.deletedRecords, stageID)
	return f.deleteRecordErrs[stageID.String()]
}

func (f *fakePrimaryStagesStorage) DeleteStageCustomTag(_ context.Context, tag string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deletedTags = append(f.deletedTags, tag)
	return f.deleteTagErrs[tag]
}

type fakeStorageManager struct {
	manager.StorageManagerInterface

	stages *fakePrimaryStagesStorage

	importMetadataErrs map[string]error

	stageDescSet      image.StageDescSet
	finalStageDescSet image.StageDescSet
}

func newFakeStorageManager() *fakeStorageManager {
	return &fakeStorageManager{
		stages: &fakePrimaryStagesStorage{
			deleteImageErrs:  map[string]error{},
			deleteRecordErrs: map[string]error{},
			deleteTagErrs:    map[string]error{},
		},
		importMetadataErrs: map[string]error{},
	}
}

func (f *fakeStorageManager) GetStagesStorage() storage.PrimaryStagesStorage {
	return f.stages
}

func (f *fakeStorageManager) ForEachRejectedStage(ctx context.Context, stageIDs []image.StageID, cb func(ctx context.Context, stageID image.StageID) error) error {
	for _, id := range stageIDs {
		if err := cb(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func (f *fakeStorageManager) ForEachDeleteStageCustomTag(ctx context.Context, tags []string, cb func(ctx context.Context, tag string, err error) error) error {
	for _, tag := range tags {
		if err := cb(ctx, tag, f.stages.deleteTagErrs[tag]); err != nil {
			return err
		}
	}
	return nil
}

func (f *fakeStorageManager) ForEachRmImportMetadata(ctx context.Context, _ string, ids []string, cb func(ctx context.Context, id string, err error) error) error {
	for _, id := range ids {
		if err := cb(ctx, id, f.importMetadataErrs[id]); err != nil {
			return err
		}
	}
	return nil
}

func TestDeleteRejectedStagesWithLinkedTags_NoRejected(t *testing.T) {
	sm := newFakeStorageManager()

	deleted, err := deleteRejectedStagesWithLinkedTags(context.Background(), sm, nil, false, nil)
	require.NoError(t, err)
	assert.Empty(t, deleted)
	assert.Empty(t, sm.stages.deletedImages)
	assert.Empty(t, sm.stages.deletedTags)
	assert.Empty(t, sm.stages.deletedRecords)
}

func TestDeleteRejectedStagesWithLinkedTags_OrderStageThenTagsThenMarker(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	stageID := image.NewStageID(digest, 1700000000)
	otherStageID := image.NewStageID(digest, 1700000999)

	sm := newFakeStorageManager()
	sm.stages.rejectedStageIDs = []image.StageID{*stageID}

	customTagsByStageID := map[string][]string{
		stageID.String():      {"v1.0.0", "latest"},
		otherStageID.String(): {"unrelated"},
	}

	deleted, err := deleteRejectedStagesWithLinkedTags(context.Background(), sm, customTagsByStageID, false, nil)
	require.NoError(t, err)

	assert.Equal(t, []string{stageID.String()}, deleted)
	assert.Equal(t, []image.StageID{*stageID}, sm.stages.deletedImages, "stage image deleted first")
	assert.Equal(t, []string{"v1.0.0", "latest"}, sm.stages.deletedTags, "linked custom tags deleted next, in given order")
	assert.Equal(t, []image.StageID{*stageID}, sm.stages.deletedRecords, "marker deleted last")
}

func TestDeleteRejectedStagesWithLinkedTags_DryRun(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	stageID := image.NewStageID(digest, 1700000000)

	sm := newFakeStorageManager()
	sm.stages.rejectedStageIDs = []image.StageID{*stageID}

	deleted, err := deleteRejectedStagesWithLinkedTags(context.Background(), sm, map[string][]string{stageID.String(): {"v1.0.0"}}, true, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{stageID.String()}, deleted)
	assert.Empty(t, sm.stages.deletedImages, "dry run must not touch registry")
	assert.Empty(t, sm.stages.deletedTags, "dry run must not touch registry")
	assert.Empty(t, sm.stages.deletedRecords, "dry run must not touch registry")
}

func TestDeleteRejectedStagesWithLinkedTags_PropagatesGetError(t *testing.T) {
	sm := newFakeStorageManager()
	sm.stages.rejectedErr = errors.New("registry down")

	_, err := deleteRejectedStagesWithLinkedTags(context.Background(), sm, nil, false, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to get rejected stage ids")
}

func TestDeleteRejectedStagesWithLinkedTags_StageImageNonFatalFailureKeepsMarker(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	stageID := image.NewStageID(digest, 1700000000)

	sm := newFakeStorageManager()
	sm.stages.rejectedStageIDs = []image.StageID{*stageID}
	sm.stages.deleteImageErrs[stageID.String()] = errors.New("temporary network glitch")

	deleted, err := deleteRejectedStagesWithLinkedTags(context.Background(), sm, map[string][]string{stageID.String(): {"v1.0.0"}}, false, nil)
	require.NoError(t, err)

	assert.Empty(t, deleted, "stage image deletion failed: stage not reported deleted, retry on next cleanup")
	assert.Equal(t, []image.StageID{*stageID}, sm.stages.deletedImages, "attempt was made")
	assert.Empty(t, sm.stages.deletedTags, "custom tags must NOT be touched when stage image delete failed")
	assert.Empty(t, sm.stages.deletedRecords, "marker must remain so retry picks up this stage")
}

func TestDeleteRejectedStagesWithLinkedTags_StageImageFatalFailurePropagates(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	stageID := image.NewStageID(digest, 1700000000)

	sm := newFakeStorageManager()
	sm.stages.rejectedStageIDs = []image.StageID{*stageID}
	sm.stages.deleteImageErrs[stageID.String()] = errors.New("UNAUTHORIZED")

	_, err := deleteRejectedStagesWithLinkedTags(context.Background(), sm, nil, false, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "UNAUTHORIZED")
}

func TestDeleteRejectedStagesWithLinkedTags_CustomTagFailureKeepsMarker(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	stageID := image.NewStageID(digest, 1700000000)

	sm := newFakeStorageManager()
	sm.stages.rejectedStageIDs = []image.StageID{*stageID}
	sm.stages.deleteTagErrs["v1.0.0"] = errors.New("temporary network glitch")

	deleted, err := deleteRejectedStagesWithLinkedTags(context.Background(), sm, map[string][]string{stageID.String(): {"v1.0.0", "latest"}}, false, nil)
	require.NoError(t, err)

	assert.Empty(t, deleted, "stage with failed custom tag must NOT be reported deleted")
	assert.Equal(t, []image.StageID{*stageID}, sm.stages.deletedImages, "stage image already deleted")
	assert.Equal(t, []string{"v1.0.0"}, sm.stages.deletedTags, "fail-fast on first custom tag failure; 'latest' not attempted")
	assert.Empty(t, sm.stages.deletedRecords, "marker MUST remain so next cleanup retries linked tags")
}

func TestDeleteRejectedStagesWithLinkedTags_MarkerFailureExcludesFromDeleted(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	stageID := image.NewStageID(digest, 1700000000)

	sm := newFakeStorageManager()
	sm.stages.rejectedStageIDs = []image.StageID{*stageID}
	sm.stages.deleteRecordErrs[stageID.String()] = errors.New("temporary network glitch")

	deleted, err := deleteRejectedStagesWithLinkedTags(context.Background(), sm, nil, false, nil)
	require.NoError(t, err)

	assert.Empty(t, deleted, "marker deletion failed: stage not in deleted list")
	assert.Equal(t, []image.StageID{*stageID}, sm.stages.deletedImages)
	assert.Equal(t, []image.StageID{*stageID}, sm.stages.deletedRecords, "attempt was made")
}

func newTestReport() *cleanup_report.Report {
	return cleanup_report.NewReport(context.Background(), "cleanup", false, "example.com/repo", cleanup_report.NewReportOptions{})
}

func TestDeleteRejectedStagesWithLinkedTags_ReportRecordsEachSubAction(t *testing.T) {
	stageID := image.NewStageID("deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef", 1700000000)

	sm := newFakeStorageManager()
	sm.stages.rejectedStageIDs = []image.StageID{*stageID}

	report := newTestReport()

	_, err := deleteRejectedStagesWithLinkedTags(context.Background(), sm, map[string][]string{stageID.String(): {"v1.0.0", "latest"}}, false, report)
	require.NoError(t, err)

	assert.ElementsMatch(t, []cleanup_report.Item{
		{Type: cleanup_report.ItemTypeRejectedStage, Tag: stageID.String()},
		{Type: cleanup_report.ItemTypeCustomTag, Tag: "v1.0.0"},
		{Type: cleanup_report.ItemTypeCustomTag, Tag: "latest"},
		{Type: cleanup_report.ItemTypeRejectedStageMarker, Tag: stageID.String() + "-rejected"},
	}, report.Deleted)
	assert.Empty(t, report.Kept)
}

func TestDeleteRejectedStagesWithLinkedTags_ReportDryRunMatchesRealRun(t *testing.T) {
	stageID := image.NewStageID("deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef", 1700000000)
	customTagsByStageID := map[string][]string{stageID.String(): {"v1.0.0", "latest"}}

	realSM := newFakeStorageManager()
	realSM.stages.rejectedStageIDs = []image.StageID{*stageID}
	realReport := newTestReport()
	_, err := deleteRejectedStagesWithLinkedTags(context.Background(), realSM, customTagsByStageID, false, realReport)
	require.NoError(t, err)

	drySM := newFakeStorageManager()
	drySM.stages.rejectedStageIDs = []image.StageID{*stageID}
	dryReport := newTestReport()
	_, err = deleteRejectedStagesWithLinkedTags(context.Background(), drySM, customTagsByStageID, true, dryReport)
	require.NoError(t, err)

	assert.ElementsMatch(t, realReport.Deleted, dryReport.Deleted, "a dry run must report exactly what a real run would delete")
	assert.Empty(t, drySM.stages.deletedImages, "dry run must not touch the registry")
}

func TestDeleteRejectedStagesWithLinkedTags_ReportSkipsFailedCustomTagAndCanceledMarker(t *testing.T) {
	stageID := image.NewStageID("deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef", 1700000000)

	sm := newFakeStorageManager()
	sm.stages.rejectedStageIDs = []image.StageID{*stageID}
	sm.stages.deleteTagErrs["v1.0.0"] = errors.New("temporary network glitch")

	report := newTestReport()

	_, err := deleteRejectedStagesWithLinkedTags(context.Background(), sm, map[string][]string{stageID.String(): {"v1.0.0", "latest"}}, false, report)
	require.NoError(t, err)

	assert.Equal(t, []cleanup_report.Item{
		{Type: cleanup_report.ItemTypeRejectedStage, Tag: stageID.String()},
	}, report.Deleted, "the failed custom tag, the tags after it and the canceled marker must not be reported")
}

func TestDeleteRejectedStagesWithLinkedTags_ReportSkipsFailedMarker(t *testing.T) {
	stageID := image.NewStageID("deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef", 1700000000)

	sm := newFakeStorageManager()
	sm.stages.rejectedStageIDs = []image.StageID{*stageID}
	sm.stages.deleteRecordErrs[stageID.String()] = errors.New("temporary network glitch")

	report := newTestReport()

	_, err := deleteRejectedStagesWithLinkedTags(context.Background(), sm, nil, false, report)
	require.NoError(t, err)

	assert.Equal(t, []cleanup_report.Item{
		{Type: cleanup_report.ItemTypeRejectedStage, Tag: stageID.String()},
	}, report.Deleted, "a marker whose deletion failed must not be reported deleted")
}

func TestDeleteRejectedStagesWithLinkedTags_NilReportChangesNothing(t *testing.T) {
	stageID := image.NewStageID("deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef", 1700000000)

	withReport := newFakeStorageManager()
	withReport.stages.rejectedStageIDs = []image.StageID{*stageID}
	deletedWith, err := deleteRejectedStagesWithLinkedTags(context.Background(), withReport, map[string][]string{stageID.String(): {"v1.0.0"}}, false, newTestReport())
	require.NoError(t, err)

	withoutReport := newFakeStorageManager()
	withoutReport.stages.rejectedStageIDs = []image.StageID{*stageID}
	deletedWithout, err := deleteRejectedStagesWithLinkedTags(context.Background(), withoutReport, map[string][]string{stageID.String(): {"v1.0.0"}}, false, nil)
	require.NoError(t, err)

	assert.Equal(t, deletedWith, deletedWithout)
	assert.Equal(t, withReport.stages.deletedImages, withoutReport.stages.deletedImages)
	assert.Equal(t, withReport.stages.deletedTags, withoutReport.stages.deletedTags)
	assert.Equal(t, withReport.stages.deletedRecords, withoutReport.stages.deletedRecords)
}

func TestDeleteCustomTags_ReportRecordsOnlySucceeded(t *testing.T) {
	sm := newFakeStorageManager()
	sm.stages.deleteTagErrs["broken"] = errors.New("temporary network glitch")

	report := newTestReport()

	require.NoError(t, deleteCustomTags(context.Background(), sm, []string{"kept-alive", "broken"}, false, report))

	assert.Equal(t, []cleanup_report.Item{
		{Type: cleanup_report.ItemTypeCustomTag, Tag: "kept-alive"},
	}, report.Deleted, "a tag logged as deleted after a failed deletion must not be reported")
}

func TestDeleteCustomTags_ReportDryRunMatchesRealRun(t *testing.T) {
	tags := []string{"one", "two"}

	realReport := newTestReport()
	require.NoError(t, deleteCustomTags(context.Background(), newFakeStorageManager(), tags, false, realReport))

	dryReport := newTestReport()
	require.NoError(t, deleteCustomTags(context.Background(), newFakeStorageManager(), tags, true, dryReport))

	assert.ElementsMatch(t, realReport.Deleted, dryReport.Deleted)
}

func TestDeleteImportsMetadata_ReportRecordsOnlySucceeded(t *testing.T) {
	sm := newFakeStorageManager()
	sm.importMetadataErrs["broken"] = errors.New("temporary network glitch")

	report := newTestReport()

	require.NoError(t, deleteImportsMetadata(context.Background(), "myproject", sm, []string{"8c4a1f9b2d7e5a3c", "broken"}, false, report))

	assert.Equal(t, []cleanup_report.Item{
		{Type: cleanup_report.ItemTypeImportMetadata, ID: "8c4a1f9b2d7e5a3c"},
	}, report.Deleted)
}

func TestDeleteImportsMetadata_ReportDryRunMatchesRealRun(t *testing.T) {
	ids := []string{"8c4a1f9b2d7e5a3c", "1e09fb543b4ef442"}

	realReport := newTestReport()
	require.NoError(t, deleteImportsMetadata(context.Background(), "myproject", newFakeStorageManager(), ids, false, realReport))

	dryReport := newTestReport()
	require.NoError(t, deleteImportsMetadata(context.Background(), "myproject", newFakeStorageManager(), ids, true, dryReport))

	assert.ElementsMatch(t, realReport.Deleted, dryReport.Deleted)
}

func TestDeleteImageMetadata_ReportUsesImageNameForManagedImages(t *testing.T) {
	report := newTestReport()
	stageIDCommitList := map[string][]string{"ff0011-1748001122334": {"a3f1c92e"}}

	require.NoError(t, deleteImageMetadata(context.Background(), "myproject", newFakeStorageManager(), "backend", stageIDCommitList, true, false, report))

	assert.Equal(t, []cleanup_report.Item{
		{Type: cleanup_report.ItemTypeImageMetadata, ImageName: "backend", StageID: "ff0011-1748001122334", Commit: "a3f1c92e"},
	}, report.Deleted)
}

func TestDeleteImageMetadata_ReportUsesIDForUnresolvableMetadata(t *testing.T) {
	report := newTestReport()
	stageIDCommitList := map[string][]string{"ff0011-1748001122334": {"a3f1c92e"}}

	require.NoError(t, deleteImageMetadata(context.Background(), "myproject", newFakeStorageManager(), "8c4a1f9b2d7e5a3c", stageIDCommitList, true, true, report))

	assert.Equal(t, []cleanup_report.Item{
		{Type: cleanup_report.ItemTypeImageMetadata, ID: "8c4a1f9b2d7e5a3c", StageID: "ff0011-1748001122334", Commit: "a3f1c92e"},
	}, report.Deleted)
}

func (f *fakeStorageManager) GetStageDescSetWithCache(_ context.Context) (image.StageDescSet, error) {
	return f.stageDescSet, nil
}

func (f *fakeStorageManager) GetFinalStageDescSet(_ context.Context) (image.StageDescSet, error) {
	return f.finalStageDescSet, nil
}

func TestCleanupFinalStages_ReportRecordsKeptFinalStageWithReason(t *testing.T) {
	ctx := context.Background()

	finalStageDesc := &image.StageDesc{
		StageID: image.NewStageID("ff0011", 1748001122334),
		Info:    &image.Info{Tag: "ff0011-1748001122334"},
	}

	sm := newFakeStorageManager()
	sm.finalStageDescSet = image.NewStageDescSet(finalStageDesc)
	sm.stageDescSet = image.NewStageDescSet()

	stageManager := stage_manager.NewManager()
	require.NoError(t, stageManager.InitStageDescSet(ctx, sm))
	require.NoError(t, stageManager.InitFinalStageDescSet(ctx, sm))

	report := newTestReport()
	m := &cleanupManager{stageManager: stageManager, StorageManager: sm, report: report}

	require.NoError(t, m.cleanupFinalStages(ctx))

	assert.Equal(t, []cleanup_report.Item{{
		Type:   cleanup_report.ItemTypeFinalStage,
		Tag:    "ff0011-1748001122334",
		Reason: stage_manager.ProtectionReasonNotFoundInRepo.String(),
	}}, report.Kept)
	assert.Empty(t, report.Deleted)
}
