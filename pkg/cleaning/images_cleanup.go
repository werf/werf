package cleaning

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/rodaine/table"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"

	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/logging"
	"github.com/werf/werf/pkg/slug"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/storage/manager"
	"github.com/werf/werf/pkg/tag_strategy"
)

type ImagesCleanupOptions struct {
	ImageNameList                           []string
	LocalGit                                GitRepo
	KubernetesContextClients                []*kube.ContextClient
	KubernetesNamespaceRestrictionByContext map[string]string
	WithoutKube                             bool
	Policies                                ImagesCleanupPolicies
	GitHistoryBasedCleanup                  bool
	GitHistoryBasedCleanupV12               bool
	GitHistoryBasedCleanupOptions           config.MetaCleanup
	DryRun                                  bool
}

func ImagesCleanup(ctx context.Context, projectName string, storageManager *manager.StorageManager, storageLockManager storage.LockManager, options ImagesCleanupOptions) error {
	m := newImagesCleanupManager(projectName, storageManager, options)

	if lock, err := storageLockManager.LockStagesAndImages(ctx, projectName, storage.LockStagesAndImagesOptions{GetOrCreateImagesOnly: false}); err != nil {
		return fmt.Errorf("unable to lock stages and images: %s", err)
	} else {
		defer storageLockManager.Unlock(ctx, lock)
	}

	return logboek.Context(ctx).Default().LogProcess("Running images cleanup").
		Options(func(options types.LogProcessOptionsInterface) {
			options.Style(style.Highlight())
		}).
		DoError(func() error {
			return m.run(ctx)
		})
}

func newImagesCleanupManager(projectName string, storageManager *manager.StorageManager, options ImagesCleanupOptions) *imagesCleanupManager {
	return &imagesCleanupManager{
		ProjectName:                             projectName,
		StorageManager:                          storageManager,
		ImageNameList:                           options.ImageNameList,
		DryRun:                                  options.DryRun,
		LocalGit:                                options.LocalGit,
		KubernetesContextClients:                options.KubernetesContextClients,
		KubernetesNamespaceRestrictionByContext: options.KubernetesNamespaceRestrictionByContext,
		WithoutKube:                             options.WithoutKube,
		Policies:                                options.Policies,
		GitHistoryBasedCleanup:                  options.GitHistoryBasedCleanup,
		GitHistoryBasedCleanupV12:               options.GitHistoryBasedCleanupV12,
		GitHistoryBasedCleanupOptions:           options.GitHistoryBasedCleanupOptions,
	}
}

type imagesCleanupManager struct {
	imageRepoImageList           *map[string][]*image.Info
	imageCommitHashImageMetadata *map[string]map[plumbing.Hash]*storage.ImageMetadata

	ProjectName                             string
	StorageManager                          *manager.StorageManager
	ImageNameList                           []string
	LocalGit                                GitRepo
	KubernetesContextClients                []*kube.ContextClient
	KubernetesNamespaceRestrictionByContext map[string]string
	WithoutKube                             bool
	Policies                                ImagesCleanupPolicies
	GitHistoryBasedCleanup                  bool
	GitHistoryBasedCleanupV12               bool
	GitHistoryBasedCleanupOptions           config.MetaCleanup
	DryRun                                  bool
}

type GitRepo interface {
	PlainOpen() (*git.Repository, error)
	IsCommitExists(ctx context.Context, commit string) (bool, error)
	TagsList(ctx context.Context) ([]string, error)
	RemoteBranchesList(ctx context.Context) ([]string, error)
}

type ImagesCleanupPolicies struct {
	GitTagStrategyHasLimit bool // No limit by default!
	GitTagStrategyLimit    int64

	GitTagStrategyHasExpiryPeriod bool // No expiration by default!
	GitTagStrategyExpiryPeriod    time.Duration

	GitCommitStrategyHasLimit bool // No limit by default!
	GitCommitStrategyLimit    int64

	GitCommitStrategyHasExpiryPeriod bool // No expiration by default!
	GitCommitStrategyExpiryPeriod    time.Duration

	StagesSignatureStrategyHasLimit bool // No limit by default!
	StagesSignatureStrategyLimit    int64

	StagesSignatureStrategyHasExpiryPeriod bool // No expiration by default!
	StagesSignatureStrategyExpiryPeriod    time.Duration
}

func (m *imagesCleanupManager) initRepoImagesData(ctx context.Context) error {
	if err := logboek.Context(ctx).Info().LogProcess("Fetching repo images").DoError(func() error {
		return m.initRepoImages(ctx)
	}); err != nil {
		return err
	}

	if m.GitHistoryBasedCleanup || m.GitHistoryBasedCleanupV12 {
		if err := logboek.Context(ctx).Info().LogProcess("Fetching images metadata").DoError(func() error {
			return m.initImageCommitHashImageMetadata(ctx)
		}); err != nil {
			return err
		}
	}

	return nil
}

func (m *imagesCleanupManager) initRepoImages(ctx context.Context) error {
	repoImages, err := selectRepoImagesFromImagesRepo(ctx, m.StorageManager, m.ImageNameList)
	if err != nil {
		return err
	}

	m.setImageRepoImageList(repoImages)

	return nil
}

func (m *imagesCleanupManager) initImageCommitHashImageMetadata(ctx context.Context) error {
	imageCommitImageMetadata := map[string]map[plumbing.Hash]*storage.ImageMetadata{}
	for _, imageName := range m.ImageNameList {
		var mutex sync.Mutex
		commitImageMetadata := map[plumbing.Hash]*storage.ImageMetadata{}
		if err := m.StorageManager.ForEachGetImageMetadataByCommit(ctx, m.ProjectName, imageName, func(commit string, imageMetadata *storage.ImageMetadata, err error) error {
			if err != nil {
				return err
			}

			mutex.Lock()
			defer mutex.Unlock()

			if imageMetadata != nil {
				commitImageMetadata[plumbing.NewHash(commit)] = imageMetadata
			}

			return nil
		}); err != nil {
			return err
		}

		imageCommitImageMetadata[imageName] = commitImageMetadata
	}

	m.setImageCommitHashImageMetadata(imageCommitImageMetadata)

	return nil
}

func (m *imagesCleanupManager) getImageRepoImageList() map[string][]*image.Info {
	return *m.imageRepoImageList
}

func (m *imagesCleanupManager) setImageRepoImageList(repoImages map[string][]*image.Info) {
	m.imageRepoImageList = &repoImages
}

func (m *imagesCleanupManager) getImageCommitHashImageMetadata() map[string]map[plumbing.Hash]*storage.ImageMetadata {
	return *m.imageCommitHashImageMetadata
}

func (m *imagesCleanupManager) setImageCommitHashImageMetadata(imageCommitHashImageMetadata map[string]map[plumbing.Hash]*storage.ImageMetadata) {
	m.imageCommitHashImageMetadata = &imageCommitHashImageMetadata
}

func (m *imagesCleanupManager) updateImageCommitImageMetadata(imageName string, commitHashesToExclude ...plumbing.Hash) {
	commitImageMetadata, ok := (*m.imageCommitHashImageMetadata)[imageName]
	if !ok {
		return
	}

	resultImage := map[plumbing.Hash]*storage.ImageMetadata{}
outerLoop:
	for commitHash, imageMetadata := range commitImageMetadata {
		for _, commitHashToExclude := range commitHashesToExclude {
			if commitHash == commitHashToExclude {
				continue outerLoop
			}
		}

		resultImage[commitHash] = imageMetadata
	}

	(*m.imageCommitHashImageMetadata)[imageName] = resultImage
}

func (m *imagesCleanupManager) run(ctx context.Context) error {
	if err := logboek.Context(ctx).LogProcess("Fetching repo images data").DoError(func() error {
		return m.initRepoImagesData(ctx)
	}); err != nil {
		return err
	}

	repoImagesToCleanup := m.getImageRepoImageList()
	exceptedRepoImages := map[string][]*image.Info{}
	resultRepoImages := map[string][]*image.Info{}

	if m.LocalGit == nil {
		logboek.Context(ctx).Default().LogLnDetails("Images cleanup skipped due to local git repository was not detected")
		return nil
	}

	var err error

	if !m.WithoutKube {
		if err := logboek.Context(ctx).LogProcess("Skipping repo images that are being used in Kubernetes").DoError(func() error {
			repoImagesToCleanup, exceptedRepoImages, err = exceptRepoImagesByWhitelist(ctx, repoImagesToCleanup, m.KubernetesContextClients, m.KubernetesNamespaceRestrictionByContext)
			return err
		}); err != nil {
			return err
		}
	}

	if m.GitHistoryBasedCleanup || m.GitHistoryBasedCleanupV12 {
		resultRepoImages, err = m.repoImagesGitHistoryBasedCleanup(ctx, repoImagesToCleanup)
		if err != nil {
			return err
		}
	} else {
		resultRepoImages, err = m.repoImagesCleanup(ctx, repoImagesToCleanup)
		if err != nil {
			return err
		}
	}

	for imageName, repoImageList := range exceptedRepoImages {
		_, ok := resultRepoImages[imageName]
		if !ok {
			resultRepoImages[imageName] = repoImageList
		} else {
			resultRepoImages[imageName] = append(resultRepoImages[imageName], repoImageList...)
		}
	}

	m.setImageRepoImageList(resultRepoImages)

	return nil
}

func exceptRepoImagesByWhitelist(ctx context.Context, repoImages map[string][]*image.Info, kubernetesContextClients []*kube.ContextClient, kubernetesNamespaceRestrictionByContext map[string]string) (map[string][]*image.Info, map[string][]*image.Info, error) {
	deployedDockerImagesNames, err := getDeployedDockerImagesNames(ctx, kubernetesContextClients, kubernetesNamespaceRestrictionByContext)
	if err != nil {
		return nil, nil, err
	}

	exceptedRepoImages := map[string][]*image.Info{}
	for imageName, repoImageList := range repoImages {
		var newRepoImages []*image.Info

		logboek.Context(ctx).Default().LogBlock(logging.ImageLogProcessName(imageName, false)).Do(func() {

		Loop:
			for _, repoImage := range repoImageList {
				dockerImageName := fmt.Sprintf("%s:%s", repoImage.Repository, repoImage.Tag)
				for _, deployedDockerImageName := range deployedDockerImagesNames {
					if deployedDockerImageName == dockerImageName {
						exceptedImageList, ok := exceptedRepoImages[imageName]
						if !ok {
							exceptedImageList = []*image.Info{}
						}

						exceptedImageList = append(exceptedImageList, repoImage)
						exceptedRepoImages[imageName] = exceptedImageList

						logboek.Context(ctx).Default().LogFDetails("  tag: %s\n", repoImage.Tag)
						logboek.Context(ctx).LogOptionalLn()
						continue Loop
					}
				}

				newRepoImages = append(newRepoImages, repoImage)
			}

			repoImages[imageName] = newRepoImages
		})
	}

	return repoImages, exceptedRepoImages, nil
}

func getDeployedDockerImagesNames(ctx context.Context, kubernetesContextClients []*kube.ContextClient, kubernetesNamespaceRestrictionByContext map[string]string) ([]string, error) {
	var deployedDockerImagesNames []string
	for _, contextClient := range kubernetesContextClients {
		if err := logboek.Context(ctx).LogProcessInline("Getting deployed docker images (context %s)", contextClient.ContextName).
			DoError(func() error {
				kubernetesClientDeployedDockerImagesNames, err := deployedDockerImages(contextClient.Client, kubernetesNamespaceRestrictionByContext[contextClient.ContextName])
				if err != nil {
					return fmt.Errorf("cannot get deployed images: %s", err)
				}

				deployedDockerImagesNames = append(deployedDockerImagesNames, kubernetesClientDeployedDockerImagesNames...)

				return nil
			}); err != nil {
			return nil, err
		}
	}

	return deployedDockerImagesNames, nil
}

func (m *imagesCleanupManager) repoImagesCleanup(ctx context.Context, repoImagesToCleanup map[string][]*image.Info) (map[string][]*image.Info, error) {
	resultRepoImages := map[string][]*image.Info{}

	for imageName, repoImageListToCleanup := range repoImagesToCleanup {
		if err := logboek.Context(ctx).Default().LogProcess("Processing %s", logging.ImageLogProcessName(imageName, false)).
			Options(func(options types.LogProcessOptionsInterface) {
				options.Style(style.Highlight())
			}).
			DoError(func() error {
				repoImageListToCleanup, err := m.repoImagesCleanupByNonexistentGitPrimitive(ctx, repoImageListToCleanup)
				if err != nil {
					return err
				}

				resultRepoImageList, err := m.repoImagesCleanupByPolicies(ctx, repoImageListToCleanup)
				if err != nil {
					return err
				}
				resultRepoImages[imageName] = resultRepoImageList

				return nil
			}); err != nil {
			return nil, err
		}
	}

	return resultRepoImages, nil
}

func (m *imagesCleanupManager) repoImagesGitHistoryBasedCleanup(ctx context.Context, repoImagesToCleanup map[string][]*image.Info) (map[string][]*image.Info, error) {
	resultRepoImages := map[string][]*image.Info{}

	gitRepository, err := m.LocalGit.PlainOpen()
	if err != nil {
		return nil, fmt.Errorf("git plain open failed: %s", err)
	}

	var referencesToScan []*referenceToScan
	if err := logboek.Context(ctx).Default().LogProcess("Preparing references to scan").DoError(func() error {
		referencesToScan, err = getReferencesToScan(ctx, gitRepository, m.GitHistoryBasedCleanupOptions.KeepPolicies)
		return err
	}); err != nil {
		return nil, err
	}

	var imageContentSignatureRepoImageListToCleanup map[string]map[string][]*image.Info
	var imageContentSignatureExistingCommitHashes map[string]map[string][]plumbing.Hash
	if err := logboek.Context(ctx).Info().LogProcess("Grouping repo images tags by content signature").DoError(func() error {
		imageContentSignatureRepoImageListToCleanup, err = m.getImageContentSignatureRepoImageListToCleanup(repoImagesToCleanup)
		if err != nil {
			return err
		}

		imageContentSignatureExistingCommitHashes, err = m.getImageContentSignatureExistingCommitHashes(ctx)
		if err != nil {
			return err
		}

		if logboek.Context(ctx).Info().IsAccepted() {
			for imageName, contentSignatureRepoImageListToCleanup := range imageContentSignatureRepoImageListToCleanup {
				if len(contentSignatureRepoImageListToCleanup) == 0 {
					continue
				}

				logProcess := logboek.Context(ctx).Info().LogProcess(logging.ImageLogProcessName(imageName, false))
				logProcess.Start()

				var rows [][]interface{}
				for contentSignature, repoImageListToCleanup := range contentSignatureRepoImageListToCleanup {
					commitHashes := imageContentSignatureExistingCommitHashes[imageName][contentSignature]
					if len(commitHashes) == 0 || len(repoImageListToCleanup) == 0 {
						continue
					}

					var maxInd int
					for _, length := range []int{len(commitHashes), len(repoImageListToCleanup)} {
						if length > maxInd {
							maxInd = length
						}
					}

					shortify := func(column string) string {
						if len(column) > 15 {
							return fmt.Sprintf("%s..%s", column[:10], column[len(column)-3:])
						} else {
							return column
						}
					}

					for ind := 0; ind < maxInd; ind++ {
						var columns []interface{}
						if ind == 0 {
							columns = append(columns, shortify(contentSignature))
						} else {
							columns = append(columns, "")
						}

						if len(commitHashes) > ind {
							columns = append(columns, shortify(commitHashes[ind].String()))
						} else {
							columns = append(columns, "")
						}

						if len(repoImageListToCleanup) > ind {
							column := repoImageListToCleanup[ind].Tag
							if logboek.Context(ctx).Streams().ContentWidth() < 100 {
								column = shortify(column)
							}
							columns = append(columns, column)
						} else {
							columns = append(columns, "")
						}

						rows = append(rows, columns)
					}
				}

				if len(rows) != 0 {
					tbl := table.New("Content Signature", "Existing Commits", "Tags")
					tbl.WithWriter(logboek.Context(ctx).ProxyOutStream())
					tbl.WithHeaderFormatter(color.New(color.Underline).SprintfFunc())
					for _, row := range rows {
						tbl.AddRow(row...)
					}
					tbl.Print()

					logboek.Context(ctx).LogOptionalLn()
				}

				for contentSignature, repoImageListToCleanup := range contentSignatureRepoImageListToCleanup {
					commitHashes := imageContentSignatureExistingCommitHashes[imageName][contentSignature]
					if len(commitHashes) == 0 && len(repoImageListToCleanup) != 0 {
						logBlockMessage := fmt.Sprintf("Content signature %s is associated with non-existing commits. The following tags will be deleted", contentSignature)
						logboek.Context(ctx).Info().LogBlock(logBlockMessage).Do(func() {
							for _, repoImage := range repoImageListToCleanup {
								logboek.Context(ctx).Info().LogFDetails("  tag: %s\n", repoImage.Tag)
								logboek.Context(ctx).LogOptionalLn()
							}
						})
					}
				}

				logProcess.End()
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	if err = logboek.Context(ctx).Default().LogProcess("Processing images tags without related image metadata").DoError(func() error {
		imageRepoImageListWithoutRelatedContentSignature, err := m.getImageRepoImageListWithoutRelatedImageMetadata(repoImagesToCleanup, imageContentSignatureRepoImageListToCleanup)
		if err != nil {
			return err
		}

		for imageName, repoImages := range imageRepoImageListWithoutRelatedContentSignature {
			logProcess := logboek.Context(ctx).Default().LogProcess(logging.ImageLogProcessName(imageName, false))
			logProcess.Start()

			if !m.GitHistoryBasedCleanupV12 {
				if len(repoImages) != 0 {
					logboek.Context(ctx).Warn().LogF("Detected tags without related image metadata.\nThese tags will be saved during cleanup.\n")
					logboek.Context(ctx).Warn().LogF("Since v1.2 git history based cleanup will delete such tags by default.\nYou can force this behaviour in current werf version with --git-history-based-cleanup-v1.2 option.\n")
				}

				for _, repoImage := range repoImages {
					logboek.Context(ctx).Default().LogFDetails("  tag: %s\n", repoImage.Tag)
					logboek.Context(ctx).LogOptionalLn()
				}

				resultRepoImages[imageName] = append(resultRepoImages[imageName], repoImages...)
				repoImagesToCleanup[imageName] = exceptRepoImageList(repoImagesToCleanup[imageName], repoImages...)
			}

			logProcess.End()
		}

		return nil
	}); err != nil {
		return nil, err
	}

	if err := logboek.Context(ctx).Default().LogProcess("Git history based cleanup").
		Options(func(options types.LogProcessOptionsInterface) {
			options.Style(style.Highlight())
		}).
		DoError(func() error {
			for imageName, repoImageListToCleanup := range repoImagesToCleanup {
				var repoImageListToSave []*image.Info
				var commitHashesToCleanup []plumbing.Hash

				if err := logboek.Context(ctx).LogProcess(logging.ImageLogProcessName(imageName, false)).DoError(func() error {
					if err := logboek.Context(ctx).LogProcess("Scanning git references history").DoError(func() error {
						var commitHashes []plumbing.Hash
						contentSignatureCommitHashes := map[string][]plumbing.Hash{}
						contentSignatureRepoImageListToCleanup := imageContentSignatureRepoImageListToCleanup[imageName]
						for contentSignature, _ := range contentSignatureRepoImageListToCleanup {
							existingCommitHashes := imageContentSignatureExistingCommitHashes[imageName][contentSignature]
							if len(existingCommitHashes) == 0 {
								continue
							}

							contentSignatureCommitHashes[contentSignature] = existingCommitHashes
							commitHashes = append(commitHashes, existingCommitHashes...)
						}

						var repoImageListToKeep []*image.Info
						if len(contentSignatureCommitHashes) != 0 {
							reachedContentSignatureList, hitCommitHashes, err := scanReferencesHistory(ctx, gitRepository, referencesToScan, contentSignatureCommitHashes)
							if err != nil {
								return err
							}

							if len(hitCommitHashes) == 0 {
								commitHashesToCleanup = commitHashes
							} else {
							commitHashesLoop:
								for _, commitHash := range commitHashes {
									for _, hitCommitHash := range hitCommitHashes {
										if commitHash == hitCommitHash {
											continue commitHashesLoop
										}
									}

									commitHashesToCleanup = append(commitHashesToCleanup, commitHash)
								}
							}

							for _, contentSignature := range reachedContentSignatureList {
								contentSignatureRepoImageListToCleanup, ok := imageContentSignatureRepoImageListToCleanup[imageName][contentSignature]
								if !ok {
									panic("runtime error")
								}

								repoImageListToKeep = append(repoImageListToKeep, contentSignatureRepoImageListToCleanup...)
							}

							repoImageListToSave = append(repoImageListToSave, repoImageListToKeep...)
							resultRepoImages[imageName] = append(resultRepoImages[imageName], repoImageListToKeep...)
							repoImageListToCleanup = exceptRepoImageList(repoImageListToCleanup, repoImageListToKeep...)
						} else {
							logboek.Context(ctx).LogLn("Scanning stopped due to nothing to seek")
						}

						return nil
					}); err != nil {
						return err
					}

					if len(repoImageListToSave) != 0 {
						logboek.Context(ctx).Default().LogBlock("Saved tags").Do(func() {
							for _, repoImage := range repoImageListToSave {
								logboek.Context(ctx).Default().LogFDetails("  tag: %s\n", repoImage.Tag)
								logboek.Context(ctx).LogOptionalLn()
							}
						})
					}

					if len(repoImageListToCleanup) != 0 {
						if err := logboek.Context(ctx).Default().LogProcess("Deleting tags").DoError(func() error {
							return deleteRepoImageInImagesRepo(ctx, m.StorageManager, m.DryRun, repoImageListToCleanup...)
						}); err != nil {
							return err
						}
					}

					if len(commitHashesToCleanup) != 0 {
						logProcess := logboek.Context(ctx).Default().LogProcess("Cleaning up images metadata")
						logProcess.Start()

						if err := deleteMetaImagesInStagesStorage(ctx, m.StorageManager, m.ProjectName, imageName, m.DryRun, commitHashesToCleanup...); err != nil {
							logProcess.Fail()
							return err
						}

						m.updateImageCommitImageMetadata(imageName, commitHashesToCleanup...)

						logProcess.End()
					}

					return nil
				}); err != nil {
					return err
				}
			}

			return nil
		}); err != nil {
		return nil, err
	}

	if err := logboek.Context(ctx).Default().LogProcess("Deleting unused images metadata").DoError(func() error {
		imageUnusedCommitHashes, err := m.getImageUnusedCommitHashes(resultRepoImages)
		if err != nil {
			return err
		}

		for imageName, unusedCommitHashes := range imageUnusedCommitHashes {
			logProcess := logboek.Context(ctx).Default().LogProcess(logging.ImageLogProcessName(imageName, false))
			logProcess.Start()

			if err := deleteMetaImagesInStagesStorage(ctx, m.StorageManager, m.ProjectName, imageName, m.DryRun, unusedCommitHashes...); err != nil {
				logProcess.Fail()
				return err
			}

			m.updateImageCommitImageMetadata(imageName, unusedCommitHashes...)

			logProcess.End()
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return resultRepoImages, nil
}

// getImageContentSignatureRepoImageListToCleanup groups images, content signatures and repo images tags to clean up.
// The map has all content signatures and each value is related to repo images tags to clean up and can be empty.
// Repo images tags to clean up are all tags for particular werf.yaml image except for ones which are using in Kubernetes.
func (m *imagesCleanupManager) getImageContentSignatureRepoImageListToCleanup(repoImagesToCleanup map[string][]*image.Info) (map[string]map[string][]*image.Info, error) {
	imageContentSignatureRepoImageListToCleanup := map[string]map[string][]*image.Info{}

	imageCommitHashImageMetadata := m.getImageCommitHashImageMetadata()
	for imageName, repoImageListToCleanup := range repoImagesToCleanup {
		imageContentSignatureRepoImageListToCleanup[imageName] = map[string][]*image.Info{}

		for _, imageMetadata := range imageCommitHashImageMetadata[imageName] {
			_, ok := imageContentSignatureRepoImageListToCleanup[imageName][imageMetadata.ContentSignature]
			if ok {
				continue
			}

			var repoImageListToCleanupBySignature []*image.Info
			for _, repoImage := range repoImageListToCleanup {
				if repoImage.Labels[image.WerfContentSignatureLabel] == imageMetadata.ContentSignature {
					repoImageListToCleanupBySignature = append(repoImageListToCleanupBySignature, repoImage)
				}
			}

			if len(repoImageListToCleanupBySignature) != 0 {
				imageContentSignatureRepoImageListToCleanup[imageName][imageMetadata.ContentSignature] = repoImageListToCleanupBySignature
			}
		}
	}

	return imageContentSignatureRepoImageListToCleanup, nil
}

func (m *imagesCleanupManager) repoImagesCleanupByNonexistentGitPrimitive(ctx context.Context, repoImages []*image.Info) ([]*image.Info, error) {
	var nonexistentGitTagRepoImages, nonexistentGitCommitRepoImages, nonexistentGitBranchRepoImages []*image.Info

	var gitTags []string
	var gitBranches []string

	if m.LocalGit != nil {
		var err error
		gitTags, err = m.LocalGit.TagsList(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot get local git tags list: %s", err)
		}

		gitBranches, err = m.LocalGit.RemoteBranchesList(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot get local git branches list: %s", err)
		}
	}

Loop:
	for _, repoImage := range repoImages {
		strategy, ok := repoImage.Labels[image.WerfTagStrategyLabel]
		if !ok {
			continue
		}

		repoImageMetaTag, ok := repoImage.Labels[image.WerfImageTagLabel]
		if !ok {
			repoImageMetaTag = repoImage.Tag
		}

		switch strategy {
		case string(tag_strategy.GitTag):
			if repoImageMetaTagMatch(repoImageMetaTag, gitTags...) {
				continue Loop
			} else {
				nonexistentGitTagRepoImages = append(nonexistentGitTagRepoImages, repoImage)
			}
		case string(tag_strategy.GitBranch):
			if repoImageMetaTagMatch(repoImageMetaTag, gitBranches...) {
				continue Loop
			} else {
				nonexistentGitBranchRepoImages = append(nonexistentGitBranchRepoImages, repoImage)
			}
		case string(tag_strategy.GitCommit):
			exist := false

			if m.LocalGit != nil {
				var err error

				exist, err = m.LocalGit.IsCommitExists(ctx, repoImageMetaTag)
				if err != nil {
					if strings.HasPrefix(err.Error(), "bad commit hash") {
						exist = false
					} else {
						return nil, err
					}
				}
			}

			if !exist {
				nonexistentGitCommitRepoImages = append(nonexistentGitCommitRepoImages, repoImage)
			}
		}
	}

	if len(nonexistentGitTagRepoImages) != 0 {
		if err := logboek.Context(ctx).Default().LogBlock("Removed tags by nonexistent git-tag policy").DoError(func() error {
			return deleteRepoImageInImagesRepo(ctx, m.StorageManager, m.DryRun, nonexistentGitTagRepoImages...)
		}); err != nil {
			return nil, err
		}

		repoImages = exceptRepoImageList(repoImages, nonexistentGitTagRepoImages...)
	}

	if len(nonexistentGitBranchRepoImages) != 0 {
		if err := logboek.Context(ctx).Default().LogBlock("Removed tags by nonexistent git-branch policy").DoError(func() error {
			return deleteRepoImageInImagesRepo(ctx, m.StorageManager, m.DryRun, nonexistentGitBranchRepoImages...)
		}); err != nil {
			return nil, err
		}

		repoImages = exceptRepoImageList(repoImages, nonexistentGitBranchRepoImages...)
	}

	if len(nonexistentGitCommitRepoImages) != 0 {
		if err := logboek.Context(ctx).Default().LogBlock("Removed tags by nonexistent git-commit policy").DoError(func() error {
			return deleteRepoImageInImagesRepo(ctx, m.StorageManager, m.DryRun, nonexistentGitCommitRepoImages...)
		}); err != nil {
			return nil, err
		}

		repoImages = exceptRepoImageList(repoImages, nonexistentGitCommitRepoImages...)
	}

	return repoImages, nil
}

func repoImageMetaTagMatch(imageMetaTag string, matches ...string) bool {
	for _, match := range matches {
		if imageMetaTag == slug.DockerTag(match) {
			return true
		}
	}

	return false
}

func (m *imagesCleanupManager) repoImagesCleanupByPolicies(ctx context.Context, repoImages []*image.Info) ([]*image.Info, error) {
	var repoImagesWithGitTagScheme, repoImagesWithGitCommitScheme, repoImagesWithStagesSignatureScheme []*image.Info

	for _, repoImage := range repoImages {
		strategy, ok := repoImage.Labels[image.WerfTagStrategyLabel]
		if !ok {
			continue
		}

		switch strategy {
		case string(tag_strategy.GitTag):
			repoImagesWithGitTagScheme = append(repoImagesWithGitTagScheme, repoImage)
		case string(tag_strategy.GitCommit):
			repoImagesWithGitCommitScheme = append(repoImagesWithGitCommitScheme, repoImage)
		case string(tag_strategy.StagesSignature):
			repoImagesWithStagesSignatureScheme = append(repoImagesWithStagesSignatureScheme, repoImage)
		}
	}

	cleanupByPolicyOptions := repoImagesCleanupByPolicyOptions{
		hasLimit:        m.Policies.GitTagStrategyHasLimit,
		limit:           m.Policies.GitTagStrategyLimit,
		hasExpiryPeriod: m.Policies.GitTagStrategyHasExpiryPeriod,
		expiryPeriod:    m.Policies.GitTagStrategyExpiryPeriod,
		schemeName:      string(tag_strategy.GitTag),
	}

	var err error
	repoImages, err = m.repoImagesCleanupByPolicy(ctx, repoImages, repoImagesWithGitTagScheme, cleanupByPolicyOptions)
	if err != nil {
		return nil, err
	}

	cleanupByPolicyOptions = repoImagesCleanupByPolicyOptions{
		hasLimit:        m.Policies.GitCommitStrategyHasLimit,
		limit:           m.Policies.GitCommitStrategyLimit,
		hasExpiryPeriod: m.Policies.GitCommitStrategyHasExpiryPeriod,
		expiryPeriod:    m.Policies.GitCommitStrategyExpiryPeriod,
		schemeName:      string(tag_strategy.GitCommit),
	}

	repoImages, err = m.repoImagesCleanupByPolicy(ctx, repoImages, repoImagesWithGitCommitScheme, cleanupByPolicyOptions)
	if err != nil {
		return nil, err
	}

	cleanupByPolicyOptions = repoImagesCleanupByPolicyOptions{
		hasLimit:        m.Policies.StagesSignatureStrategyHasLimit,
		limit:           m.Policies.StagesSignatureStrategyLimit,
		hasExpiryPeriod: m.Policies.StagesSignatureStrategyHasExpiryPeriod,
		expiryPeriod:    m.Policies.StagesSignatureStrategyExpiryPeriod,
		schemeName:      string(tag_strategy.StagesSignature),
	}

	repoImages, err = m.repoImagesCleanupByPolicy(ctx, repoImages, repoImagesWithStagesSignatureScheme, cleanupByPolicyOptions)
	if err != nil {
		return nil, err
	}

	return repoImages, nil
}

type repoImagesCleanupByPolicyOptions struct {
	hasLimit        bool
	limit           int64
	hasExpiryPeriod bool
	expiryPeriod    time.Duration
	schemeName      string
}

func (m *imagesCleanupManager) repoImagesCleanupByPolicy(ctx context.Context, repoImages, repoImagesWithScheme []*image.Info, options repoImagesCleanupByPolicyOptions) ([]*image.Info, error) {
	var expiryTime time.Time
	if options.hasExpiryPeriod {
		expiryTime = time.Now().Add(-options.expiryPeriod)
	}

	sort.Slice(repoImagesWithScheme, func(i, j int) bool {
		iCreated := repoImagesWithScheme[i].GetCreatedAt()
		jCreated := repoImagesWithScheme[j].GetCreatedAt()
		return iCreated.Before(jCreated)
	})

	var notExpiredRepoImages, expiredRepoImages []*image.Info
	for _, repoImage := range repoImagesWithScheme {
		if options.hasExpiryPeriod && repoImage.GetCreatedAt().Before(expiryTime) {
			expiredRepoImages = append(expiredRepoImages, repoImage)
		} else {
			notExpiredRepoImages = append(notExpiredRepoImages, repoImage)
		}
	}

	if len(expiredRepoImages) != 0 {
		logBlockMessage := fmt.Sprintf("Removed tags by %s date policy (created before %s)", options.schemeName, expiryTime.Format("2006-01-02T15:04:05-0700"))
		if err := logboek.Context(ctx).Default().LogBlock(logBlockMessage).DoError(func() error {
			return deleteRepoImageInImagesRepo(ctx, m.StorageManager, m.DryRun, expiredRepoImages...)
		}); err != nil {
			return nil, err
		}

		repoImages = exceptRepoImageList(repoImages, expiredRepoImages...)
	}

	if options.hasLimit && int64(len(notExpiredRepoImages)) > options.limit {
		excessImagesByLimit := notExpiredRepoImages[:int64(len(notExpiredRepoImages))-options.limit]

		logBlockMessage := fmt.Sprintf("Removed tags by %s limit policy (> %d)", options.schemeName, options.limit)
		if err := logboek.Context(ctx).Default().LogBlock(logBlockMessage).DoError(func() error {
			return deleteRepoImageInImagesRepo(ctx, m.StorageManager, m.DryRun, excessImagesByLimit...)
		},
		); err != nil {
			return nil, err
		}

		repoImages = exceptRepoImageList(repoImages, excessImagesByLimit...)
	}

	return repoImages, nil
}

func (m *imagesCleanupManager) getImageUnusedCommitHashes(resultImageRepoImageList map[string][]*image.Info) (map[string][]plumbing.Hash, error) {
	unusedImageCommitHashes := map[string][]plumbing.Hash{}

	for imageName, commitHashImageMetadata := range m.getImageCommitHashImageMetadata() {
		var unusedCommitHashes []plumbing.Hash
		repoImageList, ok := resultImageRepoImageList[imageName]
		if !ok {
			repoImageList = []*image.Info{}
		}

	outerLoop:
		for commitHash, imageMetadata := range commitHashImageMetadata {
			for _, repoImage := range repoImageList {
				if repoImage.Labels[image.WerfContentSignatureLabel] == imageMetadata.ContentSignature {
					continue outerLoop
				}
			}

			unusedCommitHashes = append(unusedCommitHashes, commitHash)
		}

		unusedImageCommitHashes[imageName] = unusedCommitHashes
	}

	return unusedImageCommitHashes, nil
}

func (m *imagesCleanupManager) getImageRepoImageListWithoutRelatedImageMetadata(imageRepoImageListToCleanup map[string][]*image.Info, imageContentSignatureRepoImageListToCleanup map[string]map[string][]*image.Info) (map[string][]*image.Info, error) {
	imageRepoImageListWithoutRelatedCommit := map[string][]*image.Info{}

	for imageName, repoImageListToCleanup := range imageRepoImageListToCleanup {
		unusedRepoImageList := repoImageListToCleanup

		contentSignatureRepoImageListToCleanup, ok := imageContentSignatureRepoImageListToCleanup[imageName]
		if !ok {
			contentSignatureRepoImageListToCleanup = map[string][]*image.Info{}
		}

		for _, filteredRepoImageListToCleanup := range contentSignatureRepoImageListToCleanup {
			unusedRepoImageList = exceptRepoImageList(unusedRepoImageList, filteredRepoImageListToCleanup...)
		}

		imageRepoImageListWithoutRelatedCommit[imageName] = unusedRepoImageList
	}

	return imageRepoImageListWithoutRelatedCommit, nil
}

// getImageContentSignatureExistingCommitHashes groups images, content signatures and commit hashes which exist in the git repo.
func (m *imagesCleanupManager) getImageContentSignatureExistingCommitHashes(ctx context.Context) (map[string]map[string][]plumbing.Hash, error) {
	imageContentSignatureCommitHashes := map[string]map[string][]plumbing.Hash{}

	for _, imageName := range m.ImageNameList {
		imageContentSignatureCommitHashes[imageName] = map[string][]plumbing.Hash{}

		for commitHash, imageMetadata := range m.getImageCommitHashImageMetadata()[imageName] {
			commitHashes, ok := imageContentSignatureCommitHashes[imageName][imageMetadata.ContentSignature]
			if !ok {
				commitHashes = []plumbing.Hash{}
			}

			exist, err := m.LocalGit.IsCommitExists(ctx, commitHash.String())
			if err != nil {
				return nil, fmt.Errorf("check git commit existence failed: %s", err)
			}

			if exist {
				commitHashes = append(commitHashes, commitHash)
				imageContentSignatureCommitHashes[imageName][imageMetadata.ContentSignature] = commitHashes
			}
		}
	}

	return imageContentSignatureCommitHashes, nil
}

func deployedDockerImages(kubernetesClient kubernetes.Interface, kubernetesNamespace string) ([]string, error) {
	var deployedDockerImages []string

	images, err := getPodsImages(kubernetesClient, kubernetesNamespace)
	if err != nil {
		return nil, fmt.Errorf("cannot get Pods images: %s", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	images, err = getReplicationControllersImages(kubernetesClient, kubernetesNamespace)
	if err != nil {
		return nil, fmt.Errorf("cannot get ReplicationControllers images: %s", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	images, err = getDeploymentsImages(kubernetesClient, kubernetesNamespace)
	if err != nil {
		return nil, fmt.Errorf("cannot get Deployments images: %s", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	images, err = getStatefulSetsImages(kubernetesClient, kubernetesNamespace)
	if err != nil {
		return nil, fmt.Errorf("cannot get StatefulSets images: %s", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	images, err = getDaemonSetsImages(kubernetesClient, kubernetesNamespace)
	if err != nil {
		return nil, fmt.Errorf("cannot get DaemonSets images: %s", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	images, err = getReplicaSetsImages(kubernetesClient, kubernetesNamespace)
	if err != nil {
		return nil, fmt.Errorf("cannot get ReplicaSets images: %s", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	images, err = getCronJobsImages(kubernetesClient, kubernetesNamespace)
	if err != nil {
		return nil, fmt.Errorf("cannot get CronJobs images: %s", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	images, err = getJobsImages(kubernetesClient, kubernetesNamespace)
	if err != nil {
		return nil, fmt.Errorf("cannot get Jobs images: %s", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	return deployedDockerImages, nil
}

func getPodsImages(kubernetesClient kubernetes.Interface, kubernetesNamespace string) ([]string, error) {
	var images []string
	list, err := kubernetesClient.CoreV1().Pods(kubernetesNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, pod := range list.Items {
		for _, container := range pod.Spec.Containers {
			images = append(images, container.Image)
		}
	}

	return images, nil
}

func getReplicationControllersImages(kubernetesClient kubernetes.Interface, kubernetesNamespace string) ([]string, error) {
	var images []string
	list, err := kubernetesClient.CoreV1().ReplicationControllers(kubernetesNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, replicationController := range list.Items {
		for _, container := range replicationController.Spec.Template.Spec.Containers {
			images = append(images, container.Image)
		}
	}

	return images, nil
}

func getDeploymentsImages(kubernetesClient kubernetes.Interface, kubernetesNamespace string) ([]string, error) {
	var images []string
	list, err := kubernetesClient.AppsV1().Deployments(kubernetesNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, deployment := range list.Items {
		for _, container := range deployment.Spec.Template.Spec.Containers {
			images = append(images, container.Image)
		}
	}

	return images, nil
}

func getStatefulSetsImages(kubernetesClient kubernetes.Interface, kubernetesNamespace string) ([]string, error) {
	var images []string
	list, err := kubernetesClient.AppsV1().StatefulSets(kubernetesNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, statefulSet := range list.Items {
		for _, container := range statefulSet.Spec.Template.Spec.Containers {
			images = append(images, container.Image)
		}
	}

	return images, nil
}

func getDaemonSetsImages(kubernetesClient kubernetes.Interface, kubernetesNamespace string) ([]string, error) {
	var images []string
	list, err := kubernetesClient.AppsV1().DaemonSets(kubernetesNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, daemonSets := range list.Items {
		for _, container := range daemonSets.Spec.Template.Spec.Containers {
			images = append(images, container.Image)
		}
	}

	return images, nil
}

func getReplicaSetsImages(kubernetesClient kubernetes.Interface, kubernetesNamespace string) ([]string, error) {
	var images []string
	list, err := kubernetesClient.AppsV1().ReplicaSets(kubernetesNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, replicaSet := range list.Items {
		for _, container := range replicaSet.Spec.Template.Spec.Containers {
			images = append(images, container.Image)
		}
	}

	return images, nil
}

func getCronJobsImages(kubernetesClient kubernetes.Interface, kubernetesNamespace string) ([]string, error) {
	var images []string
	list, err := kubernetesClient.BatchV1beta1().CronJobs(kubernetesNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, cronJob := range list.Items {
		for _, container := range cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers {
			images = append(images, container.Image)
		}
	}

	return images, nil
}

func getJobsImages(kubernetesClient kubernetes.Interface, kubernetesNamespace string) ([]string, error) {
	var images []string
	list, err := kubernetesClient.BatchV1().Jobs(kubernetesNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, job := range list.Items {
		for _, container := range job.Spec.Template.Spec.Containers {
			images = append(images, container.Image)
		}
	}

	return images, nil
}

func deleteMetaImagesInStagesStorage(ctx context.Context, storageManager *manager.StorageManager, projectName, imageName string, dryRun bool, commitHashes ...plumbing.Hash) error {
	var commits []string
	for _, commitHash := range commitHashes {
		commits = append(commits, commitHash.String())
	}

	if dryRun {
		for _, commit := range commits {
			logboek.Context(ctx).Info().LogLn(commit)
		}
		return nil
	}

	return storageManager.ForEachRmImageCommit(ctx, projectName, imageName, commits, func(commit string, err error) error {
		if err != nil {
			logboek.Context(ctx).Warn().LogF(
				"WARNING: Metadata image deletion (image %s, commit: %s) failed: %s\n",
				logging.ImageLogName(imageName, false), commit, err,
			)
			logboek.Context(ctx).LogOptionalLn()
		}

		return nil
	})
}
