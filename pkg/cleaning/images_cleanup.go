package cleaning

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/slug"
	"github.com/flant/werf/pkg/tag_strategy"
)

type ImagesCleanupPolicies struct {
	GitTagStrategyHasLimit bool // No limit by default!
	GitTagStrategyLimit    int64

	GitTagStrategyHasExpiryPeriod bool // No expiration by default!
	GitTagStrategyExpiryPeriod    time.Duration

	GitCommitStrategyHasLimit bool // No limit by default!
	GitCommitStrategyLimit    int64

	GitCommitStrategyHasExpiryPeriod bool // No expiration by default!
	GitCommitStrategyExpiryPeriod    time.Duration
}

type ImagesCleanupOptions struct {
	CommonRepoOptions CommonRepoOptions
	LocalGit          GitRepo
	KubernetesClients []kubernetes.Interface
	WithoutKube       bool
	Policies          ImagesCleanupPolicies
}

func ImagesCleanup(options ImagesCleanupOptions) error {
	logProcessOptions := logboek.LogProcessOptions{ColorizeMsgFunc: logboek.ColorizeHighlight}
	return logboek.LogProcess("Running images cleanup", logProcessOptions, func() error {
		return imagesCleanup(options)
	})
}

func imagesCleanup(options ImagesCleanupOptions) error {
	imagesCleanupLockName := fmt.Sprintf("images-cleanup.%s", options.CommonRepoOptions.ImagesRepo)
	return lock.WithLock(imagesCleanupLockName, lock.LockOptions{Timeout: time.Second * 600}, func() error {
		repoImages, err := repoImages(options.CommonRepoOptions)
		if err != nil {
			return err
		}

		if options.LocalGit != nil {
			if !options.WithoutKube {
				if err := logboek.LogProcess("Skipping repo images that are being used in kubernetes", logboek.LogProcessOptions{}, func() error {
					repoImages, err = exceptRepoImagesByWhitelist(repoImages, options.KubernetesClients)
					return err
				}); err != nil {
					return err
				}
			}

			repoImages, err = repoImagesCleanupByNonexistentGitPrimitive(repoImages, options)
			if err != nil {
				return err
			}

			repoImages, err = repoImagesCleanupByPolicies(repoImages, options)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func exceptRepoImagesByWhitelist(repoImages []docker_registry.RepoImage, kubernetesClients []kubernetes.Interface) ([]docker_registry.RepoImage, error) {
	var newRepoImages, exceptedRepoImages []docker_registry.RepoImage

	var deployedDockerImagesNames []string
	for _, kubernetesClient := range kubernetesClients {
		if err := logboek.LogProcessInline("Getting deployed docker images", logboek.LogProcessInlineOptions{}, func() error {
			kubernetesClientDeployedDockerImagesNames, err := deployedDockerImages(kubernetesClient)
			if err != nil {
				return fmt.Errorf("cannot get deployed images: %s", err)
			}

			deployedDockerImagesNames = append(deployedDockerImagesNames, kubernetesClientDeployedDockerImagesNames...)

			return nil
		}); err != nil {
			return nil, err
		}
	}

Loop:
	for _, repoImage := range repoImages {
		imageName := fmt.Sprintf("%s:%s", repoImage.Repository, repoImage.Tag)
		for _, deployedDockerImageName := range deployedDockerImagesNames {
			if deployedDockerImageName == imageName {
				exceptedRepoImages = append(exceptedRepoImages, repoImage)
				logboek.LogLn(imageName)
				continue Loop
			}
		}

		newRepoImages = append(newRepoImages, repoImage)
	}

	return newRepoImages, nil
}

func repoImagesCleanupByNonexistentGitPrimitive(repoImages []docker_registry.RepoImage, options ImagesCleanupOptions) ([]docker_registry.RepoImage, error) {
	var nonexistentGitTagRepoImages, nonexistentGitCommitRepoImages, nonexistentGitBranchRepoImages []docker_registry.RepoImage

	gitTags := []string{}
	gitBranches := []string{}

	if options.LocalGit != nil {
		var err error
		gitTags, err = options.LocalGit.TagsList()
		if err != nil {
			return nil, fmt.Errorf("cannot get local git tags list: %s", err)
		}

		gitBranches, err = options.LocalGit.RemoteBranchesList()
		if err != nil {
			return nil, fmt.Errorf("cannot get local git branches list: %s", err)
		}
	}

Loop:
	for _, repoImage := range repoImages {
		labels, err := repoImageLabels(repoImage)
		if err != nil {
			return nil, err
		}

		strategy, ok := labels[image.WerfTagStrategyLabel]
		if !ok {
			continue
		}

		switch strategy {
		case string(tag_strategy.GitTag):
			if repoImageTagMatch(repoImage, gitTags...) {
				continue Loop
			} else {
				nonexistentGitTagRepoImages = append(nonexistentGitTagRepoImages, repoImage)
			}
		case string(tag_strategy.GitBranch):
			if repoImageTagMatch(repoImage, gitBranches...) {
				continue Loop
			} else {
				nonexistentGitBranchRepoImages = append(nonexistentGitBranchRepoImages, repoImage)
			}
		case string(tag_strategy.GitCommit):
			exist := false

			if options.LocalGit != nil {
				var err error

				exist, err = options.LocalGit.IsCommitExists(repoImage.Tag)
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
		logboek.LogLn("Removed tags by nonexistent git-tag policy:")
		if err := logboek.WithIndent(func() error {
			return repoImagesRemove(nonexistentGitTagRepoImages, options.CommonRepoOptions)
		}); err != nil {
			return nil, err
		}
		logboek.LogOptionalLn()

		repoImages = exceptRepoImages(repoImages, nonexistentGitTagRepoImages...)
	}

	if len(nonexistentGitBranchRepoImages) != 0 {
		logboek.LogLn("Removed tags by nonexistent git-branch policy:")
		if err := logboek.WithIndent(func() error {
			return repoImagesRemove(nonexistentGitBranchRepoImages, options.CommonRepoOptions)
		}); err != nil {
			return nil, err
		}
		logboek.LogOptionalLn()

		repoImages = exceptRepoImages(repoImages, nonexistentGitBranchRepoImages...)
	}

	if len(nonexistentGitCommitRepoImages) != 0 {
		logboek.LogLn("Removed tags by nonexistent git-commit policy:")
		if err := logboek.WithIndent(func() error {
			return repoImagesRemove(nonexistentGitCommitRepoImages, options.CommonRepoOptions)
		}); err != nil {
			return nil, err
		}
		logboek.LogOptionalLn()

		repoImages = exceptRepoImages(repoImages, nonexistentGitCommitRepoImages...)
	}

	return repoImages, nil
}

func repoImageTagMatch(repoImage docker_registry.RepoImage, matches ...string) bool {
	for _, match := range matches {
		if repoImage.Tag == slug.DockerTag(match) {
			return true
		}
	}

	return false
}

func repoImagesCleanupByPolicies(repoImages []docker_registry.RepoImage, options ImagesCleanupOptions) ([]docker_registry.RepoImage, error) {
	var repoImagesWithGitTagScheme, repoImagesWithGitCommitScheme []docker_registry.RepoImage

	for _, repoImage := range repoImages {
		labels, err := repoImageLabels(repoImage)
		if err != nil {
			return nil, err
		}

		strategy, ok := labels[image.WerfTagStrategyLabel]
		if !ok {
			continue
		}

		switch strategy {
		case string(tag_strategy.GitTag):
			repoImagesWithGitTagScheme = append(repoImagesWithGitTagScheme, repoImage)
		case string(tag_strategy.GitCommit):
			repoImagesWithGitCommitScheme = append(repoImagesWithGitCommitScheme, repoImage)
		}
	}

	cleanupByPolicyOptions := repoImagesCleanupByPolicyOptions{
		hasLimit:          options.Policies.GitTagStrategyHasLimit,
		limit:             options.Policies.GitTagStrategyLimit,
		hasExpiryPeriod:   options.Policies.GitTagStrategyHasExpiryPeriod,
		expiryPeriod:      options.Policies.GitTagStrategyExpiryPeriod,
		gitPrimitive:      "tag",
		commonRepoOptions: options.CommonRepoOptions,
	}

	var err error
	repoImages, err = repoImagesCleanupByPolicy(repoImages, repoImagesWithGitTagScheme, cleanupByPolicyOptions)
	if err != nil {
		return nil, err
	}

	cleanupByPolicyOptions = repoImagesCleanupByPolicyOptions{
		hasLimit:          options.Policies.GitCommitStrategyHasLimit,
		limit:             options.Policies.GitCommitStrategyLimit,
		hasExpiryPeriod:   options.Policies.GitCommitStrategyHasExpiryPeriod,
		expiryPeriod:      options.Policies.GitCommitStrategyExpiryPeriod,
		gitPrimitive:      "commit",
		commonRepoOptions: options.CommonRepoOptions,
	}

	repoImages, err = repoImagesCleanupByPolicy(repoImages, repoImagesWithGitCommitScheme, cleanupByPolicyOptions)
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

	gitPrimitive      string
	commonRepoOptions CommonRepoOptions
}

func repoImagesCleanupByPolicy(repoImages, repoImagesWithScheme []docker_registry.RepoImage, options repoImagesCleanupByPolicyOptions) ([]docker_registry.RepoImage, error) {
	repoImagesByRepository := make(map[string][]docker_registry.RepoImage)

	for _, repoImageWithScheme := range repoImagesWithScheme {
		if _, ok := repoImagesByRepository[repoImageWithScheme.Repository]; !ok {
			repoImagesByRepository[repoImageWithScheme.Repository] = []docker_registry.RepoImage{}
		}

		repoImagesByRepository[repoImageWithScheme.Repository] = append(repoImagesByRepository[repoImageWithScheme.Repository], repoImageWithScheme)
	}

	var expiryTime time.Time
	if options.hasExpiryPeriod {
		expiryTime = time.Now().Add(time.Duration(-options.expiryPeriod))
	}

	for repository, repositoryRepoImages := range repoImagesByRepository {
		sort.Slice(repositoryRepoImages, func(i, j int) bool {
			iCreated, err := repoImageCreated(repositoryRepoImages[i])
			if err != nil {
				log.Fatal(err)
			}

			jCreated, err := repoImageCreated(repositoryRepoImages[j])
			if err != nil {
				log.Fatal(err)
			}

			return iCreated.Before(jCreated)
		})

		var notExpiredRepoImages, expiredRepoImages []docker_registry.RepoImage
		for _, repositoryRepoImage := range repositoryRepoImages {
			created, err := repoImageCreated(repositoryRepoImage)
			if err != nil {
				return nil, err
			}

			if options.hasExpiryPeriod && created.Before(expiryTime) {
				expiredRepoImages = append(expiredRepoImages, repositoryRepoImage)
			} else {
				notExpiredRepoImages = append(notExpiredRepoImages, repositoryRepoImage)
			}
		}

		if len(expiredRepoImages) != 0 {
			logboek.LogF("Removed repository %s tags by git-%s date policy (created before %s):\n", repository, options.gitPrimitive, expiryTime.String())
			if err := logboek.WithIndent(func() error {
				return repoImagesRemove(expiredRepoImages, options.commonRepoOptions)
			}); err != nil {
				return nil, err
			}
			logboek.LogOptionalLn()

			repoImages = exceptRepoImages(repoImages, expiredRepoImages...)
		}

		if options.hasLimit && int64(len(notExpiredRepoImages)) > options.limit {
			logboek.LogF("Removed repository %s tags by git-%s limit policy (> %d):\n", repository, options.gitPrimitive, options.limit)
			if err := logboek.WithIndent(func() error {
				return repoImagesRemove(notExpiredRepoImages[options.limit:], options.commonRepoOptions)
			}); err != nil {
				return nil, err
			}
			logboek.LogOptionalLn()

			repoImages = exceptRepoImages(repoImages, notExpiredRepoImages[options.limit:]...)
		}
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
	list, err := kubernetesClient.AppsV1beta1().Deployments("").List(v1.ListOptions{})
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
	list, err := kubernetesClient.AppsV1beta1().StatefulSets("").List(v1.ListOptions{})
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
	list, err := kubernetesClient.ExtensionsV1beta1().DaemonSets("").List(v1.ListOptions{})
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
	list, err := kubernetesClient.ExtensionsV1beta1().ReplicaSets("").List(v1.ListOptions{})
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
