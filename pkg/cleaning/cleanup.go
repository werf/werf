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
	"github.com/werf/werf/v2/pkg/cleaning/allow_list"
	"github.com/werf/werf/v2/pkg/cleaning/git_history_based_cleanup"
	"github.com/werf/werf/v2/pkg/cleaning/stage_manager"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/logging"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/storage/manager"
	"github.com/werf/werf/v2/pkg/util"
)

type CleanupOptions struct {
	ImageNameList                           []string
	LocalGit                                GitRepo
	KubernetesContextClients                []*kube.ContextClient
	KubernetesNamespaceRestrictionByContext map[string]string
	WithoutKube                             bool // legacy
	ConfigMetaCleanup                       config.MetaCleanup
	KeepStagesBuiltWithinLastNHours         uint64
	DryRun                                  bool
}

func Cleanup(ctx context.Context, projectName string, storageManager *manager.StorageManager, options CleanupOptions) error {
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
		ConfigMetaCleanup:                       options.ConfigMetaCleanup,
		KeepStagesBuiltWithinLastNHours:         options.KeepStagesBuiltWithinLastNHours,
	}
}

type cleanupManager struct {
	stageManager stage_manager.Manager

	nonexistentImportMetadataIDs []string

	ProjectName                             string
	StorageManager                          manager.StorageManagerInterface
	ImageNameList                           []string
	LocalGit                                GitRepo
	KubernetesContextClients                []*kube.ContextClient
	KubernetesNamespaceRestrictionByContext map[string]string
	WithoutKube                             bool
	ConfigMetaCleanup                       config.MetaCleanup
	KeepStagesBuiltWithinLastNHours         uint64
	DryRun                                  bool
}

type GitRepo interface {
	PlainOpen() (*git.Repository, error)
	IsCommitExists(ctx context.Context, commit string) (bool, error)
}

func (m *cleanupManager) init(ctx context.Context) error {
	if err := logboek.Context(ctx).Info().LogProcess("Fetching manifests").DoError(func() error {
		return m.stageManager.InitStageDescSet(ctx, m.StorageManager)
	}); err != nil {
		return err
	}

	if m.StorageManager.GetFinalStagesStorage() != nil {
		if err := logboek.Context(ctx).Info().LogProcess("Fetching final repo manifests").DoError(func() error {
			return m.stageManager.InitFinalStageDescSet(ctx, m.StorageManager)
		}); err != nil {
			return err
		}
	}

	if err := logboek.Context(ctx).Info().LogProcess("Fetching metadata").DoError(func() error {
		if err := m.stageManager.InitImagesMetadata(ctx, m.StorageManager, m.LocalGit, m.ProjectName, m.ImageNameList); err != nil {
			return fmt.Errorf("unable to init images metadata: %w", err)
		}

		if err := m.stageManager.InitCustomTagsMetadata(ctx, m.StorageManager); err != nil {
			return fmt.Errorf("unable to init custom tags metadata: %w", err)
		}

		return nil
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

	if !(m.WithoutKube || m.ConfigMetaCleanup.DisableKubernetesBasedPolicy) {
		if len(m.KubernetesContextClients) == 0 {
			return fmt.Errorf("no kubernetes configs found to skip images being used in the Kubernetes, pass --without-kube option (or WERF_WITHOUT_KUBE env var) to suppress this error")
		}

		deployedDockerImages, err := m.deployedDockerImages(ctx)
		if err != nil {
			return fmt.Errorf("error getting deployed docker images names from Kubernetes: %w", err)
		}

		if err := logboek.Context(ctx).LogProcess("Skipping repo tags that are being used in Kubernetes").DoError(func() error {
			return m.skipStageIDsThatAreUsedInKubernetes(ctx, deployedDockerImages)
		}); err != nil {
			return err
		}

		if err := logboek.Context(ctx).LogProcess("Skipping final repo tags that are being used in Kubernetes").DoError(func() error {
			return m.skipFinalStageIDsThatAreUsedInKubernetes(ctx, deployedDockerImages)
		}); err != nil {
			return err
		}
	}

	if !m.ConfigMetaCleanup.DisableGitHistoryBasedPolicy {
		if err := logboek.Context(ctx).LogProcess("Git history-based cleanup").DoError(func() error {
			return m.gitHistoryBasedCleanup(ctx)
		}); err != nil {
			return err
		}
	} else {
		if err := m.purgeImageMetadata(ctx); err != nil {
			return err
		}
	}

	// Built within last N hours policy.
	{
		keepImagesBuiltWithinLastNHours := m.ConfigMetaCleanup.KeepImagesBuiltWithinLastNHours
		if m.KeepStagesBuiltWithinLastNHours != 0 {
			keepImagesBuiltWithinLastNHours = m.KeepStagesBuiltWithinLastNHours
		}

		if !(m.ConfigMetaCleanup.DisableBuiltWithinLastNHoursPolicy || keepImagesBuiltWithinLastNHours == 0) {
			reason := fmt.Sprintf("built within last %d hours", keepImagesBuiltWithinLastNHours)
			for stageDescToDelete := range m.stageManager.GetStageDescSet().Iter() {
				if (time.Since(stageDescToDelete.Info.GetCreatedAt()).Hours()) <= float64(keepImagesBuiltWithinLastNHours) {
					m.stageManager.MarkStageDescAsProtected(stageDescToDelete, reason)
				}
			}
		}
	}

	if err := logboek.Context(ctx).LogProcess("Cleanup unused stages").DoError(func() error {
		return m.cleanupUnusedStages(ctx)
	}); err != nil {
		return err
	}

	if m.StorageManager.GetFinalStagesStorage() != nil {
		if err := logboek.Context(ctx).LogProcess("Cleanup final stages").DoError(func() error {
			return m.cleanupFinalStages(ctx)
		}); err != nil {
			return err
		}
	}

	return nil
}

func (m *cleanupManager) purgeImageMetadata(ctx context.Context) error {
	return purgeImageMetadata(ctx, m.ProjectName, m.StorageManager, m.DryRun)
}

func (m *cleanupManager) skipStageIDsThatAreUsedInKubernetes(ctx context.Context, deployedDockerImages []*DeployedDockerImage) error {
	handledDeployedStages := map[string]bool{}
	handleTagFunc := func(tag, stageID string, f func()) {
		dockerImageName := fmt.Sprintf("%s:%s", m.StorageManager.GetStagesStorage().Address(), tag)
		for _, deployedDockerImage := range deployedDockerImages {
			if deployedDockerImage.Name == dockerImageName {
				if !handledDeployedStages[stageID] {
					f()

					logboek.Context(ctx).Default().LogFDetails("tag: %s\n", tag)
					logboek.Context(ctx).Default().LogBlock("used by resources").Do(func() {
						for _, cr := range deployedDockerImage.ContextResources {
							for _, r := range cr.ResourcesNames {
								logboek.Context(ctx).Default().LogF("ctx/%s %s\n", cr.ContextName, r)
							}
						}
					})

					logboek.Context(ctx).LogOptionalLn()
					handledDeployedStages[stageID] = true
				}

				break
			}
		}
	}

	for stageDesc := range m.stageManager.GetStageDescSet().Iter() {
		tag := stageDesc.StageID.String()
		stageID := stageDesc.StageID.String()

		handleTagFunc(tag, stageID, func() {
			m.stageManager.MarkStageDescAsProtected(stageDesc, "used in the Kubernetes")
		})
	}

	for stageID, customTagList := range m.stageManager.GetCustomTagsMetadata() {
		for _, customTag := range customTagList {
			handleTagFunc(customTag, stageID, func() {
				stageDesc := m.stageManager.GetStageDescByStageID(stageID)
				if stageDesc != nil {
					// keep existent stage and associated custom tags
					m.stageManager.MarkStageDescAsProtected(stageDesc, "used in the Kubernetes")
				} else {
					// keep custom tags that do not have associated existent stage
					m.stageManager.ForgetCustomTagsByStageID(stageID)
				}
			})
		}
	}

	return nil
}

func (m *cleanupManager) skipFinalStageIDsThatAreUsedInKubernetes(ctx context.Context, deployedDockerImages []*DeployedDockerImage) error {
	handledDeployedFinalStages := map[string]bool{}
Loop:
	for stageDesc := range m.stageManager.GetFinalStageDescSet().Iter() {
		stageID := stageDesc.StageID.String()
		dockerImageName := fmt.Sprintf("%s:%s", m.StorageManager.GetFinalStagesStorage().Address(), stageID)

		for _, deployedDockerImage := range deployedDockerImages {
			if deployedDockerImage.Name == dockerImageName {
				if !handledDeployedFinalStages[stageID] {
					m.stageManager.MarkFinalStageDescAsProtected(stageDesc, "used in the Kubernetes")

					logboek.Context(ctx).Default().LogFDetails("  tag: %s\n", stageID)
					logboek.Context(ctx).LogOptionalLn()
					handledDeployedFinalStages[stageID] = true
				}

				continue Loop
			}
		}
	}

	return nil
}

type DeployedDockerImage struct {
	Name             string
	ContextResources []*ContextResources
}

type ContextResources struct {
	ContextName    string
	ResourcesNames []string
}

func AppendContextDeployedDockerImages(list []*DeployedDockerImage, contextName string, images []*allow_list.DeployedImage) (res []*DeployedDockerImage) {
	for _, desc := range list {
		res = append(res, &DeployedDockerImage{
			Name:             desc.Name,
			ContextResources: desc.ContextResources,
		})
	}

AppendNewImages:
	for _, i := range images {
		for _, desc := range res {
			if desc.Name == i.Name {
				for _, contextResources := range desc.ContextResources {
					if contextResources.ContextName == contextName {
						contextResources.ResourcesNames = append(contextResources.ResourcesNames, i.ResourcesNames...)
						continue AppendNewImages
					}
				}

				desc.ContextResources = append(desc.ContextResources, &ContextResources{
					ContextName:    contextName,
					ResourcesNames: i.ResourcesNames,
				})
				continue AppendNewImages
			}
		}

		res = append(res, &DeployedDockerImage{
			Name: i.Name,
			ContextResources: []*ContextResources{
				{
					ContextName:    contextName,
					ResourcesNames: i.ResourcesNames,
				},
			},
		})
	}

	return
}

func (m *cleanupManager) deployedDockerImages(ctx context.Context) ([]*DeployedDockerImage, error) {
	var deployedDockerImages []*DeployedDockerImage
	for _, contextClient := range m.KubernetesContextClients {
		if err := logboek.Context(ctx).LogProcessInline("Getting deployed docker images (context %s)", contextClient.ContextName).
			DoError(func() error {
				contextDeployedImages, err := allow_list.DeployedDockerImages(ctx, contextClient.Client, m.KubernetesNamespaceRestrictionByContext[contextClient.ContextName])
				if err != nil {
					return fmt.Errorf("cannot get deployed imagesStageList: %w", err)
				}

				deployedDockerImages = AppendContextDeployedDockerImages(deployedDockerImages, contextClient.ContextName, contextDeployedImages)

				return nil
			}); err != nil {
			return nil, err
		}
	}

	return deployedDockerImages, nil
}

func (m *cleanupManager) gitHistoryBasedCleanup(ctx context.Context) error {
	gitRepository, err := m.LocalGit.PlainOpen()
	if err != nil {
		return fmt.Errorf("git plain open failed: %w", err)
	}

	var referencesToScan []*git_history_based_cleanup.ReferenceToScan
	if err := logboek.Context(ctx).Default().LogProcess("Preparing references to scan").DoError(func() error {
		referencesToScan, err = git_history_based_cleanup.ReferencesToScan(ctx, gitRepository, m.ConfigMetaCleanup.KeepPolicies)
		return err
	}); err != nil {
		return err
	}

	for imageName, stageIDCommitList := range m.stageManager.GetImageStageIDCommitListToCleanup() {
		var reachedStageIDs []string
		var hitStageIDCommitList map[string][]string
		// TODO(multiarch): iterate target platforms
		if err := logboek.Context(ctx).LogProcess(logging.ImageLogProcessName(imageName, false, "")).DoError(func() error {
			if logboek.Context(ctx).Streams().Width() > 120 {
				m.printStageIDCommitListTable(ctx, imageName)
				m.printStageIDCustomTagListTable(ctx)
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
		return fmt.Errorf("ubable to cleanup nonexistent image metadata: %w", err)
	}

	return nil
}

func (m *cleanupManager) printStageIDCommitListTable(ctx context.Context, imageName string) {
	stageIDCommitList := m.stageManager.GetStageIDCommitListToCleanup(imageName)
	if countStageIDCommitList(stageIDCommitList) == 0 {
		return
	}

	rows := m.prepareStageIDTableRows(ctx, stageIDCommitList)
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

func (m *cleanupManager) printStageIDCustomTagListTable(ctx context.Context) {
	stageIDCustomTagList := m.stageManager.GetCustomTagsMetadata()
	if len(stageIDCustomTagList) == 0 {
		return
	}

	rows := m.prepareStageIDTableRows(ctx, stageIDCustomTagList)
	if len(rows) != 0 {
		tbl := table.New("Tag", "Custom tags")
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

func (m *cleanupManager) prepareStageIDTableRows(ctx context.Context, stageIDCustomTagOrCommitList map[string][]string) [][]interface{} {
	var rows [][]interface{}
	for stageID, anyList := range stageIDCustomTagOrCommitList {
		for ind, customTagOrCommit := range anyList {
			var columns []interface{}
			if ind == 0 {
				stageIDColumn := stageID

				space := len(stageID) - len(customTagOrCommit) - 1
				if logboek.Context(ctx).Streams().ContentWidth() < space {
					stageIDColumn = fmt.Sprintf("%s..%s", stageID[:space-5], stageID[space-3:])
				}

				columns = append(columns, stageIDColumn)
			} else {
				columns = append(columns, "")
			}

			columns = append(columns, customTagOrCommit)
			rows = append(rows, columns)
		}
	}

	return rows
}

func (m *cleanupManager) handleSavedStageIDs(ctx context.Context, savedStageIDs []string) {
	logboek.Context(ctx).Default().LogBlock("Saved tags").Do(func() {
		for _, stageID := range savedStageIDs {
			m.stageManager.MarkStageDescAsProtectedByStageID(stageID, "found in the git history")
			logboek.Context(ctx).Default().LogFDetails("  tag: %s\n", stageID)
			logboek.Context(ctx).LogOptionalLn()
		}
	})
}

func (m *cleanupManager) deleteStages(ctx context.Context, stageDescSet image.StageDescSet, isFinal bool) error {
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

	return deleteStageDescSet(ctx, m.StorageManager, m.DryRun, deleteStageOptions, stageDescSet, isFinal)
}

func deleteStageDescSet(ctx context.Context, storageManager manager.StorageManagerInterface, dryRun bool, deleteStageOptions manager.ForEachDeleteStageOptions, stageDescSet image.StageDescSet, isFinal bool) error {
	if dryRun {
		for stageDesc := range stageDescSet.Iter() {
			logboek.Context(ctx).Default().LogFDetails("  tag: %s\n", stageDesc.StageID.String())
			logboek.Context(ctx).LogOptionalLn()
		}
		return nil
	}

	onDeleteFunc := func(ctx context.Context, stageDesc *image.StageDesc, err error) error {
		if err != nil {
			if err := handleDeletionError(err); err != nil {
				return err
			}

			logboek.Context(ctx).Warn().LogF("WARNING: Image %s deletion failed: %s\n", stageDesc.Info.Name, err)

			return nil
		}

		logboek.Context(ctx).Default().LogFDetails("  tag: %s\n", stageDesc.Info.Tag)

		return nil
	}

	if isFinal {
		return storageManager.ForEachDeleteFinalStage(ctx, deleteStageOptions, stageDescSet, onDeleteFunc)
	}

	return storageManager.ForEachDeleteStage(ctx, deleteStageOptions, stageDescSet, onDeleteFunc)
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

			logProcessDoError := logboek.Context(ctx).Default().LogProcess(header).DoError
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

func purgeImageMetadata(ctx context.Context, projectName string, storageManager manager.StorageManagerInterface, dryRun bool) error {
	return logboek.Context(ctx).Default().LogProcess("Deleting images metadata").DoError(func() error {
		_, imageMetadataByImageName, err := storageManager.GetStagesStorage().GetAllAndGroupImageMetadataByImageName(ctx, projectName, []string{}, storage.WithCache())
		if err != nil {
			return err
		}

		for imageNameID, stageIDCommitList := range imageMetadataByImageName {
			if err := deleteImageMetadata(ctx, projectName, storageManager, imageNameID, stageIDCommitList, dryRun); err != nil {
				return err
			}
		}

		return nil
	})
}

func deleteImageMetadata(ctx context.Context, projectName string, storageManager manager.StorageManagerInterface, imageNameOrID string, stageIDCommitList map[string][]string, dryRun bool) error {
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

			logboek.Context(ctx).Warn().LogF("WARNING: Image metadata %q commit %s stage ID %s deletion failed: %s\n", imageNameOrID, commit, stageID, err)

			return nil
		}

		logboek.Context(ctx).Info().LogFDetails("  imageName: %s\n", imageNameOrID)
		logboek.Context(ctx).Info().LogFDetails("  stageID: %s\n", stageID)
		logboek.Context(ctx).Info().LogFDetails("  commit: %s\n", commit)

		return nil
	})
}

func (m *cleanupManager) cleanupUnusedStages(ctx context.Context) error {
	stageDescCount := m.stageManager.GetStageDescSet().Cardinality()

	if err := logboek.Context(ctx).Info().LogProcess("Fetching imports metadata").DoError(func() error {
		return m.initImportsMetadata(ctx)
	}); err != nil {
		return fmt.Errorf("unable to init imports metadata: %w", err)
	}

	// skip stages and their relatives covered by Kubernetes- or git history-based cleanup policies
	stageDescSetToDelete := stageDescSet

	{
		excludedSDListByReason := make(map[string]image.StageDescSet)

		for reason, sdList := range m.stageManager.GetProtectedStageDescListByReason() {
			for _, sd := range sdList {
				var excludedSDListBySD image.StageDescSet

				stageDescSetToDelete, excludedSDListBySD = m.excludeStageAndRelativesByImage(stageDescSetToDelete, sd.Info)

				for _, exclSD := range excludedSDListBySD {
					if sd.Info.Name == exclSD.Info.Name {
						excludedSDListByReason[reason] = append(excludedSDListByReason[reason], exclSD)
					} else {
						ancestorReason := fmt.Sprintf("ancestors of images %s", reason)
						excludedSDListByReason[ancestorReason] = append(excludedSDListByReason[ancestorReason], exclSD)
					}
				}
			}
		}

		excludedCount := 0
		for _, list := range excludedSDListByReason {
			excludedCount += len(list)
		}

		logboek.Context(ctx).Default().LogBlock("Saved stages (%d/%d)", excludedCount, len(stageDescSet)).Do(func() {
			for reason, list := range excludedSDListByReason {
				logboek.Context(ctx).Default().LogProcess("%s (%d)", reason, len(list)).Do(func() {
					for _, excludedSD := range list {
						logboek.Context(ctx).Default().LogFDetails("%s\n", excludedSD.Info.Tag)
					}
				})
			}
		})
	}

	if len(stageDescSetToDelete) != 0 {
		if err := logboek.Context(ctx).Default().LogProcess("Deleting stages tags (%d/%d)", len(stageDescSetToDelete), stageDescSetCount).DoError(func() error {
			return m.deleteStages(ctx, stageDescSetToDelete, false)
		}); err != nil {
			return err
		}

		m.stageManager.ForgetDeletedStages(stageDescSetToDelete)
	}

	if err := m.deleteUnusedCustomTags(ctx); err != nil {
		return fmt.Errorf("unable to cleanup custom tags metadata: %w", err)
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

func (m *cleanupManager) cleanupFinalStages(ctx context.Context) error {
	finalStagesDescriptionListFull := m.stageManager.GetFinalStageDescList(stage_manager.StageDescListOptions{})
	finalStageDescListFullCount := len(finalStagesDescriptionListFull)

	finalStagesDescriptionList := m.stageManager.GetFinalStageDescList(stage_manager.StageDescListOptions{ExcludeProtected: true})
	stagesDescriptionList := m.stageManager.GetStageDescList(stage_manager.StageDescListOptions{})

	var finalStagesDescriptionListToDelete image.StageDescSet

FilterOutFinalStages:
	for _, finalStg := range finalStagesDescriptionList {
		for _, stg := range stagesDescriptionList {
			if stg.StageID.IsEqual(*finalStg.StageID) {
				continue FilterOutFinalStages
			}
		}

		finalStagesDescriptionListToDelete = append(finalStagesDescriptionListToDelete, finalStg)
	}

	if len(finalStagesDescriptionListToDelete) != 0 {
		if err := logboek.Context(ctx).Default().LogProcess("Deleting final stages tags (%d/%d)", len(finalStagesDescriptionListToDelete), finalStageDescListFullCount).DoError(func() error {
			return m.deleteStages(ctx, finalStagesDescriptionListToDelete, true)
		}); err != nil {
			return err
		}

		m.stageManager.ForgetDeletedFinalStages(finalStagesDescriptionListToDelete)
	}

	return nil
}

func (m *cleanupManager) initImportsMetadata(ctx context.Context) error {
	importMetadataIDs, err := m.StorageManager.GetStagesStorage().GetImportMetadataIDs(ctx, m.ProjectName, storage.WithCache())
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
				return fmt.Errorf("unable to delete import metadata %s: %w", metadataID, err)
			}

			return nil
		}

		mutex.Lock()
		defer mutex.Unlock()

		sourceImageID := metadata.SourceImageID
		importSourceID := metadata.ImportSourceID
		stage := findStageDescByImageID(m.stageManager.GetStageDescSet(), sourceImageID)
		if stage == nil {
			m.nonexistentImportMetadataIDs = append(m.nonexistentImportMetadataIDs, importSourceID)
		}

		return nil
	})
}

func (m *cleanupManager) deleteImportsMetadata(ctx context.Context, importMetadataIDs []string) error {
	return deleteImportsMetadata(ctx, m.ProjectName, m.StorageManager, importMetadataIDs, m.DryRun)
}

func deleteImportsMetadata(ctx context.Context, projectName string, storageManager manager.StorageManagerInterface, importMetadataIDs []string, dryRun bool) error {
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

func (m *cleanupManager) getRelativeStageDescSetByImageInfo(stageDescSet image.StageDescSet, imageInfo *image.Info) image.StageDescSet {
	parentStageDescSet := findStageDescSetByImageID(stageDescSet, imageInfo)
	if parentStageDescSet.IsEmpty() {
		return image.NewStageDescSet()
	}
	return m.getRelativeStageDescSet(stageDescSet, parentStageDescSet)
}

// findStageDescSetByMatch could return multiple stages when target image.Info is an image index
func findStageDescSetByMatch(stageDescSet image.StageDescSet, target *image.Info, imageMatcher, indexMatcher func(stageDesc *image.StageDesc, target *image.Info) bool) image.StageDescSet {
	resultStageDescSet := image.NewStageDescSet()
	if target.IsIndex {
		if indexMatcher != nil {
			for stg := range stageDescSet.Iter() {
				if indexMatcher(stg, target) {
					resultStageDescSet.Add(stg)
				}
			}
		}
		for _, subimg := range target.Index {
			resultStageDescSet = resultStageDescSet.Union(findStageDescSetByMatch(stageDescSet, subimg, imageMatcher, indexMatcher))
		}
	} else if imageMatcher != nil {
		for stg := range stageDescSet.Iter() {
			if imageMatcher(stg, target) {
				resultStageDescSet.Add(stg)
			}
		}
	}

	return resultStageDescSet
}

func findStageDescSetByImageID(stageDescSet image.StageDescSet, target *image.Info) image.StageDescSet {
	return findStageDescSetByMatch(
		stageDescSet, target,
		func(stageDesc *image.StageDesc, target *image.Info) bool {
			return stageDesc.Info.ID == target.ID
		},
		func(stageDesc *image.StageDesc, target *image.Info) bool {
			return stageDesc.Info.Name == target.Name
		},
	)
}

func findStageDescSetByImageParentID(stageDescSet image.StageDescSet, target *image.Info) image.StageDescSet {
	return findStageDescSetByImage(
		stageDescSet, target,
		func(stageDesc *image.StageDesc, target *image.Info) bool {
			return stageDesc.Info.ID == target.ParentID
		}, nil,
	)
}

func findStageDescByImageID(stageDescSet image.StageDescSet, imageID string) *image.StageDesc {
	for stage := range stageDescSet.Iter() {
		if stage.Info.ID == imageID {
			return stage
		}
	}
	return nil
}

func (m *cleanupManager) getRelativeStageDescSet(stageDescSet, parentStageDescSet image.StageDescSet) image.StageDescSet {
	currentStageDescSet := parentStageDescSet
	relativeStageDescSet := image.NewStageDescSet()
	for !currentStageDescSet.IsEmpty() {
		relativeStageDescSet = relativeStageDescSet.Union(currentStageDescSet)
		nextStageDescSet := image.NewStageDescSet()
		for stageDescToExclude := range currentStageDescSet.Iter() {
			relativeStageDescSetByParentID := findStageDescSetByImageParentID(stageDescSet, stageDescToExclude.Info)
			nextStageDescSet = nextStageDescSet.Union(relativeStageDescSetByParentID)
		}

		currentStageDescSet = nextStageDescSet
	}

	return relativeStageDescSet
}

func (m *cleanupManager) deleteUnusedCustomTags(ctx context.Context) error {
	stageIDCustomTagList := m.stageManager.GetCustomTagsMetadata()
	if len(stageIDCustomTagList) == 0 {
		return nil
	}

	var customTagListToDelete []string
	var customTagListToKeep []string
	var numberOfCustomTags int
	for stageID, customTagList := range stageIDCustomTagList {
		numberOfCustomTags += len(customTagList)
		if !m.stageManager.ContainsStageDescByStageID(stageID) {
			customTagListToDelete = append(customTagListToDelete, customTagList...)
		} else {
			customTagListToKeep = append(customTagListToKeep, customTagList...)
		}
	}

	if len(customTagListToKeep) != 0 {
		header := fmt.Sprintf("Saved custom tags (%d/%d)", len(customTagListToKeep), numberOfCustomTags)
		logboek.Context(ctx).Default().LogBlock(header).Do(func() {
			for _, customTag := range customTagListToKeep {
				logboek.Context(ctx).Default().LogFDetails("  tag: %s\n", customTag)
				logboek.Context(ctx).LogOptionalLn()
			}
		})
	}

	if len(customTagListToDelete) != 0 {
		header := fmt.Sprintf("Deleting unused custom tags (%d/%d)", len(customTagListToDelete), numberOfCustomTags)
		if err := logboek.Context(ctx).LogProcess(header).DoError(func() error {
			if err := deleteCustomTags(ctx, m.StorageManager, customTagListToDelete, m.DryRun); err != nil {
				return err
			}

			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

func handleDeletionError(err error) error {
	switch {
	case docker_registry.IsDockerHubUnauthorizedErr(err):
		return fmt.Errorf(`%w

You should specify Docker Hub token or username and password to remove tags with Docker Hub API.
Check --repo-docker-hub-token, --repo-docker-hub-username and --repo-docker-hub-password options.
Be aware that access to the resource is forbidden with personal access token.`, err)
	case docker_registry.IsGitHubPackagesUnauthorizedErr(err), docker_registry.IsGitHubPackagesForbiddenErr(err):
		return fmt.Errorf(`%w

You should specify a token with delete:packages and read:packages scopes to remove package versions.
Check --repo-github-token option.
Be aware that the token provided to GitHub Actions workflow is not enough to remove package versions.`, err)
	default:
		if storage.IsImageDeletionFailedDueToUsingByContainerErr(err) {
			return err
		} else if strings.Contains(err.Error(), "UNAUTHORIZED") || strings.Contains(err.Error(), "UNSUPPORTED") {
			return err
		}

		return nil
	}
}
