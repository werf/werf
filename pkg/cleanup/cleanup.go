package cleanup

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flant/kubedog/pkg/kube"

	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/logger"
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
	WithoutKube       bool
	Policies          ImagesCleanupPolicies
}

type StagesCleanupOptions struct {
	CommonRepoOptions    CommonRepoOptions
	CommonProjectOptions CommonProjectOptions
}

type CleanupOptions struct {
	ImagesCleanupOptions ImagesCleanupOptions
	StagesCleanupOptions StagesCleanupOptions
}

func Cleanup(options CleanupOptions) error {
	if err := ImagesCleanup(options.ImagesCleanupOptions); err != nil {
		return err
	}

	if err := StagesCleanup(options.StagesCleanupOptions); err != nil {
		return err
	}

	return nil
}

func ImagesCleanup(options ImagesCleanupOptions) error {
	err := lock.WithLock(options.CommonRepoOptions.ImagesRepo, lock.LockOptions{Timeout: time.Second * 600}, func() error {
		repoImages, err := repoImages(options.CommonRepoOptions)
		if err != nil {
			return err
		}

		if options.LocalGit != nil {
			if !options.WithoutKube {
				repoImages, err = exceptRepoImagesByWhitelist(repoImages)
				if err != nil {
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

	if err != nil {
		return err
	}

	return nil
}

func StagesCleanup(options StagesCleanupOptions) error {
	err := lock.WithLock(options.CommonRepoOptions.ImagesRepo, lock.LockOptions{Timeout: time.Second * 600}, func() error {
		repoImages, err := repoImages(options.CommonRepoOptions)
		if err != nil {
			return err
		}

		if options.CommonRepoOptions.StagesRepo == localStagesRepo {
			if err := projectImageStagesSyncByRepoImages(repoImages, options.CommonProjectOptions); err != nil {
				return err
			}
		} else {
			if err := repoImageStagesSyncByRepoImages(repoImages, options.CommonRepoOptions); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func exceptRepoImagesByWhitelist(repoImages []docker_registry.RepoImage) ([]docker_registry.RepoImage, error) {
	var newRepoImages, exceptedRepoImages []docker_registry.RepoImage

	deployedDockerImages, err := deployedDockerImages()
	if err != nil {
		return nil, fmt.Errorf("cannot get deployed images: %s", err)
	}

Loop:
	for _, repoImage := range repoImages {
		imageName := fmt.Sprintf("%s:%s", repoImage.Repository, repoImage.Tag)
		for _, deployedDockerImage := range deployedDockerImages {
			if deployedDockerImage == imageName {
				exceptedRepoImages = append(exceptedRepoImages, repoImage)
				continue Loop
			}
		}

		newRepoImages = append(newRepoImages, repoImage)
	}

	if len(exceptedRepoImages) != 0 {
		fmt.Fprintln(logger.GetOutStream(), "Keep in repo images that are being used in kubernetes")
		for _, exceptedRepoImage := range exceptedRepoImages {
			imageName := fmt.Sprintf("%s:%s", exceptedRepoImage.Repository, exceptedRepoImage.Tag)
			fmt.Fprintln(logger.GetOutStream(), imageName)
		}
		fmt.Fprintln(logger.GetOutStream())
	}

	return newRepoImages, nil
}

func repoImagesCleanupByNonexistentGitPrimitive(repoImages []docker_registry.RepoImage, options ImagesCleanupOptions) ([]docker_registry.RepoImage, error) {
	var nonexistentGitTagRepoImages, nonexistentGitCommitRepoImages, nonexistentGitBranchRepoImages []docker_registry.RepoImage

	gitTags, err := options.LocalGit.TagsList()
	if err != nil {
		return nil, fmt.Errorf("cannot get local git tags list: %s", err)
	}

	gitBranches, err := options.LocalGit.RemoteBranchesList()
	if err != nil {
		return nil, fmt.Errorf("cannot get local git branches list: %s", err)
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
			exist, err := options.LocalGit.IsCommitExists(repoImage.Tag)
			if err != nil {
				if strings.HasPrefix(err.Error(), "bad commit hash") {
					exist = false
				} else {
					return nil, err
				}
			}

			if !exist {
				nonexistentGitCommitRepoImages = append(nonexistentGitCommitRepoImages, repoImage)
			}
		}
	}

	if len(nonexistentGitTagRepoImages) != 0 {
		fmt.Fprintln(logger.GetOutStream(), "git tag nonexistent")
		if err := repoImagesRemove(nonexistentGitTagRepoImages, options.CommonRepoOptions); err != nil {
			return nil, err
		}
		fmt.Fprintln(logger.GetOutStream())
		repoImages = exceptRepoImages(repoImages, nonexistentGitTagRepoImages...)
	}

	if len(nonexistentGitBranchRepoImages) != 0 {
		fmt.Fprintln(logger.GetOutStream(), "git branch nonexistent")
		if err := repoImagesRemove(nonexistentGitBranchRepoImages, options.CommonRepoOptions); err != nil {
			return nil, err
		}
		fmt.Fprintln(logger.GetOutStream())
		repoImages = exceptRepoImages(repoImages, nonexistentGitBranchRepoImages...)
	}

	if len(nonexistentGitCommitRepoImages) != 0 {
		fmt.Fprintln(logger.GetOutStream(), "git commit nonexistent")
		if err := repoImagesRemove(nonexistentGitCommitRepoImages, options.CommonRepoOptions); err != nil {
			return nil, err
		}
		fmt.Fprintln(logger.GetOutStream())
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

func policyValue(envKey string, defaultValue int64) int64 {
	envValue := os.Getenv(envKey)
	if envValue != "" {
		value, err := strconv.ParseInt(envValue, 10, 64)
		if err != nil {
			logger.LogErrorF("WARNING: '%s' value '%s' is ignored (using default value '%s'\n", envKey, envValue, defaultValue)
		} else {
			return value
		}
	}

	return defaultValue
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
			fmt.Fprintf(logger.GetOutStream(), "%s: git %s date policy (created before %s)\n", repository, options.gitPrimitive, expiryTime.String())
			repoImagesRemove(expiredRepoImages, options.commonRepoOptions)
			fmt.Fprintln(logger.GetOutStream())
			repoImages = exceptRepoImages(repoImages, expiredRepoImages...)
		}

		if options.hasLimit && int64(len(notExpiredRepoImages)) > options.limit {
			fmt.Fprintf(logger.GetOutStream(), "%s: git %s limit policy (> %d)\n", repository, options.gitPrimitive, options.limit)
			if err := repoImagesRemove(notExpiredRepoImages[options.limit:], options.commonRepoOptions); err != nil {
				return nil, err
			}
			fmt.Fprintln(logger.GetOutStream())
			repoImages = exceptRepoImages(repoImages, notExpiredRepoImages[options.limit:]...)
		}
	}

	return repoImages, nil
}

func deployedDockerImages() ([]string, error) {
	var deployedDockerImages []string

	images, err := getPodsImages()
	if err != nil {
		return nil, fmt.Errorf("cannot get Pods images: %s", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	images, err = getReplicationControllersImages()
	if err != nil {
		return nil, fmt.Errorf("cannot get ReplicationControllers images: %s", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	images, err = getDeploymentsImages()
	if err != nil {
		return nil, fmt.Errorf("cannot get Deployments images: %s", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	images, err = getStatefulSetsImages()
	if err != nil {
		return nil, fmt.Errorf("cannot get StatefulSets images: %s", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	images, err = getDaemonSetsImages()
	if err != nil {
		return nil, fmt.Errorf("cannot get DaemonSets images: %s", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	images, err = getReplicaSetsImages()
	if err != nil {
		return nil, fmt.Errorf("cannot get ReplicaSets images: %s", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	images, err = getCronJobsImages()
	if err != nil {
		return nil, fmt.Errorf("cannot get CronJobs images: %s", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	images, err = getJobsImages()
	if err != nil {
		return nil, fmt.Errorf("cannot get Jobs images: %s", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	return deployedDockerImages, nil
}

func getPodsImages() ([]string, error) {
	var images []string
	list, err := kube.Kubernetes.CoreV1().Pods("").List(v1.ListOptions{})
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

func getReplicationControllersImages() ([]string, error) {
	var images []string
	list, err := kube.Kubernetes.CoreV1().ReplicationControllers("").List(v1.ListOptions{})
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

func getDeploymentsImages() ([]string, error) {
	var images []string
	list, err := kube.Kubernetes.AppsV1beta1().Deployments("").List(v1.ListOptions{})
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

func getStatefulSetsImages() ([]string, error) {
	var images []string
	list, err := kube.Kubernetes.AppsV1beta1().StatefulSets("").List(v1.ListOptions{})
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

func getDaemonSetsImages() ([]string, error) {
	var images []string
	list, err := kube.Kubernetes.ExtensionsV1beta1().DaemonSets("").List(v1.ListOptions{})
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

func getReplicaSetsImages() ([]string, error) {
	var images []string
	list, err := kube.Kubernetes.ExtensionsV1beta1().ReplicaSets("").List(v1.ListOptions{})
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

func getCronJobsImages() ([]string, error) {
	var images []string
	list, err := kube.Kubernetes.BatchV1beta1().CronJobs("").List(v1.ListOptions{})
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

func getJobsImages() ([]string, error) {
	var images []string
	list, err := kube.Kubernetes.BatchV1().Jobs("").List(v1.ListOptions{})
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
