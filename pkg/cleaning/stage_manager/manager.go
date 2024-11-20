package stage_manager

import (
	"context"
	"fmt"
	"sync"

	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/storage/manager"
)

type Manager struct {
	stageDescSet         managedStageDescSet
	finalStageDescSet    managedStageDescSet
	stageIDCustomTagList map[string][]string
	imageMetadataList    []*imageMetadata
}

func NewManager() Manager {
	return Manager{
		stageIDCustomTagList: map[string][]string{},
	}
}

type imageMetadata struct {
	stageID            string
	imageName          string
	commitList         []string
	commitListToDelete []string // either a commit that does not exist or a commit that should be deleted without any checks

	isNonexistentImage bool
	isNonexistentStage bool
}

func (m *Manager) getOrCreateImageMetadata(imageName, stageID string) *imageMetadata {
	for _, im := range m.imageMetadataList {
		if im.imageName == imageName && im.stageID == stageID {
			return im
		}
	}

	im := m.newImageMetadata(imageName, stageID)
	m.imageMetadataList = append(m.imageMetadataList, im)

	return im
}

func (m *Manager) newImageMetadata(imageName, stageID string) *imageMetadata {
	return &imageMetadata{imageName: imageName, stageID: stageID, isNonexistentStage: !m.ContainsStageDescByStageID(stageID)}
}

func (m *Manager) InitStageDescSet(ctx context.Context, storageManager manager.StorageManagerInterface) error {
	stageDescSet, err := storageManager.GetStageDescSetWithCache(ctx)
	if err != nil {
		return err
	}

	m.stageDescSet = newManagedStageDescSet(stageDescSet)

	return nil
}

func (m *Manager) InitFinalStageDescSet(ctx context.Context, storageManager manager.StorageManagerInterface) error {
	finalStageDescSet, err := storageManager.GetFinalStageDescSet(ctx)
	if err != nil {
		return err
	}

	m.finalStageDescSet = newManagedStageDescSet(finalStageDescSet)

	return nil
}

type GitRepo interface {
	IsCommitExists(ctx context.Context, commit string) (bool, error)
}

func (m *Manager) InitImagesMetadata(ctx context.Context, storageManager manager.StorageManagerInterface, localGit GitRepo, projectName string, imageNameList []string) error {
	imageMetadataByImageName, imageMetadataByNotManagedImageName, err := storageManager.GetStagesStorage().GetAllAndGroupImageMetadataByImageName(ctx, projectName, imageNameList, storage.WithCache())
	if err != nil {
		return err
	}

	for imageName, stageIDCommitList := range imageMetadataByNotManagedImageName {
		for stageID, commitList := range stageIDCommitList {
			im := m.getOrCreateImageMetadata(imageName, stageID)
			im.isNonexistentImage = true
			im.commitListToDelete = commitList
		}
	}

	for imageName, stageIDCommitList := range imageMetadataByImageName {
		for stageID, commitList := range stageIDCommitList {
			im := m.getOrCreateImageMetadata(imageName, stageID)
			for _, commit := range commitList {
				exist, err := localGit.IsCommitExists(ctx, commit)
				if err != nil {
					return fmt.Errorf("check commit %s in local git failed: %w", commit, err)
				}

				if exist {
					im.commitList = append(im.commitList, commit)
				} else {
					im.commitListToDelete = append(im.commitListToDelete, commit)
				}
			}
		}
	}

	return nil
}

func (m *Manager) InitCustomTagsMetadata(ctx context.Context, storageManager manager.StorageManagerInterface) error {
	stageIDCustomTagList, err := GetCustomTagsMetadata(ctx, storageManager)
	if err != nil {
		return err
	}

	m.stageIDCustomTagList = stageIDCustomTagList
	return nil
}

func GetCustomTagsMetadata(ctx context.Context, storageManager manager.StorageManagerInterface) (map[string][]string, error) {
	stageCustomTagMetadataIDs, err := storageManager.GetStagesStorage().GetStageCustomTagMetadataIDs(ctx, storage.WithCache())
	if err != nil {
		return nil, fmt.Errorf("unable to get stage custom tag metadata IDs: %w", err)
	}

	var mutex sync.Mutex
	stageIDCustomTagList := make(map[string][]string)
	err = storageManager.ForEachGetStageCustomTagMetadata(ctx, stageCustomTagMetadataIDs, func(ctx context.Context, metadataID string, metadata *storage.CustomTagMetadata, err error) error {
		if err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		stageIDCustomTagList[metadata.StageID] = append(stageIDCustomTagList[metadata.StageID], metadata.Tag)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return stageIDCustomTagList, nil
}

func (m *Manager) MarkStageDescAsProtectedByStageID(stageID, reason string) {
	stageDesc := m.GetStageDescByStageID(stageID)
	if stageDesc == nil {
		panic(fmt.Sprintf("stage description %s not found", stageID))
	}

	m.stageDescSet.MarkStageDescAsProtected(m.GetStageDescByStageID(stageID), reason)
}

func (m *Manager) MarkStageDescAsProtected(stageDesc *image.StageDesc, reason string) {
	m.stageDescSet.MarkStageDescAsProtected(stageDesc, reason)
}

func (m *Manager) MarkFinalStageDescAsProtected(stageDesc *image.StageDesc, reason string) {
	m.finalStageDescSet.MarkStageDescAsProtected(stageDesc, reason)
}

// GetImageStageIDCommitListToCleanup method returns existing stage IDs and related existing commits (for each managed image)
func (m *Manager) GetImageStageIDCommitListToCleanup() map[string]map[string][]string {
	result := map[string]map[string][]string{}
	for _, im := range m.imageMetadataList {
		if im.isNonexistentImage || im.isNonexistentStage {
			continue
		}

		stageIDCommitList, ok := result[im.imageName]
		if !ok {
			stageIDCommitList = map[string][]string{}
		}

		stageIDCommitList[im.stageID] = append(stageIDCommitList[im.stageID], im.commitList...)
		result[im.imageName] = stageIDCommitList
	}

	return result
}

// GetStageIDCommitListToCleanup method is shortcut for GetImageStageIDCommitListToCleanup
func (m *Manager) GetStageIDCommitListToCleanup(imageName string) map[string][]string {
	result, ok := m.GetImageStageIDCommitListToCleanup()[imageName]
	if !ok {
		return map[string][]string{}
	}

	return result
}

// GetNonexistentStageIDCommitList method returns nonexistent stage IDs and all related commits for certain image
func (m *Manager) GetNonexistentStageIDCommitList(imageName string) map[string][]string {
	result := map[string][]string{}
	for _, im := range m.imageMetadataList {
		if !im.isNonexistentStage {
			continue
		}

		if im.imageName != imageName {
			continue
		}

		commitList := result[im.stageID]
		commitList = append(commitList, im.commitList...)
		commitList = append(commitList, im.commitListToDelete...)
		result[im.stageID] = commitList
	}

	return result
}

// GetStageIDCommitListByNonexistentImage method returns all stage IDs and related commits for each nonexistent image
func (m *Manager) GetStageIDCommitListByNonexistentImage() map[string]map[string][]string {
	result := map[string]map[string][]string{}
	for _, im := range m.imageMetadataList {
		if !im.isNonexistentImage {
			continue
		}

		stageIDCommitList, ok := result[im.imageName]
		if !ok {
			stageIDCommitList = map[string][]string{}
		}

		commitList := stageIDCommitList[im.stageID]
		commitList = append(commitList, im.commitList...)
		commitList = append(commitList, im.commitListToDelete...)
		stageIDCommitList[im.stageID] = commitList

		result[im.imageName] = stageIDCommitList
	}

	return result
}

// GetStageIDNonexistentCommitList method returns stage IDs and related nonexistent commits for certain image
func (m *Manager) GetStageIDNonexistentCommitList(imageName string) map[string][]string {
	result := map[string][]string{}
	for _, im := range m.imageMetadataList {
		if im.isNonexistentStage {
			continue
		}

		if im.imageName != imageName {
			continue
		}

		result[im.stageID] = append(result[im.stageID], im.commitListToDelete...)
	}

	return result
}

func (m *Manager) ForgetDeletedStageDescSet(stageDescSet image.StageDescSet) {
	m.stageDescSet.DifferenceInPlace(stageDescSet)
}

func (m *Manager) ForgetDeletedFinalStageDescSet(stageDescSet image.StageDescSet) {
	m.finalStageDescSet.DifferenceInPlace(stageDescSet)
}

func (m *Manager) GetStageDescSet() image.StageDescSet {
	return m.stageDescSet.StageDescSet()
}

func (m *Manager) GetFinalStageDescSet() image.StageDescSet {
	return m.finalStageDescSet.StageDescSet()
}

func (m *Manager) GetProtectedStageDescSet() image.StageDescSet {
	return m.stageDescSet.GetProtectedStageDescSet()
}

func (m *Manager) GetFinalProtectedStageDescSet() image.StageDescSet {
	return m.finalStageDescSet.GetProtectedStageDescSet()
}

func (m *Manager) GetProtectedStageDescSetByReason() map[string]image.StageDescSet {
	return m.stageDescSet.GetProtectedStageDescSetByReason()
}

func (m *Manager) GetStageDescByStageID(stageID string) *image.StageDesc {
	return m.stageDescSet.GetStageDescByStageID(stageID)
}

func (m *Manager) ContainsStageDescByStageID(stageID string) bool {
	return m.stageDescSet.GetStageDescByStageID(stageID) != nil
}

func (m *Manager) GetCustomTagsMetadata() map[string][]string {
	return m.stageIDCustomTagList
}

func (m *Manager) ForgetCustomTagsByStageID(stageID string) {
	delete(m.stageIDCustomTagList, stageID)
}
