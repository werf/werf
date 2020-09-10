package cleaning

import (
	"context"
	"fmt"
	"time"

	"github.com/werf/kubedog/pkg/kube"

	"github.com/fatih/color"
	"github.com/rodaine/table"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"

	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/logging"
	"github.com/werf/werf/pkg/stages_manager"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/werf"
)

type ImagesCleanupOptions struct {
	ImageNameList                           []string
	LocalGit                                GitRepo
	KubernetesContextClients                []*kube.ContextClient
	KubernetesNamespaceRestrictionByContext map[string]string
	WithoutKube                             bool
	GitHistoryBasedCleanupOptions           config.MetaCleanup
	DryRun                                  bool
}

func ImagesCleanup(ctx context.Context, projectName string, imagesRepo storage.ImagesRepo, stagesManager *stages_manager.StagesManager, storageLockManager storage.LockManager, options ImagesCleanupOptions) error {
	m := newImagesCleanupManager(projectName, imagesRepo, stagesManager, options)

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

func newImagesCleanupManager(projectName string, imagesRepo storage.ImagesRepo, stagesManager *stages_manager.StagesManager, options ImagesCleanupOptions) *imagesCleanupManager {
	return &imagesCleanupManager{
		ProjectName:                             projectName,
		ImagesRepo:                              imagesRepo,
		StagesManager:                           stagesManager,
		ImageNameList:                           options.ImageNameList,
		DryRun:                                  options.DryRun,
		LocalGit:                                options.LocalGit,
		KubernetesContextClients:                options.KubernetesContextClients,
		KubernetesNamespaceRestrictionByContext: options.KubernetesNamespaceRestrictionByContext,
		WithoutKube:                             options.WithoutKube,
		GitHistoryBasedCleanupOptions:           options.GitHistoryBasedCleanupOptions,
	}
}

type imagesCleanupManager struct {
	imageRepoImageList           *map[string][]*image.Info
	imageCommitHashImageMetadata *map[string]map[plumbing.Hash]*storage.ImageMetadata

	ProjectName                             string
	ImagesRepo                              storage.ImagesRepo
	StagesManager                           *stages_manager.StagesManager
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

func (m *imagesCleanupManager) initRepoImagesData(ctx context.Context) error {
	if err := logboek.Context(ctx).Info().LogProcess("Fetching repo images").DoError(func() error {
		return m.initRepoImages(ctx)
	}); err != nil {
		return err
	}

	if err := logboek.Context(ctx).Info().LogProcess("Fetching images metadata").DoError(func() error {
		return m.initImageCommitHashImageMetadata(ctx)
	}); err != nil {
		return err
	}

	return nil
}

func (m *imagesCleanupManager) initRepoImages(ctx context.Context) error {
	repoImages, err := selectRepoImagesFromImagesRepo(ctx, m.ImagesRepo, m.ImageNameList)
	if err != nil {
		return err
	}

	m.setImageRepoImageList(repoImages)

	return nil
}

func (m *imagesCleanupManager) initImageCommitHashImageMetadata(ctx context.Context) error {
	imageCommitImageMetadata := map[string]map[plumbing.Hash]*storage.ImageMetadata{}
	for _, imageName := range m.ImageNameList {
		commits, err := m.StagesManager.StagesStorage.GetImageCommits(ctx, m.ProjectName, imageName)
		if err != nil {
			return fmt.Errorf("get image %s commits failed: %s", imageName, err)
		}

		commitImageMetadata := map[plumbing.Hash]*storage.ImageMetadata{}
		for _, commit := range commits {
			imageMetadata, err := m.StagesManager.StagesStorage.GetImageMetadataByCommit(ctx, m.ProjectName, imageName, commit)
			if err != nil {
				return fmt.Errorf("get image %s metadata by commit %s failed", imageName, commit)
			}

			if imageMetadata != nil {
				commitImageMetadata[plumbing.NewHash(commit)] = imageMetadata
			}
		}

		imageCommitImageMetadata[imageName] = commitImageMetadata
	}

	m.setImageCommitImageMetadata(imageCommitImageMetadata)

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

func (m *imagesCleanupManager) setImageCommitImageMetadata(imageCommitImageMetadata map[string]map[plumbing.Hash]*storage.ImageMetadata) {
	m.imageCommitHashImageMetadata = &imageCommitImageMetadata
}

func (m *imagesCleanupManager) run(ctx context.Context) error {
	imagesCleanupLockName := fmt.Sprintf("images-cleanup.%s", m.ImagesRepo.String())
	return werf.WithHostLock(ctx, imagesCleanupLockName, lockgate.AcquireOptions{Timeout: time.Second * 600}, func() error {
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

		resultRepoImages, err = m.repoImagesGitHistoryBasedCleanup(ctx, repoImagesToCleanup)
		if err != nil {
			return err
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
	})
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
					return fmt.Errorf("cannot get deployed imagesRepoImageList: %s", err)
				}

				deployedDockerImagesNames = append(deployedDockerImagesNames, kubernetesClientDeployedDockerImagesNames...)

				return nil
			}); err != nil {
			return nil, err
		}
	}

	return deployedDockerImagesNames, nil
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

	if err := logboek.Context(ctx).Default().LogProcess("Git history based cleanup").
		Options(func(options types.LogProcessOptionsInterface) {
			options.Style(style.Highlight())
		}).
		DoError(func() error {
			for imageName, repoImageListToCleanup := range repoImagesToCleanup {
				var repoImageListToSave []*image.Info
				if err := logboek.Context(ctx).LogProcess(logging.ImageLogProcessName(imageName, false)).DoError(func() error {
					if err := logboek.Context(ctx).LogProcess("Scanning git references history").DoError(func() error {
						contentSignatureCommitHashes := map[string][]plumbing.Hash{}
						contentSignatureRepoImageListToCleanup := imageContentSignatureRepoImageListToCleanup[imageName]
						for contentSignature, _ := range contentSignatureRepoImageListToCleanup {
							existingCommitHashes := imageContentSignatureExistingCommitHashes[imageName][contentSignature]
							if len(existingCommitHashes) == 0 {
								continue
							}

							contentSignatureCommitHashes[contentSignature] = existingCommitHashes
						}

						var repoImageListToKeep []*image.Info
						if len(contentSignatureCommitHashes) != 0 {
							reachedContentSignatureList, err := scanReferencesHistory(ctx, gitRepository, referencesToScan, contentSignatureCommitHashes)
							if err != nil {
								return err
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

					if err := logboek.Context(ctx).Default().LogProcess("Deleting tags").DoError(func() error {
						return deleteRepoImageInImagesRepo(ctx, m.ImagesRepo, m.DryRun, repoImageListToCleanup...)
					}); err != nil {
						return err
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

		for imageName, commitHashes := range imageUnusedCommitHashes {
			logProcess := logboek.Context(ctx).Default().LogProcess(logging.ImageLogProcessName(imageName, false))
			logProcess.Start()

			for _, commitHash := range commitHashes {
				if m.DryRun {
					logboek.Context(ctx).Default().LogLn(commitHash)
				} else {
					if err := m.StagesManager.StagesStorage.RmImageCommit(ctx, m.ProjectName, imageName, commitHash.String()); err != nil {
						logboek.Context(ctx).Warn().LogF(
							"WARNING: Metadata image deletion (image %s, commit: %s) failed: %s\n",
							logging.ImageLogName(imageName, false), commitHash.String(), err,
						)
						logboek.Context(ctx).LogOptionalLn()
					}
				}
			}

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

			imageContentSignatureRepoImageListToCleanup[imageName][imageMetadata.ContentSignature] = repoImageListToCleanupBySignature
		}
	}

	return imageContentSignatureRepoImageListToCleanup, nil
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
