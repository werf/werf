package cleaning

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/gookit/color"
	"github.com/rodaine/table"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/cleaning/allow_list"
	"github.com/werf/werf/pkg/cleaning/git_history_based_cleanup"
	"github.com/werf/werf/pkg/cleaning/stage_manager"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/logging"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/storage/manager"
	"github.com/werf/werf/pkg/util"
)

type CleanupOptions struct {
	ImageNameList                           []string
	LocalGit                                GitRepo
	KubernetesContextClients                []*kube.ContextClient
	KubernetesNamespaceRestrictionByContext map[string]string
	WithoutKube                             bool
	GitHistoryBasedCleanupOptions           config.MetaCleanup
	KeepStagesBuiltWithinLastNHours         uint64
	DryRun                                  bool
}

func Cleanup(ctx context.Context, projectName string, storageManager *manager.StorageManager, storageLockManager storage.LockManager, options CleanupOptions) error {
	return newCleanupManager(projectName, storageManager, options).run(ctx)
}

func newCleanupManager(projectName string, storageManager *manager.StorageManager, options CleanupOptions) *cleanupManager {
	return &cleanupManager{
		stageManager:                            stage_manager.NewManager(),
		ProjectName:                             projectName,
		StorageManager:                          storageManager,
		ImageNameList:                           options.ImageNameList,
		DryRun:                                  options.DryRun,
		LocalGit:                                options.LocalGit,
		KubernetesContextClients:                options.KubernetesContextClients,
		KubernetesNamespaceRestrictionByContext: options.KubernetesNamespaceRestrictionByContext,
		WithoutKube:                             options.WithoutKube,
		GitHistoryBasedCleanupOptions:           options.GitHistoryBasedCleanupOptions,
		KeepStagesBuiltWithinLastNHours:         options.KeepStagesBuiltWithinLastNHours,
	}
}

type cleanupManager struct {
	stageManager stage_manager.Manager

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
	KeepStagesBuiltWithinLastNHours         uint64
	DryRun                                  bool
}

type GitRepo interface {
	PlainOpen() (*git.Repository, error)
	IsCommitExists(ctx context.Context, commit string) (bool, error)
}

func (m *cleanupManager) init(ctx context.Context) error {
	if err := logboek.Context(ctx).Info().LogProcess("Fetching manifests").DoError(func() error {
		return m.stageManager.InitStages(ctx, m.StorageManager)
	}); err != nil {
		return err
	}

	if err := logboek.Context(ctx).Info().LogProcess("Fetching metadata").DoError(func() error {
		return m.stageManager.InitImagesMetadata(ctx, m.StorageManager, m.LocalGit, m.ProjectName, m.ImageNameList)
	}); err != nil {
		return err
	}

	return nil
}

func (m *cleanupManager) run(ctx context.Context) error {
	if err := logboek.Context(ctx).LogProcess("Fetching manifests and metadata").DoError(func() error {
		return m.init(ctx)
	}); err != nil {
		return err
	}

	if m.LocalGit != nil {
		if !m.WithoutKube {
			if err := logboek.Context(ctx).LogProcess("Skipping tags that are being used in Kubernetes").DoError(func() error {
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

	handledDeployedStages := map[string]bool{}
Loop:
	for _, stageID := range m.stageManager.GetStageIDList() {
		dockerImageName := fmt.Sprintf("%s:%s", m.StorageManager.StagesStorage.String(), stageID)
		for _, deployedDockerImageName := range deployedDockerImagesNames {
			if deployedDockerImageName == dockerImageName {
				if !handledDeployedStages[stageID] {
					m.stageManager.MarkStageAsProtected(stageID)

					logboek.Context(ctx).Default().LogFDetails("  tag: %s\n", stageID)
					logboek.Context(ctx).LogOptionalLn()
					handledDeployedStages[stageID] = true
				}

				continue Loop
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

	for imageName, stageIDCommitList := range m.stageManager.GetImageStageIDCommitListToCleanup() {
		var reachedStageIDs []string
		var hitStageIDCommitList map[string][]string
		if err := logboek.Context(ctx).LogProcess(logging.ImageLogProcessName(imageName, false)).DoError(func() error {
			if logboek.Context(ctx).Streams().Width() > 90 {
				m.printStageIDCommitListTable(ctx, imageName)
			}

			if err := logboek.Context(ctx).LogProcess("Scanning git references history").DoError(func() error {
				if countStageIDCommitList(stageIDCommitList) != 0 {
					reachedStageIDs, hitStageIDCommitList, err = git_history_based_cleanup.ScanReferencesHistory(ctx, gitRepository, referencesToScan, stageIDCommitList)
				} else {
					logboek.Context(ctx).LogLn("Scanning stopped due to nothing to seek")
				}

				return nil
			}); err != nil {
				return err
			}

			var stageIDToUnlink []string
		outerLoop:
			for stageID := range stageIDCommitList {
				for _, reachedStageID := range reachedStageIDs {
					if stageID == reachedStageID {
						continue outerLoop
					}
				}

				stageIDToUnlink = append(stageIDToUnlink, stageID)
			}

			if len(reachedStageIDs) != 0 {
				m.handleSavedStageIDs(ctx, reachedStageIDs)
			}

			if err := logboek.Context(ctx).LogProcess("Cleaning image metadata").DoError(func() error {
				return m.cleanupImageMetadata(ctx, imageName, hitStageIDCommitList, stageIDToUnlink)
			}); err != nil {
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
	if logboek.Context(ctx).Streams().ContentWidth() < 120 {
		return
	}

	stageIDCommitList := m.stageManager.GetStageIDCommitListToCleanup(imageName)
	if countStageIDCommitList(stageIDCommitList) == 0 {
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
		tbl.WithWriter(logboek.Context(ctx).OutStream())
		tbl.WithHeaderFormatter(func(format string, a ...interface{}) string {
			return logboek.ColorizeF(color.New(color.OpUnderscore), format, a...)
		})
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
			m.stageManager.MarkStageAsProtected(stageID)
			logboek.Context(ctx).Default().LogFDetails("  tag: %s\n", stageID)
			logboek.Context(ctx).LogOptionalLn()
		}
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

func (m *cleanupManager) cleanupImageMetadata(ctx context.Context, imageName string, hitStageIDCommitList map[string][]string, stageIDsToUnlink []string) error {
	if countStageIDCommitList(hitStageIDCommitList) != 0 || len(stageIDsToUnlink) != 0 {
		stageIDCommitListToDelete := map[string][]string{}
		stageIDCommitListToCleanup := m.stageManager.GetStageIDCommitListToCleanup(imageName)
	stageIDCommitListLoop:
		for stageID, commitList := range stageIDCommitListToCleanup {
			commitListToCheck := commitList
			for _, stageIDToUnlink := range stageIDsToUnlink {
				if stageIDToUnlink == stageID {
					stageIDCommitListToDelete[stageID] = commitListToCheck
					continue stageIDCommitListLoop
				}
			}

			hitCommitList, ok := hitStageIDCommitList[stageID]
			if ok {
				commitListToCheck = util.ExcludeFromStringArray(commitListToCheck, hitCommitList...)
			}

			stageIDCommitListToDelete[stageID] = commitListToCheck
		}

		if countStageIDCommitList(stageIDCommitListToDelete) != 0 {
			header := fmt.Sprintf("Cleaning up metadata (%d/%d)", countStageIDCommitList(stageIDCommitListToDelete), countStageIDCommitList(stageIDCommitListToCleanup))

			logProcessDoError := logboek.Context(ctx).Default().LogProcessInline(header).DoError
			if logboek.Context(ctx).Info().IsAccepted() {
				logProcessDoError = logboek.Context(ctx).Info().LogProcess(header).DoError
			}

			if err := logProcessDoError(func() error {
				return m.deleteImageMetadata(ctx, imageName, stageIDCommitListToDelete)
			}); err != nil {
				return err
			}
		}
	}

	nonexistentStageIDCommitList := m.stageManager.GetNonexistentStageIDCommitList(imageName)
	if countStageIDCommitList(nonexistentStageIDCommitList) != 0 {
		header := fmt.Sprintf("Deleting metadata for nonexistent stageIDs (%d)", countStageIDCommitList(nonexistentStageIDCommitList))

		logProcessDoError := logboek.Context(ctx).Default().LogProcessInline(header).DoError
		if logboek.Context(ctx).Info().IsAccepted() {
			logProcessDoError = logboek.Context(ctx).Info().LogProcess(header).DoError
		}

		if err := logProcessDoError(func() error {
			return m.deleteImageMetadata(ctx, imageName, nonexistentStageIDCommitList)
		}); err != nil {
			return err
		}
	}

	stageIDNonexistentCommitList := m.stageManager.GetStageIDNonexistentCommitList(imageName)
	if countStageIDCommitList(stageIDNonexistentCommitList) != 0 {
		header := fmt.Sprintf("Deleting metadata for nonexistent commits (%d)", countStageIDCommitList(stageIDNonexistentCommitList))

		logProcessDoError := logboek.Context(ctx).Default().LogProcessInline(header).DoError
		if logboek.Context(ctx).Info().IsAccepted() {
			logProcessDoError = logboek.Context(ctx).Info().LogProcess(header).DoError
		}

		if err := logProcessDoError(func() error {
			return m.deleteImageMetadata(ctx, imageName, stageIDNonexistentCommitList)
		}); err != nil {
			return err
		}
	}

	return nil
}

func countStageIDCommitList(stageIDCommitList map[string][]string) int {
	var result int
	for _, commitList := range stageIDCommitList {
		for range commitList {
			result++
		}
	}

	return result
}

func (m *cleanupManager) cleanupNonexistentImageMetadata(ctx context.Context) error {
	var counter int
	stageIDCommitListByNonexistentImage := m.stageManager.GetStageIDCommitListByNonexistentImage()
	for _, stageIDCommitList := range stageIDCommitListByNonexistentImage {
		counter += countStageIDCommitList(stageIDCommitList)
	}

	if counter == 0 {
		return nil
	}

	return logboek.Context(ctx).Default().LogProcess("Deleting metadata for nonexistent images (%d)", counter).DoError(func() error {
		for imageName, stageIDCommitList := range stageIDCommitListByNonexistentImage {
			if err := m.deleteImageMetadata(ctx, imageName, stageIDCommitList); err != nil {
				return err
			}
		}

		return nil
	})
}

func (m *cleanupManager) deleteImageMetadata(ctx context.Context, imageName string, stageIDCommitList map[string][]string) error {
	if err := deleteImageMetadata(ctx, m.ProjectName, m.StorageManager, imageName, stageIDCommitList, m.DryRun); err != nil {
		return err
	}

	return nil
}

func deleteImageMetadata(ctx context.Context, projectName string, storageManager *manager.StorageManager, imageNameOrID string, stageIDCommitList map[string][]string, dryRun bool) error {
	if dryRun {
		for stageID, commitList := range stageIDCommitList {
			if len(commitList) == 0 {
				continue
			}

			logboek.Context(ctx).Info().LogFDetails("  imageName: %s\n", imageNameOrID)
			logboek.Context(ctx).Info().LogFDetails("  stageID: %s\n", stageID)
			logboek.Context(ctx).Info().LogFDetails("  commits: %d\n", len(commitList))
			logboek.Context(ctx).Info().LogOptionalLn()
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

		logboek.Context(ctx).Info().LogFDetails("  imageName: %s\n", imageNameOrID)
		logboek.Context(ctx).Info().LogFDetails("  stageID: %s\n", stageID)
		logboek.Context(ctx).Info().LogFDetails("  commit: %s\n", commit)

		return nil
	})
}

func (m *cleanupManager) cleanupUnusedStages(ctx context.Context) error {
	stageDescriptionList := m.stageManager.GetStageDescriptionList()
	stageDescriptionListCount := len(stageDescriptionList)

	if err := logboek.Context(ctx).Info().LogProcess("Fetching imports metadata").DoError(func() error {
		return m.initImportsMetadata(ctx, stageDescriptionList)
	}); err != nil {
		return fmt.Errorf("unable to init imports metadata: %s", err)
	}

	stageDescriptionListToDelete := stageDescriptionList
	// skip stages and their relatives based on deployed images in k8s and git history based cleanup policies
	{
		var excludedSDList []*image.StageDescription
		for _, sd := range m.stageManager.GetProtectedStageDescriptionList() {
			var excludedSDListBySD []*image.StageDescription
			stageDescriptionListToDelete, excludedSDListBySD = m.excludeStageAndRelativesByImageID(stageDescriptionListToDelete, sd.Info.ID)
			excludedSDList = append(excludedSDList, excludedSDListBySD...)
		}

		logboek.Context(ctx).Default().LogBlock("Saved stages (%d/%d)", len(excludedSDList), len(stageDescriptionList)).Do(func() {
			for _, excludedSD := range excludedSDList {
				logboek.Context(ctx).Default().LogFDetails("  tag: %s\n", excludedSD.Info.Tag)
				logboek.Context(ctx).Default().LogOptionalLn()
			}
		})
	}

	// skip stages and their relatives based on KeepStagesBuiltWithinLastNHours policy
	{
		if m.KeepStagesBuiltWithinLastNHours != 0 {
			var excludedSDList []*image.StageDescription
			for _, sd := range stageDescriptionListToDelete {
				if (time.Since(sd.Info.GetCreatedAt()).Hours()) <= float64(m.KeepStagesBuiltWithinLastNHours) {
					var excludedRelativesSDList []*image.StageDescription
					stageDescriptionListToDelete, excludedRelativesSDList = m.excludeStageAndRelativesByImageID(stageDescriptionListToDelete, sd.Info.ID)
					excludedSDList = append(excludedSDList, excludedRelativesSDList...)
				}
			}

			if len(excludedSDList) != 0 {
				logboek.Context(ctx).Default().LogBlock("Saved stages that were built within last %d hours (%d/%d)", m.KeepStagesBuiltWithinLastNHours, len(excludedSDList), len(stageDescriptionList)).Do(func() {
					for _, stage := range excludedSDList {
						logboek.Context(ctx).Default().LogFDetails("  tag: %s\n", stage.Info.Tag)
						logboek.Context(ctx).LogOptionalLn()
					}
				})
			}
		}
	}

	if len(stageDescriptionListToDelete) != 0 {
		if err := logboek.Context(ctx).Default().LogProcess("Deleting stages tags (%d/%d)", len(stageDescriptionListToDelete), stageDescriptionListCount).DoError(func() error {
			return m.deleteStages(ctx, stageDescriptionListToDelete)
		}); err != nil {
			return err
		}
	}

	if len(m.nonexistentImportMetadataIDs) != 0 {
		if err := logboek.Context(ctx).Default().LogProcess("Cleaning imports metadata (%d)", len(m.nonexistentImportMetadataIDs)).DoError(func() error {
			return m.deleteImportsMetadata(ctx, m.nonexistentImportMetadataIDs)
		}); err != nil {
			return err
		}
	}

	return nil
}

func (m *cleanupManager) initImportsMetadata(ctx context.Context, stageDescriptionList []*image.StageDescription) error {
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
			if err := logboek.Context(ctx).Warn().LogProcess("Deleting invalid import metadata %s", metadataID).
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

		stage := findStageByImageID(stageDescriptionList, sourceImageID)
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
			logboek.Context(ctx).Info().LogFDetails("  importMetadataID: %s\n", importMetadataID)
			logboek.Context(ctx).Info().LogOptionalLn()
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

		logboek.Context(ctx).Info().LogFDetails("  importMetadataID: %s\n", importMetadataID)

		return nil
	})
}

func (m *cleanupManager) excludeStageAndRelativesByImageID(stages []*image.StageDescription, imageID string) ([]*image.StageDescription, []*image.StageDescription) {
	stage := findStageByImageID(stages, imageID)
	if stage == nil {
		return stages, []*image.StageDescription{}
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

func (m *cleanupManager) excludeStageAndRelativesByStage(stages []*image.StageDescription, stage *image.StageDescription) ([]*image.StageDescription, []*image.StageDescription) {
	var excludedStages []*image.StageDescription
	currentStage := stage
	for {
		stages = excludeStages(stages, currentStage)
		excludedStages = append(excludedStages, currentStage)

		currentStage = findStageByImageID(stages, currentStage.Info.ParentID)
		if currentStage == nil {
			break
		}
	}

	for label, checksum := range stage.Info.Labels {
		if strings.HasPrefix(label, image.WerfImportChecksumLabelPrefix) {
			sourceImageIDs, ok := m.checksumSourceImageIDs[checksum]
			if ok {
				for _, sourceImageID := range sourceImageIDs {
					var excludedImportStages []*image.StageDescription
					stages, excludedImportStages = m.excludeStageAndRelativesByImageID(stages, sourceImageID)
					excludedStages = append(excludedStages, excludedImportStages...)
				}
			}
		}
	}

	return stages, excludedStages
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
