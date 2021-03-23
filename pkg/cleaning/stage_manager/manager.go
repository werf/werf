package stage_manager

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/storage/manager"
)

type Manager struct {
	stages            map[string]*stage
	imageMetadataList []*imageMetadata
}

func NewManager() Manager {
	return Manager{
		stages: map[string]*stage{},
	}
}

type stage struct {
	stageID     string
	description *image.StageDescription
	isProtected bool
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

func (m *Manager) getOrCreateImageMetadata(imageName string, stageID string) *imageMetadata {
	for _, im := range m.imageMetadataList {
		if im.imageName == imageName && im.stageID == stageID {
			return im
		}
	}

	im := m.newImageMetadata(imageName, stageID)
	m.imageMetadataList = append(m.imageMetadataList, im)

	return im
}

func (m *Manager) newImageMetadata(imageName string, stageID string) *imageMetadata {
	return &imageMetadata{imageName: imageName, stageID: stageID, isNonexistentStage: !m.isStageExist(stageID)}
}

func (m *Manager) InitStages(ctx context.Context, storageManager *manager.StorageManager) error {
	stageDescriptionList, err := storageManager.GetStageDescriptionList(ctx)
	if err != nil {
		return err
	}

	for _, description := range stageDescriptionList {
		stageID := description.Info.Tag
		m.stages[stageID] = newStage(stageID, description)
	}

	return nil
}

type GitRepo interface {
	IsCommitExists(ctx context.Context, commit string) (bool, error)
}

func (m *Manager) InitImagesMetadata(ctx context.Context, storageManager *manager.StorageManager, localGit GitRepo, projectName string, imageNameList []string) error {
	imageMetadataByImageName, imageMetadataByNotManagedImageName, err := storageManager.StagesStorage.GetAllAndGroupImageMetadataByImageName(ctx, projectName, imageNameList)
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
					return fmt.Errorf("check commit %s in local git failed: %s", commit, err)
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

func (m *Manager) GetStageIDList() []string {
	var result []string
	for stageID := range m.stages {
		result = append(result, stageID)
	}

	return result
}

func (m *Manager) MarkStageAsProtected(stageID string) {
	m.stages[stageID].isProtected = true
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

		commitList, ok := stageIDCommitList[im.stageID]
		if !ok {
			commitList = []string{}
		}

		stageIDCommitList[im.stageID] = append(commitList, im.commitList...)
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

		commitList, ok := result[im.stageID]
		if !ok {
			commitList = []string{}
		}

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

		commitList, ok := stageIDCommitList[im.stageID]
		if !ok {
			commitList = []string{}
		}

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

		commitList, ok := result[im.stageID]
		if !ok {
			commitList = []string{}
		}

		result[im.stageID] = append(commitList, im.commitListToDelete...)
	}

	return result
}

func (m *Manager) GetStageDescriptionList() []*image.StageDescription {
	var result []*image.StageDescription
	for _, stage := range m.stages {
		result = append(result, stage.description)
	}

	return result
}

func (m *Manager) GetProtectedStageDescriptionList() []*image.StageDescription {
	var result []*image.StageDescription
	for _, stage := range m.stages {
		if stage.isProtected {
			result = append(result, stage.description)
		}
	}

	return result
}

func (m *Manager) isStageExist(stageID string) bool {
	_, exist := m.stages[stageID]
	return exist
}
