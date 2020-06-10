package cleaning

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/flant/lockgate"
	"github.com/werf/werf/pkg/werf"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/flant/logboek"

	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/logging"
	"github.com/werf/werf/pkg/slug"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/tag_strategy"
)

type ImagesCleanupOptions struct {
	ImageNameList             []string
	LocalGit                  GitRepo
	KubernetesContextsClients map[string]kubernetes.Interface
	WithoutKube               bool
	Policies                  ImagesCleanupPolicies
	DryRun                    bool
}

func ImagesCleanup(projectName string, imagesRepo storage.ImagesRepo, storageLockManager storage.LockManager, options ImagesCleanupOptions) error {
	m := newImagesCleanupManager(imagesRepo, options)

	if lock, err := storageLockManager.LockStagesAndImages(projectName, storage.LockStagesAndImagesOptions{GetOrCreateImagesOnly: false}); err != nil {
		return fmt.Errorf("unable to lock stages and images: %s", err)
	} else {
		defer storageLockManager.Unlock(lock)
	}

	return logboek.Default.LogProcess(
		"Running images cleanup",
		logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()},
		m.run,
	)
}

func newImagesCleanupManager(imagesRepo storage.ImagesRepo, options ImagesCleanupOptions) *imagesCleanupManager {
	return &imagesCleanupManager{
		ImagesRepo:                imagesRepo,
		ImageNameList:             options.ImageNameList,
		DryRun:                    options.DryRun,
		LocalGit:                  options.LocalGit,
		KubernetesContextsClients: options.KubernetesContextsClients,
		WithoutKube:               options.WithoutKube,
		Policies:                  options.Policies,
	}
}

type imagesCleanupManager struct {
	imagesRepoImages *map[string][]*image.Info

	ImagesRepo                storage.ImagesRepo
	ImageNameList             []string
	LocalGit                  GitRepo
	KubernetesContextsClients map[string]kubernetes.Interface
	WithoutKube               bool
	Policies                  ImagesCleanupPolicies
	DryRun                    bool
}

type GitRepo interface {
	IsCommitExists(commit string) (bool, error)
	TagsList() ([]string, error)
	RemoteBranchesList() ([]string, error)
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

func (m *imagesCleanupManager) initRepoImages() error {
	repoImages, err := selectRepoImagesFromImagesRepo(m.ImagesRepo, m.ImageNameList)
	if err != nil {
		return err
	}

	m.setImagesRepoImages(repoImages)

	return nil
}

func (m *imagesCleanupManager) getImagesRepoImages() map[string][]*image.Info {
	return *m.imagesRepoImages
}

func (m *imagesCleanupManager) setImagesRepoImages(repoImages map[string][]*image.Info) {
	m.imagesRepoImages = &repoImages
}

func (m *imagesCleanupManager) run() error {
	imagesCleanupLockName := fmt.Sprintf("images-cleanup.%s", m.ImagesRepo.String())
	return werf.WithHostLock(imagesCleanupLockName, lockgate.AcquireOptions{Timeout: time.Second * 600}, func() error {
		if err := m.initRepoImages(); err != nil {
			return err
		}

		repoImagesToCleanup := m.getImagesRepoImages()
		resultRepoImages := map[string][]*image.Info{}

		var err error
		if m.LocalGit != nil {
			if !m.WithoutKube {
				if err := logboek.LogProcess("Skipping repo images that are being used in Kubernetes", logboek.LogProcessOptions{}, func() error {
					repoImagesToCleanup, resultRepoImages, err = exceptRepoImagesByWhitelist(repoImagesToCleanup, m.KubernetesContextsClients)
					return err
				}); err != nil {
					return err
				}
			}

			for imageName, repoImageListToCleanup := range repoImagesToCleanup {
				logProcessMessage := fmt.Sprintf("Processing image %s", logging.ImageLogName(imageName, false))
				if err := logboek.Default.LogProcess(
					logProcessMessage,
					logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()},
					func() error {
						repoImageListToCleanup, err = m.repoImagesCleanupByNonexistentGitPrimitive(repoImageListToCleanup)
						if err != nil {
							return err
						}

						resultRepoImageList, err := m.repoImagesCleanupByPolicies(repoImageListToCleanup)
						if err != nil {
							return err
						}

						_, ok := resultRepoImages[imageName]
						if !ok {
							resultRepoImages[imageName] = resultRepoImageList
						} else {
							resultRepoImages[imageName] = append(resultRepoImages[imageName], resultRepoImageList...)
						}

						return nil
					},
				); err != nil {
					return err
				}
			}
		}

		m.setImagesRepoImages(resultRepoImages)

		return nil
	})
}

func exceptRepoImagesByWhitelist(repoImages map[string][]*image.Info, kubernetesContextsClients map[string]kubernetes.Interface) (map[string][]*image.Info, map[string][]*image.Info, error) {
	var deployedDockerImagesNames []string
	for contextName, kubernetesClient := range kubernetesContextsClients {
		if err := logboek.LogProcessInline(fmt.Sprintf("Getting deployed docker images (context %s)", contextName), logboek.LogProcessInlineOptions{}, func() error {
			kubernetesClientDeployedDockerImagesNames, err := deployedDockerImages(kubernetesClient)
			if err != nil {
				return fmt.Errorf("cannot get deployed imagesRepoImageList: %s", err)
			}

			deployedDockerImagesNames = append(deployedDockerImagesNames, kubernetesClientDeployedDockerImagesNames...)

			return nil
		}); err != nil {
			return nil, nil, err
		}
	}

	exceptedRepoImages := map[string][]*image.Info{}
	for imageName, repoImageList := range repoImages {
		var newRepoImages []*image.Info

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

					logboek.Default.LogLnDetails(dockerImageName)
					continue Loop
				}
			}

			newRepoImages = append(newRepoImages, repoImage)
		}

		repoImages[imageName] = newRepoImages
	}

	return repoImages, exceptedRepoImages, nil
}

func (m *imagesCleanupManager) repoImagesCleanupByNonexistentGitPrimitive(repoImages []*image.Info) ([]*image.Info, error) {
	var nonexistentGitTagRepoImages, nonexistentGitCommitRepoImages, nonexistentGitBranchRepoImages []*image.Info

	var gitTags []string
	var gitBranches []string

	if m.LocalGit != nil {
		var err error
		gitTags, err = m.LocalGit.TagsList()
		if err != nil {
			return nil, fmt.Errorf("cannot get local git tags list: %s", err)
		}

		gitBranches, err = m.LocalGit.RemoteBranchesList()
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

				exist, err = m.LocalGit.IsCommitExists(repoImageMetaTag)
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
		if err := logboek.Default.LogBlock(
			"Removed tags by nonexistent git-tag policy",
			logboek.LevelLogBlockOptions{},
			func() error {
				return deleteRepoImageInImagesRepo(m.ImagesRepo, m.DryRun, nonexistentGitTagRepoImages...)
			},
		); err != nil {
			return nil, err
		}

		repoImages = exceptRepoImageList(repoImages, nonexistentGitTagRepoImages...)
	}

	if len(nonexistentGitBranchRepoImages) != 0 {
		if err := logboek.Default.LogBlock(
			"Removed tags by nonexistent git-branch policy",
			logboek.LevelLogBlockOptions{},
			func() error {
				return deleteRepoImageInImagesRepo(m.ImagesRepo, m.DryRun, nonexistentGitBranchRepoImages...)
			},
		); err != nil {
			return nil, err
		}

		repoImages = exceptRepoImageList(repoImages, nonexistentGitBranchRepoImages...)
	}

	if len(nonexistentGitCommitRepoImages) != 0 {
		if err := logboek.Default.LogBlock(
			"Removed tags by nonexistent git-commit policy",
			logboek.LevelLogBlockOptions{},
			func() error {
				return deleteRepoImageInImagesRepo(m.ImagesRepo, m.DryRun, nonexistentGitCommitRepoImages...)
			},
		); err != nil {
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

func (m *imagesCleanupManager) repoImagesCleanupByPolicies(repoImages []*image.Info) ([]*image.Info, error) {
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
	repoImages, err = m.repoImagesCleanupByPolicy(repoImages, repoImagesWithGitTagScheme, cleanupByPolicyOptions)
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

	repoImages, err = m.repoImagesCleanupByPolicy(repoImages, repoImagesWithGitCommitScheme, cleanupByPolicyOptions)
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

	repoImages, err = m.repoImagesCleanupByPolicy(repoImages, repoImagesWithStagesSignatureScheme, cleanupByPolicyOptions)
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

func (m *imagesCleanupManager) repoImagesCleanupByPolicy(repoImages, repoImagesWithScheme []*image.Info, options repoImagesCleanupByPolicyOptions) ([]*image.Info, error) {
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
		if err := logboek.Default.LogBlock(
			logBlockMessage,
			logboek.LevelLogBlockOptions{},
			func() error {
				return deleteRepoImageInImagesRepo(m.ImagesRepo, m.DryRun, expiredRepoImages...)
			},
		); err != nil {
			return nil, err
		}

		repoImages = exceptRepoImageList(repoImages, expiredRepoImages...)
	}

	if options.hasLimit && int64(len(notExpiredRepoImages)) > options.limit {
		excessImagesByLimit := notExpiredRepoImages[:int64(len(notExpiredRepoImages))-options.limit]

		logBlockMessage := fmt.Sprintf("Removed tags by %s limit policy (> %d)", options.schemeName, options.limit)
		if err := logboek.Default.LogBlock(
			logBlockMessage,
			logboek.LevelLogBlockOptions{},
			func() error {
				return deleteRepoImageInImagesRepo(m.ImagesRepo, m.DryRun, excessImagesByLimit...)
			},
		); err != nil {
			return nil, err
		}

		repoImages = exceptRepoImageList(repoImages, excessImagesByLimit...)
	}

	return repoImages, nil
}

func deployedDockerImages(kubernetesClient kubernetes.Interface) ([]string, error) {
	var deployedDockerImages []string

	images, err := getPodsImages(kubernetesClient)
	if err != nil {
		return nil, fmt.Errorf("cannot get Pods images: %s", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	images, err = getReplicationControllersImages(kubernetesClient)
	if err != nil {
		return nil, fmt.Errorf("cannot get ReplicationControllers images: %s", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	images, err = getDeploymentsImages(kubernetesClient)
	if err != nil {
		return nil, fmt.Errorf("cannot get Deployments images: %s", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	images, err = getStatefulSetsImages(kubernetesClient)
	if err != nil {
		return nil, fmt.Errorf("cannot get StatefulSets images: %s", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	images, err = getDaemonSetsImages(kubernetesClient)
	if err != nil {
		return nil, fmt.Errorf("cannot get DaemonSets images: %s", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	images, err = getReplicaSetsImages(kubernetesClient)
	if err != nil {
		return nil, fmt.Errorf("cannot get ReplicaSets images: %s", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	images, err = getCronJobsImages(kubernetesClient)
	if err != nil {
		return nil, fmt.Errorf("cannot get CronJobs images: %s", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	images, err = getJobsImages(kubernetesClient)
	if err != nil {
		return nil, fmt.Errorf("cannot get Jobs images: %s", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	return deployedDockerImages, nil
}

func getPodsImages(kubernetesClient kubernetes.Interface) ([]string, error) {
	var images []string
	list, err := kubernetesClient.CoreV1().Pods("").List(v1.ListOptions{})
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

func getReplicationControllersImages(kubernetesClient kubernetes.Interface) ([]string, error) {
	var images []string
	list, err := kubernetesClient.CoreV1().ReplicationControllers("").List(v1.ListOptions{})
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

func getDeploymentsImages(kubernetesClient kubernetes.Interface) ([]string, error) {
	var images []string
	list, err := kubernetesClient.AppsV1().Deployments("").List(v1.ListOptions{})
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

func getStatefulSetsImages(kubernetesClient kubernetes.Interface) ([]string, error) {
	var images []string
	list, err := kubernetesClient.AppsV1().StatefulSets("").List(v1.ListOptions{})
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

func getDaemonSetsImages(kubernetesClient kubernetes.Interface) ([]string, error) {
	var images []string
	list, err := kubernetesClient.AppsV1().DaemonSets("").List(v1.ListOptions{})
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

func getReplicaSetsImages(kubernetesClient kubernetes.Interface) ([]string, error) {
	var images []string
	list, err := kubernetesClient.AppsV1().ReplicaSets("").List(v1.ListOptions{})
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

func getCronJobsImages(kubernetesClient kubernetes.Interface) ([]string, error) {
	var images []string
	list, err := kubernetesClient.BatchV1beta1().CronJobs("").List(v1.ListOptions{})
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

func getJobsImages(kubernetesClient kubernetes.Interface) ([]string, error) {
	var images []string
	list, err := kubernetesClient.BatchV1().Jobs("").List(v1.ListOptions{})
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
