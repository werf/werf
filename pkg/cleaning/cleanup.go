package cleaning

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/rodaine/table"

	"github.com/go-git/go-git/v5"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/cleaning/allow_list"
	"github.com/werf/werf/pkg/cleaning/git_history_based_cleanup"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/logging"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/storage/manager"
	"github.com/werf/werf/pkg/util"
)

const stagesCleanupDefaultIgnorePeriodPolicy = 2 * 60 * 60

type CleanupOptions struct {
	ImageNameList                           []string
	LocalGit                                GitRepo
	KubernetesContextClients                []*kube.ContextClient
	KubernetesNamespaceRestrictionByContext map[string]string
	WithoutKube                             bool
	GitHistoryBasedCleanupOptions           config.MetaCleanup
	DryRun                                  bool
}

func Cleanup(ctx context.Context, projectName string, storageManager *manager.StorageManager, storageLockManager storage.LockManager, options CleanupOptions) error {
	m := newCleanupManager(projectName, storageManager, options)

	if lock, err := storageLockManager.LockStagesAndImages(ctx, projectName, storage.LockStagesAndImagesOptions{GetOrCreateImagesOnly: false}); err != nil {
		return fmt.Errorf("unable to lock stages and images: %s", err)
	} else {
		defer storageLockManager.Unlock(ctx, lock)
	}

	return m.run(ctx)
}

func newCleanupManager(projectName string, storageManager *manager.StorageManager, options CleanupOptions) *cleanupManager {
	return &cleanupManager{
		ProjectName:                             projectName,
		StorageManager:                          storageManager,
		ImageNameList:                           options.ImageNameList,
		DryRun:                                  options.DryRun,
		LocalGit:                                options.LocalGit,
		KubernetesContextClients:                options.KubernetesContextClients,
		KubernetesNamespaceRestrictionByContext: options.KubernetesNamespaceRestrictionByContext,
		WithoutKube:                             options.WithoutKube,
		GitHistoryBasedCleanupOptions:           options.GitHistoryBasedCleanupOptions,
	}
}

type cleanupManager struct {
	stages                     []*image.StageDescription
	imageNameLinkListByStageID map[string][]string

	imageNameStageIDCommitList            map[string]map[string][]string
	imageNameStageIDCommitListToCleanup   map[string]map[string][]string
	imageNameNonexistentStageIDCommitList map[string]map[string][]string
	imageNameStageIDNonexistentCommitList map[string]map[string][]string
	nonexistentImageNameStageIDCommitList map[string]map[string][]string

	checksumSourceImageIDs       map[string][]string
	nonexistentImportMetadataIDs []string

	ProjectName                             string
	StorageManager                          *manager.StorageManager
	ImageNameList                           []string
	LocalGit                                GitRepo
	KubernetesContextClients                []*kube.ContextClient
	KubernetesNamespaceRestrictionByContext map[string]string
	WithoutKube                             bool
	GitHistoryBasedCleanupOptions           config.MetaCleanup
	DryRun                                  bool
}

type GitRepo interface {
	PlainOpen() (*git.Repository, error)
	IsCommitExists(ctx context.Context, commit string) (bool, error)
}

func (m *cleanupManager) init(ctx context.Context) error {
	if err := logboek.Context(ctx).Info().LogProcess("Fetching manifests").DoError(func() error {
		return m.initStages(ctx)
	}); err != nil {
		return err
	}

	if err := logboek.Context(ctx).Info().LogProcess("Fetching metadata").DoError(func() error {
		return m.initImagesMetadata(ctx)
	}); err != nil {
		return err
	}

	return nil
}

func (m *cleanupManager) initStages(ctx context.Context) error {
	stages, err := m.StorageManager.GetStageDescriptionList(ctx)
	if err != nil {
		return err
	}

	m.stages = stages

	return nil
}

func (m *cleanupManager) isStageExist(stageID string) bool {
	stage := m.getStage(stageID)
	return stage != nil
}

func (m *cleanupManager) mustGetStage(stageID string) *image.StageDescription {
	stage := m.getStage(stageID)
	if stage == nil {
		panic(fmt.Sprintf("runtime error: stage was not found in memory by stageID %s", stageID))
	}

	return stage
}

func (m *cleanupManager) getStage(stageID string) *image.StageDescription {
	for _, stage := range m.stages {
		if stageID == stage.Info.Tag {
			return stage
		}
	}

	return nil
}

func (m *cleanupManager) shouldStageBeDeleted(stageID string) bool {
	return len(m.imageNameLinkListByStageID[stageID]) == 0
}

func (m *cleanupManager) deleteStagesFromCache(stagesToDelete []*image.StageDescription) {
	var result []*image.StageDescription

outerLoop:
	for _, stage := range m.stages {
		for _, stageToDelete := range stagesToDelete {
			if stage == stageToDelete {
				continue outerLoop
			}
		}

		result = append(result, stage)
	}

	m.stages = result
}

func (m *cleanupManager) initImagesMetadata(ctx context.Context) error {
	m.imageNameLinkListByStageID = map[string][]string{}

	m.imageNameStageIDCommitList = map[string]map[string][]string{}
	m.imageNameStageIDCommitListToCleanup = map[string]map[string][]string{}
	m.imageNameNonexistentStageIDCommitList = map[string]map[string][]string{}
	m.imageNameStageIDNonexistentCommitList = map[string]map[string][]string{}
	m.nonexistentImageNameStageIDCommitList = map[string]map[string][]string{}

	imageMetadataByImageName, imageMetadataByNotManagedImageName, err := m.StorageManager.StagesStorage.GetAllAndGroupImageMetadataByImageName(ctx, m.ProjectName, m.ImageNameList)
	if err != nil {
		return err
	}

	m.nonexistentImageNameStageIDCommitList = imageMetadataByNotManagedImageName

	for imageName, stageIDCommitList := range imageMetadataByImageName {
		m.imageNameStageIDCommitList[imageName] = map[string][]string{}
		m.imageNameStageIDCommitListToCleanup[imageName] = map[string][]string{}
		m.imageNameNonexistentStageIDCommitList[imageName] = map[string][]string{}
		m.imageNameStageIDNonexistentCommitList[imageName] = map[string][]string{}

		for stageID, stageIDCommitList := range stageIDCommitList {
			if !m.isStageExist(stageID) {
				m.imageNameNonexistentStageIDCommitList[imageName][stageID] = stageIDCommitList
				continue
			}

			stageIDImageList, ok := m.imageNameLinkListByStageID[stageID]
			if !ok {
				stageIDImageList = []string{}
			}

			stageIDImageList = append(stageIDImageList, imageName)
			m.imageNameLinkListByStageID[stageID] = stageIDImageList

			var commitList, nonexistentCommitList []string
			for _, commit := range stageIDCommitList {
				exist, err := m.LocalGit.IsCommitExists(ctx, commit)
				if err != nil {
					return fmt.Errorf("check commit %s in local git failed: %s", commit, err)
				}

				if exist {
					commitList = append(commitList, commit)
				} else {
					nonexistentCommitList = append(nonexistentCommitList, commit)
				}
			}

			if len(commitList) != 0 {
				m.imageNameStageIDCommitList[imageName][stageID] = commitList
				m.imageNameStageIDCommitListToCleanup[imageName][stageID] = commitList
			}

			if len(nonexistentCommitList) != 0 {
				m.imageNameStageIDNonexistentCommitList[imageName][stageID] = nonexistentCommitList
			}
		}
	}

	return nil
}

func (m *cleanupManager) unlinkStageIDImageName(stageID, imageNameToUnlink string) {
	m.imageNameLinkListByStageID[stageID] = util.ExcludeFromStringArray(m.imageNameLinkListByStageID[stageID], imageNameToUnlink)
}

func (m *cleanupManager) keepStageID(imageName string, stageID string) {
	delete(m.imageNameStageIDCommitListToCleanup[imageName], stageID)
}

func (m *cleanupManager) deleteImageMetadataFromCache(imageName string, stageIDCommitListToDelete map[string][]string) {
	for stageIDToDelete, commitListToDelete := range stageIDCommitListToDelete {
		var resultCommitList []string

	outerLoop:
		for _, commit := range m.imageNameStageIDCommitListToCleanup[imageName][stageIDToDelete] {
			for _, commitToDelete := range commitListToDelete {
				if commitToDelete == commit {
					continue outerLoop
				}
			}

			resultCommitList = append(resultCommitList, commit)
		}

		if len(resultCommitList) == 0 {
			delete(m.imageNameStageIDCommitListToCleanup[imageName], stageIDToDelete)
		} else {
			m.imageNameStageIDCommitListToCleanup[imageName][stageIDToDelete] = resultCommitList
		}
	}
}

func (m *cleanupManager) deleteStageIDFromCache(imageName string, stageID string) {
	delete(m.imageNameStageIDCommitList[imageName], stageID)
}

func (m *cleanupManager) run(ctx context.Context) error {
	if err := logboek.Context(ctx).LogProcess("Fetching manifests and metadata").DoError(func() error {
		return m.init(ctx)
	}); err != nil {
		return err
	}

	if m.LocalGit != nil {
		if !m.WithoutKube {
			if err := logboek.Context(ctx).LogProcess("Skipping repo images that are being used in Kubernetes").DoError(func() error {
				return m.skipStageIDsThatAreUsedInKubernetes(ctx)
			}); err != nil {
				return err
			}
		}

		if err := logboek.Context(ctx).LogProcess("Git history-based cleanup").DoError(func() error {
			return m.gitHistoryBasedCleanup(ctx)
		}); err != nil {
			return err
		}
	} else {
		logboek.Context(ctx).Warn().LogLn("WARNING: Git history-based cleanup skipped due to local git repository was not detected")
		logboek.Context(ctx).Default().LogOptionalLn()
	}

	if err := logboek.Context(ctx).LogProcess("Cleanup unused stages").DoError(func() error {
		return m.cleanupUnusedStages(ctx)
	}); err != nil {
		return err
	}

	return nil
}

func (m *cleanupManager) skipStageIDsThatAreUsedInKubernetes(ctx context.Context) error {
	deployedDockerImagesNames, err := m.deployedDockerImagesNames(ctx)
	if err != nil {
		return err
	}

	skippedDeployedImages := map[string]bool{}
	for imageName, stageIDCommitList := range m.imageNameStageIDCommitListToCleanup {
	Loop:
		for stageID, _ := range stageIDCommitList {
			dockerImageName := fmt.Sprintf("%s:%s", m.StorageManager.StagesStorage.String(), stageID)
			for _, deployedDockerImageName := range deployedDockerImagesNames {
				if deployedDockerImageName == dockerImageName {
					m.keepStageID(imageName, stageID)

					if !skippedDeployedImages[stageID] {
						logboek.Context(ctx).Default().LogFDetails("  tag: %s\n", stageID)
						logboek.Context(ctx).LogOptionalLn()
						skippedDeployedImages[stageID] = true
					}

					continue Loop
				}
			}
		}
	}

	return nil
}

func (m *cleanupManager) deployedDockerImagesNames(ctx context.Context) ([]string, error) {
	var deployedDockerImagesNames []string
	for _, contextClient := range m.KubernetesContextClients {
		if err := logboek.Context(ctx).LogProcessInline("Getting deployed docker images (context %s)", contextClient.ContextName).
			DoError(func() error {
				kubernetesClientDeployedDockerImagesNames, err := allow_list.DeployedDockerImages(contextClient.Client, m.KubernetesNamespaceRestrictionByContext[contextClient.ContextName])
				if err != nil {
					return fmt.Errorf("cannot get deployed imagesStageList: %s", err)
				}

				deployedDockerImagesNames = append(deployedDockerImagesNames, kubernetesClientDeployedDockerImagesNames...)

				return nil
			}); err != nil {
			return nil, err
		}
	}

	return deployedDockerImagesNames, nil
}

func (m *cleanupManager) gitHistoryBasedCleanup(ctx context.Context) error {
	gitRepository, err := m.LocalGit.PlainOpen()
	if err != nil {
		return fmt.Errorf("git plain open failed: %s", err)
	}

	var referencesToScan []*git_history_based_cleanup.ReferenceToScan
	if err := logboek.Context(ctx).Default().LogProcess("Preparing references to scan").DoError(func() error {
		referencesToScan, err = git_history_based_cleanup.ReferencesToScan(ctx, gitRepository, m.GitHistoryBasedCleanupOptions.KeepPolicies)
		return err
	}); err != nil {
		return err
	}

	for imageName, stageIDCommitList := range m.imageNameStageIDCommitListToCleanup {
		var reachedStageIDs []string
		var hitStageIDCommitList map[string][]string
		if err := logboek.Context(ctx).LogProcess(logging.ImageLogProcessName(imageName, false)).DoError(func() error {
			if logboek.Context(ctx).Streams().Width() > 90 {
				m.printStageIDCommitListTable(ctx, imageName)
			}

			if err := logboek.Context(ctx).LogProcess("Scanning git references history").DoError(func() error {
				if len(stageIDCommitList) != 0 {
					reachedStageIDs, hitStageIDCommitList, err = git_history_based_cleanup.ScanReferencesHistory(ctx, gitRepository, referencesToScan, stageIDCommitList)
				} else {
					logboek.Context(ctx).LogLn("Scanning stopped due to nothing to seek")
				}

				return nil
			}); err != nil {
				return err
			}

			var stageIDToDelete []string
			var stagesToDelete []*image.StageDescription

		outerLoop:
			for stageID, _ := range stageIDCommitList {
				for _, reachedStageID := range reachedStageIDs {
					if stageID == reachedStageID {
						continue outerLoop
					}
				}

				stageIDToDelete = append(stageIDToDelete, stageID)

				m.unlinkStageIDImageName(stageID, imageName)
				if m.shouldStageBeDeleted(stageID) {
					stagesToDelete = append(stagesToDelete, m.mustGetStage(stageID))
				}
			}

			if len(reachedStageIDs) != 0 {
				m.handleSavedStageIDs(ctx, reachedStageIDs)
			}

			if len(stagesToDelete) != 0 {
				if err := m.handleStagesToDelete(ctx, stagesToDelete); err != nil {
					return err
				}
			}

			if err := m.cleanupImageMetadata(ctx, imageName, hitStageIDCommitList, stageIDToDelete); err != nil {
				return err
			}

			return nil
		}); err != nil {
			return err
		}
	}

	if err := m.cleanupNonexistentImageMetadata(ctx); err != nil {
		return err
	}

	return nil
}

func (m *cleanupManager) printStageIDCommitListTable(ctx context.Context, imageName string) {
	stageIDCommitList := m.imageNameStageIDCommitListToCleanup[imageName]

	if len(stageIDCommitList) == 0 {
		return
	}

	var rows [][]interface{}
	for stageID, commitList := range stageIDCommitList {
		for ind, commit := range commitList {
			var columns []interface{}
			if ind == 0 {
				stageIDColumn := stageID

				space := len(stageID) - len(commit) - 1
				if logboek.Context(ctx).Streams().ContentWidth() < space {
					stageIDColumn = fmt.Sprintf("%s..%s", stageID[:space-5], stageID[space-3:])
				}

				columns = append(columns, stageIDColumn)
			} else {
				columns = append(columns, "")
			}

			columns = append(columns, commit)
			rows = append(rows, columns)
		}
	}

	if len(rows) != 0 {
		tbl := table.New("Tag", "Commits")
		tbl.WithWriter(logboek.Context(ctx).ProxyOutStream())
		tbl.WithHeaderFormatter(color.New(color.Underline).SprintfFunc())
		for _, row := range rows {
			tbl.AddRow(row...)
		}
		tbl.Print()

		logboek.Context(ctx).LogOptionalLn()
	}
}

func (m *cleanupManager) handleSavedStageIDs(ctx context.Context, savedStageIDs []string) {
	logboek.Context(ctx).Default().LogBlock("Saved tags").Do(func() {
		for _, stageID := range savedStageIDs {
			logboek.Context(ctx).Default().LogFDetails("  tag: %s\n", stageID)
			logboek.Context(ctx).LogOptionalLn()
		}
	})
}

func (m *cleanupManager) handleStagesToDelete(ctx context.Context, stagesToDelete []*image.StageDescription) error {
	return logboek.Context(ctx).Default().LogProcess("Deleting tags").DoError(func() error {
		return m.deleteStages(ctx, stagesToDelete)
	})
}

func (m *cleanupManager) deleteStages(ctx context.Context, stages []*image.StageDescription) error {
	deleteStageOptions := manager.ForEachDeleteStageOptions{
		DeleteImageOptions: storage.DeleteImageOptions{
			RmiForce: false,
		},
		FilterStagesAndProcessRelatedDataOptions: storage.FilterStagesAndProcessRelatedDataOptions{
			SkipUsedImage:            true,
			RmForce:                  false,
			RmContainersThatUseImage: false,
		},
	}

	m.deleteStagesFromCache(stages)

	return deleteStages(ctx, m.StorageManager, m.DryRun, deleteStageOptions, stages)
}

func deleteStages(ctx context.Context, storageManager *manager.StorageManager, dryRun bool, deleteStageOptions manager.ForEachDeleteStageOptions, stages []*image.StageDescription) error {
	if dryRun {
		for _, stageDesc := range stages {
			logboek.Context(ctx).Default().LogFDetails("  tag: %s\n", stageDesc.Info.Tag)
			logboek.Context(ctx).LogOptionalLn()
		}
		return nil
	}

	return storageManager.ForEachDeleteStage(ctx, deleteStageOptions, stages, func(ctx context.Context, stageDesc *image.StageDescription, err error) error {
		if err != nil {
			if err := handleDeletionError(err); err != nil {
				return err
			}

			logboek.Context(ctx).Warn().LogF("WARNING: Image %s deletion failed: %s\n", stageDesc.Info.Name, err)

			return nil
		}

		logboek.Context(ctx).Default().LogFDetails("  tag: %s\n", stageDesc.Info.Tag)

		return nil
	})
}

func (m *cleanupManager) cleanupImageMetadata(ctx context.Context, imageName string, hitStageIDCommitList map[string][]string, stageIDsToDelete []string) error {
	stageIDCommitList := m.imageNameStageIDCommitListToCleanup[imageName]
	nonexistentStageIDCommitList := m.imageNameNonexistentStageIDCommitList[imageName]
	stageIDNonexistentCommitList := m.imageNameStageIDNonexistentCommitList[imageName]

	stageIDCommitListToDelete := map[string][]string{}
	if len(hitStageIDCommitList) != 0 || len(stageIDsToDelete) != 0 {
	stageIDCommitListLoop:
		for stageID, commitList := range stageIDCommitList {
			var commitListToCleanup []string
			for _, stageIDToDelete := range stageIDsToDelete {
				if stageIDToDelete == stageID {
					commitListToCleanup = commitList
					stageIDCommitListToDelete[stageID] = commitListToCleanup
					m.deleteStageIDFromCache(imageName, stageID)
					continue stageIDCommitListLoop
				}
			}

			hitCommitList, ok := hitStageIDCommitList[stageID]
			if ok {
				commitListToCleanup = util.ExcludeFromStringArray(commitList, hitCommitList...)
			} else {
				commitListToCleanup = commitList
			}

			stageIDCommitListToDelete[stageID] = commitListToCleanup
		}

		if len(stageIDCommitListToDelete) != 0 {
			if err := logboek.Context(ctx).Default().LogProcess("Cleaning up metadata").DoError(func() error {
				return m.deleteImagesMetadata(ctx, imageName, stageIDCommitListToDelete, true)
			}); err != nil {
				return err
			}
		}
	}

	if len(nonexistentStageIDCommitList) != 0 {
		if err := logboek.Context(ctx).Default().LogProcess("Deleting metadata for nonexistent stageIDs").DoError(func() error {
			return m.deleteImagesMetadata(ctx, imageName, nonexistentStageIDCommitList, false)
		}); err != nil {
			return err
		}
	}

	if len(stageIDNonexistentCommitList) != 0 {
		if err := logboek.Context(ctx).Default().LogProcess("Deleting metadata for nonexistent commits").DoError(func() error {
			return m.deleteImagesMetadata(ctx, imageName, stageIDNonexistentCommitList, false)
		}); err != nil {
			return err
		}
	}

	return nil
}

func (m *cleanupManager) cleanupNonexistentImageMetadata(ctx context.Context) error {
	if len(m.nonexistentImageNameStageIDCommitList) == 0 {
		return nil
	}

	return logboek.Context(ctx).Default().LogProcess("Deleting metadata for nonexistent images").DoError(func() error {
		for imageName, stageIDCommitList := range m.nonexistentImageNameStageIDCommitList {
			if err := m.deleteImagesMetadata(ctx, imageName, stageIDCommitList, false); err != nil {
				return err
			}
		}

		return nil
	})
}

func (m *cleanupManager) deleteImagesMetadata(ctx context.Context, imageName string, stageIDCommitList map[string][]string, updateCache bool) error {
	if updateCache {
		m.deleteImageMetadataFromCache(imageName, stageIDCommitList)
	}

	return deleteImagesMetadata(ctx, m.ProjectName, m.StorageManager, imageName, stageIDCommitList, m.DryRun)
}

func deleteImagesMetadata(ctx context.Context, projectName string, storageManager *manager.StorageManager, imageNameOrID string, stageIDCommitList map[string][]string, dryRun bool) error {
	if dryRun {
		for stageID, commitList := range stageIDCommitList {
			logboek.Context(ctx).Default().LogFDetails("  imageName: %s\n", imageNameOrID)
			logboek.Context(ctx).Default().LogFDetails("  stageID: %s\n", stageID)
			logboek.Context(ctx).Default().LogFDetails("  commits: %s\n", len(commitList))
			logboek.Context(ctx).LogOptionalLn()
		}
		return nil
	}

	return storageManager.ForEachRmImageMetadata(ctx, projectName, imageNameOrID, stageIDCommitList, func(ctx context.Context, commit, stageID string, err error) error {
		if err != nil {
			if err := handleDeletionError(err); err != nil {
				return err
			}

			logboek.Context(ctx).Warn().LogF("WARNING: Image metadata %s commit %s stage ID %s deletion failed: %s\n", imageNameOrID, commit, stageID, err)

			return nil
		}

		logboek.Context(ctx).Default().LogFDetails("  imageName: %s\n", imageNameOrID)
		logboek.Context(ctx).Default().LogFDetails("  stageID: %s\n", stageID)
		logboek.Context(ctx).Default().LogFDetails("  commit: %s\n", commit)

		return nil
	})
}

func (m *cleanupManager) cleanupUnusedStages(ctx context.Context) error {
	if err := logboek.Context(ctx).Info().LogProcess("Fetching imports metadata").DoError(func() error {
		return m.initImportsMetadata(ctx)
	}); err != nil {
		return fmt.Errorf("unable to init imports metadata: %s", err)
	}

	stagesToDelete := m.stages
	for _, stageIDCommitList := range m.imageNameStageIDCommitList {
		for stageID, _ := range stageIDCommitList {
			stagesToDelete = m.excludeStageAndRelativesByImageID(stagesToDelete, m.mustGetStage(stageID).Info.ID)
		}
	}

	var stagesToSkip []*image.StageDescription
	if os.Getenv("WERF_DISABLE_STAGES_CLEANUP_DATE_PERIOD_POLICY") != "1" {
		for _, stage := range stagesToDelete {
			if time.Now().Unix()-stage.Info.GetCreatedAt().Unix() < stagesCleanupDefaultIgnorePeriodPolicy {
				stagesToSkip = append(stagesToSkip, stage)
			}
		}

		if len(stagesToSkip) != 0 {
			logboek.Context(ctx).Default().LogBlock("Skipping stages that were built within last two hours").Do(func() {
				for _, stage := range stagesToSkip {
					logboek.Context(ctx).Default().LogFDetails("  tag: %s\n", stage.Info.Tag)
					logboek.Context(ctx).LogOptionalLn()
				}
			})
		}
	}

	stagesToDelete = excludeStages(stagesToDelete, stagesToSkip...)

	if len(stagesToDelete) != 0 {
		if err := logboek.Context(ctx).Default().LogProcess("Deleting stages tags").DoError(func() error {
			return m.deleteStages(ctx, stagesToDelete)
		}); err != nil {
			return err
		}
	}

	if len(m.nonexistentImportMetadataIDs) != 0 {
		if err := logboek.Context(ctx).Default().LogProcess("Cleaning imports metadata").DoError(func() error {
			return m.deleteImportsMetadata(ctx, m.nonexistentImportMetadataIDs)
		}); err != nil {
			return err
		}
	}

	return nil
}

func (m *cleanupManager) initImportsMetadata(ctx context.Context) error {
	m.checksumSourceImageIDs = map[string][]string{}

	importMetadataIDs, err := m.StorageManager.StagesStorage.GetImportMetadataIDs(ctx, m.ProjectName)
	if err != nil {
		return err
	}

	var mutex sync.Mutex
	return m.StorageManager.ForEachGetImportMetadata(ctx, m.ProjectName, importMetadataIDs, func(ctx context.Context, metadataID string, metadata *storage.ImportMetadata, err error) error {
		if err != nil {
			return err
		}

		if metadata == nil {
			if err := logboek.Context(ctx).Warn().LogProcessInline("Deleting invalid import metadata %s", metadataID).
				DoError(func() error {
					return m.deleteImportsMetadata(ctx, []string{metadataID})
				}); err != nil {
				return fmt.Errorf("unable to delete import metadata %s: %s", metadataID, err)
			}

			return nil
		}

		importSourceID := metadata.ImportSourceID
		sourceImageID := metadata.SourceImageID
		checksum := metadata.Checksum

		mutex.Lock()
		defer mutex.Unlock()

		stage := findStageByImageID(m.stages, sourceImageID)
		if stage != nil {
			sourceImageIDs, ok := m.checksumSourceImageIDs[checksum]
			if !ok {
				sourceImageIDs = []string{}
			}

			m.checksumSourceImageIDs[checksum] = append(sourceImageIDs, sourceImageID)
		} else {
			m.nonexistentImportMetadataIDs = append(m.nonexistentImportMetadataIDs, importSourceID)
		}

		return nil
	})
}

func (m *cleanupManager) deleteImportsMetadata(ctx context.Context, importMetadataIDs []string) error {
	return deleteImportsMetadata(ctx, m.ProjectName, m.StorageManager, importMetadataIDs, m.DryRun)
}

func deleteImportsMetadata(ctx context.Context, projectName string, storageManager *manager.StorageManager, importMetadataIDs []string, dryRun bool) error {
	if dryRun {
		for _, importMetadataID := range importMetadataIDs {
			logboek.Context(ctx).Default().LogFDetails("  importMetadataID: %s\n", importMetadataID)
			logboek.Context(ctx).LogOptionalLn()
		}
		return nil
	}

	return storageManager.ForEachRmImportMetadata(ctx, projectName, importMetadataIDs, func(ctx context.Context, importMetadataID string, err error) error {
		if err != nil {
			if err := handleDeletionError(err); err != nil {
				return err
			}

			logboek.Context(ctx).Warn().LogF("WARNING: Import metadata ID %s deletion failed: %s\n", importMetadataID, err)

			return nil
		}

		logboek.Context(ctx).Default().LogFDetails("  importMetadataID: %s\n", importMetadataID)

		return nil
	})
}

func (m *cleanupManager) excludeStageAndRelativesByImageID(stages []*image.StageDescription, imageID string) []*image.StageDescription {
	stage := findStageByImageID(stages, imageID)
	if stage == nil {
		return stages
	}

	return m.excludeStageAndRelativesByStage(stages, stage)
}

func findStageByImageID(stages []*image.StageDescription, imageID string) *image.StageDescription {
	for _, stage := range stages {
		if stage.Info.ID == imageID {
			return stage
		}
	}

	return nil
}

func (m *cleanupManager) excludeStageAndRelativesByStage(stages []*image.StageDescription, stage *image.StageDescription) []*image.StageDescription {
	for label, checksum := range stage.Info.Labels {
		if strings.HasPrefix(label, image.WerfImportChecksumLabelPrefix) {
			sourceImageIDs, ok := m.checksumSourceImageIDs[checksum]
			if ok {
				for _, sourceImageID := range sourceImageIDs {
					stages = m.excludeStageAndRelativesByImageID(stages, sourceImageID)
				}
			}
		}
	}

	currentStage := stage
	for {
		stages = excludeStages(stages, currentStage)
		currentStage = findStageByImageID(stages, currentStage.Info.ParentID)
		if currentStage == nil {
			break
		}
	}

	return stages
}

func excludeStages(stages []*image.StageDescription, stagesToExclude ...*image.StageDescription) []*image.StageDescription {
	var updatedStageList []*image.StageDescription

loop:
	for _, stage := range stages {
		for _, stageToExclude := range stagesToExclude {
			if stage == stageToExclude {
				continue loop
			}
		}

		updatedStageList = append(updatedStageList, stage)
	}

	return updatedStageList
}

func handleDeletionError(err error) error {
	switch err.(type) {
	case docker_registry.DockerHubUnauthorizedError:
		return fmt.Errorf(`%s
You should specify Docker Hub token or username and password to remove tags with Docker Hub API.
Check --repo-docker-hub-token/username/password --repo-docker-hub-token/username/password options.
Be aware that access to the resource is forbidden with personal access token.
Read more details here https://werf.io/documentation/reference/working_with_docker_registries.html#docker-hub`, err)
	case docker_registry.GitHubPackagesUnauthorizedError:
		return fmt.Errorf(`%s
You should specify a token with the read:packages, write:packages, delete:packages and repo scopes to remove package versions.
Check --repo-github-token and --repo-github-token options.
Read more details here https://werf.io/documentation/reference/working_with_docker_registries.html#github-packages`, err)
	default:
		if storage.IsImageDeletionFailedDueToUsingByContainerError(err) {
			return err
		} else if strings.Contains(err.Error(), "UNAUTHORIZED") || strings.Contains(err.Error(), "UNSUPPORTED") {
			return err
		}

		return nil
	}
}
