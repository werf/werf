package stage_manager

import (
	"context"
	"fmt"
	"sync"

	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/storage/manager"
)

type Manager struct {
	stages               map[string]*stage
	stageIDCustomTagList map[string][]string
	finalStages          map[string]*stage
	imageMetadataList    []*imageMetadata
}

func NewManager() Manager {
	return Manager{
		stages:               map[string]*stage{},
		stageIDCustomTagList: map[string][]string{},
		finalStages:          map[string]*stage{},
	}
}

type stage struct {
	stageID          string
	isMultiplatform  bool //nolint:unused
	description      *image.StageDescription
	isProtected      bool
	protectionReason string
}

func newStage(stageID string, description *image.StageDescription) *stage {
	return &stage{stageID: stageID, description: description}
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
	return &imageMetadata{imageName: imageName, stageID: stageID, isNonexistentStage: !m.IsStageExist(stageID)}
}

func (m *Manager) InitStages(ctx context.Context, storageManager manager.StorageManagerInterface) error {
	stageDescriptionList, err := storageManager.GetStageDescriptionListWithCache(ctx)
	if err != nil {
		return err
	}

	for _, description := range stageDescriptionList {
		stageID := description.Info.Tag
		m.stages[stageID] = newStage(stageID, description)
	}

	return nil
}

func (m *Manager) InitFinalStages(ctx context.Context, storageManager manager.StorageManagerInterface) error {
	finalStageDescriptionList, err := storageManager.GetFinalStageDescriptionList(ctx)
	if err != nil {
		return err
	}

	for _, description := range finalStageDescriptionList {
		stageID := description.Info.Tag
		m.finalStages[stageID] = newStage(stageID, description)
	}

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

func (m *Manager) GetStageIDList() []string {
	var result []string
	for stageID := range m.stages {
		result = append(result, stageID)
	}

	return result
}

func (m *Manager) GetFinalStageIDList() []string {
	var result []string
	for stageID := range m.finalStages {
		result = append(result, stageID)
	}

	return result
}

func (m *Manager) MarkStageAsProtected(stageID, reason string) {
	m.stages[stageID].isProtected = true
	m.stages[stageID].protectionReason = reason
}

func (m *Manager) MarkFinalStageAsProtected(stageID, reason string) {
	m.finalStages[stageID].isProtected = true
	m.finalStages[stageID].protectionReason = reason
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

func (m *Manager) ForgetDeletedStages(stages []*image.StageDescription) {
	for _, stg := range stages {
		if _, hasKey := m.stages[stg.StageID.String()]; hasKey {
			delete(m.stages, stg.StageID.String())
		}
	}
}

func (m *Manager) ForgetDeletedFinalStages(stages []*image.StageDescription) {
	for _, stg := range stages {
		if _, hasKey := m.finalStages[stg.StageID.String()]; hasKey {
			delete(m.finalStages, stg.StageID.String())
		}
	}
}

type StageDescriptionListOptions struct {
	ExcludeProtected bool
	OnlyProtected    bool
}

func getStageDescriptionList(stages map[string]*stage, opts StageDescriptionListOptions) (result []*image.StageDescription) {
	for _, stage := range stages {
		if stage.isProtected && opts.ExcludeProtected {
			continue
		}
		if !stage.isProtected && opts.OnlyProtected {
			continue
		}
		result = append(result, stage.description)
	}

	return
}

func (m *Manager) GetStageDescriptionList(opts StageDescriptionListOptions) []*image.StageDescription {
	return getStageDescriptionList(m.stages, opts)
}

func (m *Manager) GetFinalStageDescriptionList(opts StageDescriptionListOptions) []*image.StageDescription {
	return getStageDescriptionList(m.finalStages, opts)
}

func (m *Manager) GetProtectedStageDescriptionListByReason() map[string][]*image.StageDescription {
	res := make(map[string][]*image.StageDescription)

	for _, stage := range m.stages {
		if !stage.isProtected {
			continue
		}
		res[stage.protectionReason] = append(res[stage.protectionReason], stage.description)
	}

	return res
}

func (m *Manager) IsStageExist(stageID string) bool {
	_, exist := m.stages[stageID]
	return exist
}

func (m *Manager) GetCustomTagsMetadata() map[string][]string {
	return m.stageIDCustomTagList
}

func (m *Manager) ForgetCustomTagsByStageID(stageID string) {
	delete(m.stageIDCustomTagList, stageID)
}
