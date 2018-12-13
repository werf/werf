package cleanup

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flant/dapp/pkg/docker_registry"
	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/dapp/pkg/logger"
	"github.com/flant/dapp/pkg/slug"
	"github.com/flant/kubedog/pkg/kube"
)

type CleanupOptions struct {
	CommonRepoOptions CommonRepoOptions
	LocalRepo         GitRepo
	WithoutKube       bool
}

const (
	gitTagsExpiryDatePeriodPolicy    = 60 * 60 * 24 * 30
	gitTagsLimitPolicy               = 10
	gitCommitsExpiryDatePeriodPolicy = 60 * 60 * 24 * 30
	gitCommitsLimitPolicy            = 50
)

func Cleanup(options CleanupOptions) error {
	err := lock.WithLock(options.CommonRepoOptions.Repository, lock.LockOptions{Timeout: time.Second * 600}, func() error {
		repoDimgs, err := repoDimgImages(options.CommonRepoOptions)
		if err != nil {
			return err
		}

		if options.LocalRepo != nil {
			if !options.WithoutKube {
				repoDimgs, err = exceptRepoDimgsByWhitelist(repoDimgs)
				if err != nil {
					return err
				}
			}

			repoDimgs, err = repoDimgsCleanupByNonexistentGitPrimitive(repoDimgs, options)
			if err != nil {
				return err
			}

			repoDimgs, err = repoDimgsCleanupByPolicies(repoDimgs, options)
			if err != nil {
				return err
			}
		}

		if err := repoDimgstagesSyncByRepoDimgs(repoDimgs, options.CommonRepoOptions); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func exceptRepoDimgsByWhitelist(repoDimgs []docker_registry.RepoImage) ([]docker_registry.RepoImage, error) {
	var newRepoDimgs, exceptedRepoDimgs []docker_registry.RepoImage

	deployedDockerImages, err := deployedDockerImages()
	if err != nil {
		return nil, fmt.Errorf("cannot get deployed images: %s", err)
	}

Loop:
	for _, repoDimg := range repoDimgs {
		imageName := fmt.Sprintf("%s:%s", repoDimg.Repository, repoDimg.Tag)
		for _, deployedDockerImage := range deployedDockerImages {
			if deployedDockerImage == imageName {
				exceptedRepoDimgs = append(exceptedRepoDimgs, repoDimg)
				continue Loop
			}
		}

		newRepoDimgs = append(newRepoDimgs, repoDimg)
	}

	if len(exceptedRepoDimgs) != 0 {
		fmt.Println("Keep in repo images that are being used in kubernetes")
		for _, exceptedRepoDimg := range exceptedRepoDimgs {
			imageName := fmt.Sprintf("%s:%s", exceptedRepoDimg.Repository, exceptedRepoDimg.Tag)
			fmt.Println(imageName)
		}
		fmt.Println()
	}

	return newRepoDimgs, nil
}

func repoDimgsCleanupByNonexistentGitPrimitive(repoDimgs []docker_registry.RepoImage, options CleanupOptions) ([]docker_registry.RepoImage, error) {
	var nonexistentGitTagRepoImages, nonexistentGitCommitRepoImages, nonexistentGitBranchRepoImages []docker_registry.RepoImage

	gitTags, err := options.LocalRepo.TagsList()
	if err != nil {
		return nil, fmt.Errorf("cannot get local git tags list: %s", err)
	}

	gitBranches, err := options.LocalRepo.RemoteBranchesList()
	if err != nil {
		return nil, fmt.Errorf("cannot get local git branches list: %s", err)
	}

Loop:
	for _, repoDimg := range repoDimgs {
		labels, err := repoImageLabels(repoDimg)
		if err != nil {
			return nil, err
		}

		scheme, ok := labels["dapp-tag-scheme"]
		if !ok {
			continue
		}

		switch scheme {
		case "git_tag":
			if repoImageTagMatch(repoDimg, gitTags...) {
				continue Loop
			} else {
				nonexistentGitTagRepoImages = append(nonexistentGitTagRepoImages, repoDimg)
			}
		case "git_branch":
			if repoImageTagMatch(repoDimg, gitBranches...) {
				continue Loop
			} else {
				nonexistentGitBranchRepoImages = append(nonexistentGitBranchRepoImages, repoDimg)
			}
		case "git_commit":
			exist, err := options.LocalRepo.IsCommitExists(repoDimg.Tag)
			if err != nil {
				return nil, err
			}

			if !exist {
				nonexistentGitCommitRepoImages = append(nonexistentGitCommitRepoImages, repoDimg)
			}
		}
	}

	if len(nonexistentGitTagRepoImages) != 0 {
		fmt.Println("git tag nonexistent")
		if err := repoImagesRemove(nonexistentGitTagRepoImages, options.CommonRepoOptions); err != nil {
			return nil, err
		}
		fmt.Println()
		repoDimgs = exceptRepoImages(repoDimgs, nonexistentGitTagRepoImages...)
	}

	if len(nonexistentGitBranchRepoImages) != 0 {
		fmt.Println("git branch nonexistent")
		if err := repoImagesRemove(nonexistentGitBranchRepoImages, options.CommonRepoOptions); err != nil {
			return nil, err
		}
		fmt.Println()
		repoDimgs = exceptRepoImages(repoDimgs, nonexistentGitBranchRepoImages...)
	}

	if len(nonexistentGitCommitRepoImages) != 0 {
		fmt.Println("git commit nonexistent")
		if err := repoImagesRemove(nonexistentGitCommitRepoImages, options.CommonRepoOptions); err != nil {
			return nil, err
		}
		fmt.Println()
		repoDimgs = exceptRepoImages(repoDimgs, nonexistentGitCommitRepoImages...)
	}

	return repoDimgs, nil
}

func repoImageTagMatch(repoImage docker_registry.RepoImage, matches ...string) bool {
	for _, match := range matches {
		if repoImage.Tag == slug.DockerTag(match) {
			return true
		}
	}

	return false
}

func repoDimgsCleanupByPolicies(repoDimgs []docker_registry.RepoImage, options CleanupOptions) ([]docker_registry.RepoImage, error) {
	var repoDimgsWithGitTagScheme, repoDimgsWithGitCommitScheme []docker_registry.RepoImage

	for _, repoDimg := range repoDimgs {
		labels, err := repoImageLabels(repoDimg)
		if err != nil {
			return nil, err
		}

		scheme, ok := labels["dapp-tag-scheme"]
		if !ok {
			continue
		}

		switch scheme {
		case "git_tag":
			repoDimgsWithGitTagScheme = append(repoDimgsWithGitTagScheme, repoDimg)
		case "git_commit":
			repoDimgsWithGitCommitScheme = append(repoDimgsWithGitCommitScheme, repoDimg)
		}
	}

	cleanupByPolicyOptions := repoDimgsCleanupByPolicyOptions{
		expiryDatePeriod:  gitTagsExpiryDatePeriodPolicyValue(),
		expiryLimit:       gitTagsLimitPolicyValue(),
		gitPrimitive:      "tag",
		commonRepoOptions: options.CommonRepoOptions,
	}

	var err error
	repoDimgs, err = repoDimgsCleanupByPolicy(repoDimgs, repoDimgsWithGitTagScheme, cleanupByPolicyOptions)
	if err != nil {
		return nil, err
	}

	cleanupByPolicyOptions = repoDimgsCleanupByPolicyOptions{
		expiryDatePeriod:  gitCommitsExpiryDatePeriodPolicyValue(),
		expiryLimit:       gitCommitsLimitPolicyValue(),
		gitPrimitive:      "commit",
		commonRepoOptions: options.CommonRepoOptions,
	}

	repoDimgs, err = repoDimgsCleanupByPolicy(repoDimgs, repoDimgsWithGitCommitScheme, cleanupByPolicyOptions)
	if err != nil {
		return nil, err
	}

	return repoDimgs, nil
}

func gitTagsExpiryDatePeriodPolicyValue() int64 {
	return policyValue("EXPIRY_DATE_PERIOD_POLICY", gitTagsExpiryDatePeriodPolicy)
}

func gitTagsLimitPolicyValue() int64 {
	return policyValue("GIT_TAGS_LIMIT_POLICY", gitTagsLimitPolicy)
}

func gitCommitsExpiryDatePeriodPolicyValue() int64 {
	return policyValue("GIT_COMMITS_EXPIRY_DATE_PERIOD_POLICY", gitCommitsExpiryDatePeriodPolicy)
}

func gitCommitsLimitPolicyValue() int64 {
	return policyValue("GIT_COMMITS_LIMIT_POLICY", gitCommitsLimitPolicy)
}

func policyValue(envKey string, defaultValue int64) int64 {
	envValue := os.Getenv(envKey)
	if envValue != "" {
		value, err := strconv.ParseInt(envValue, 10, 64)
		if err != nil {
			logger.LogWarningF("WARNING: '%s' value '%s' is ignored (using default value '%s'\n", envKey, envValue, defaultValue)
		} else {
			return value
		}
	}

	return defaultValue
}

type repoDimgsCleanupByPolicyOptions struct {
	expiryDatePeriod  int64
	expiryLimit       int64
	gitPrimitive      string
	commonRepoOptions CommonRepoOptions
}

func repoDimgsCleanupByPolicy(repoDimgs, repoDimgsWithScheme []docker_registry.RepoImage, options repoDimgsCleanupByPolicyOptions) ([]docker_registry.RepoImage, error) {
	repoDimgsByRepository := make(map[string][]docker_registry.RepoImage)

	for _, repoDimgWithScheme := range repoDimgsWithScheme {
		if _, ok := repoDimgsByRepository[repoDimgWithScheme.Repository]; !ok {
			repoDimgsByRepository[repoDimgWithScheme.Repository] = []docker_registry.RepoImage{}
		}

		repoDimgsByRepository[repoDimgWithScheme.Repository] = append(repoDimgsByRepository[repoDimgWithScheme.Repository], repoDimgWithScheme)
	}

	expiryTime := time.Unix(time.Now().Unix()-options.expiryDatePeriod, 0)
	for repository, repositoryRepoDimgs := range repoDimgsByRepository {
		sort.Slice(repositoryRepoDimgs, func(i, j int) bool {
			iCreated, err := repoImageCreated(repositoryRepoDimgs[i])
			if err != nil {
				log.Fatal(err)
			}

			jCreated, err := repoImageCreated(repositoryRepoDimgs[j])
			if err != nil {
				log.Fatal(err)
			}

			return iCreated.Before(jCreated)
		})

		var notExpiredRepoDimgs, expiredRepoDimgs []docker_registry.RepoImage
		for _, repositoryRepoDimg := range repositoryRepoDimgs {
			created, err := repoImageCreated(repositoryRepoDimg)
			if err != nil {
				return nil, err
			}

			if created.Before(expiryTime) {
				expiredRepoDimgs = append(expiredRepoDimgs, repositoryRepoDimg)
			} else {
				notExpiredRepoDimgs = append(notExpiredRepoDimgs, repositoryRepoDimg)
			}
		}

		if len(expiredRepoDimgs) != 0 {
			fmt.Printf("%s: git %s date policy (created before %s)\n", repository, options.gitPrimitive, expiryTime.String())
			repoImagesRemove(expiredRepoDimgs, options.commonRepoOptions)
			fmt.Println()
			repoDimgs = exceptRepoImages(repoDimgs, expiredRepoDimgs...)
		}

		if int64(len(notExpiredRepoDimgs)) > options.expiryLimit {
			fmt.Printf("%s: git %s limit policy (> %d)\n", repository, options.gitPrimitive, options.expiryLimit)
			if err := repoImagesRemove(notExpiredRepoDimgs[options.expiryLimit:], options.commonRepoOptions); err != nil {
				return nil, err
			}
			fmt.Println()
			repoDimgs = exceptRepoImages(repoDimgs, notExpiredRepoDimgs[options.expiryLimit:]...)
		}
	}

	return repoDimgs, nil
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
