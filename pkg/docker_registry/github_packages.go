package docker_registry

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"

	"github.com/flant/werf/pkg/image"
)

const GitHubPackagesImplementationName = "github"
const gitHubPackagesMetaTag = "docker-base-layer"

var gitHubPackagesPatterns = []string{"^docker\\.pkg\\.github\\.com"}

type gitHubPackages struct {
	*defaultImplementation
	gitHubApi
	gitHubCredentials
}

type gitHubPackagesOptions struct {
	defaultImplementationOptions
	gitHubCredentials
}

type gitHubCredentials struct {
	token string
}

func newGitHubPackages(options gitHubPackagesOptions) (*gitHubPackages, error) {
	d, err := newDefaultImplementation(options.defaultImplementationOptions)
	if err != nil {
		return nil, err
	}

	gitHub := &gitHubPackages{
		defaultImplementation: d,
		gitHubApi:             newGitHubApi(),
		gitHubCredentials:     options.gitHubCredentials,
	}

	return gitHub, nil
}

func (r *gitHubPackages) Tags(reference string) ([]string, error) {
	tags, err := r.api.Tags(reference)
	if err != nil {
		return nil, err
	}

	return r.exceptMetaTag(tags), nil
}

func (r *gitHubPackages) exceptMetaTag(tags []string) []string {
	var result []string

	for _, tag := range tags {
		if tag == gitHubPackagesMetaTag {
			continue
		}

		result = append(result, tag)
	}

	return result
}

func (r *gitHubPackages) SelectRepoImageList(reference string, f func(*image.Info) bool) ([]*image.Info, error) {
	tags, err := r.Tags(reference)
	if err != nil {
		return nil, err
	}

	return r.defaultImplementation.selectRepoImageListByTags(reference, tags, f)
}

func (r *gitHubPackages) DeleteRepoImage(repoImageList ...*image.Info) error {
	for _, repoImage := range repoImageList {
		if err := r.deleteRepoImage(repoImage); err != nil {
			return err
		}
	}

	return nil
}

func (r *gitHubPackages) deleteRepoImage(repoImage *image.Info) error {
	owner, project, packageName, err := r.parseReference(repoImage.Repository)
	if err != nil {
		return err
	}

	err = r.deletePackageVersion(owner, project, packageName, repoImage.Tag)
	if err != nil {
		return err
	}

	return nil
}

func (r *gitHubPackages) deletePackageVersion(owner, project, packageName, packageVersion string) error {
	processError := func(resp *http.Response, err error) error {
		if resp != nil && resp.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("authorization failed: you should use a token with the read:packages, write:packages, delete:packages and repo scopes. For more information see https://help.github.com/en/packages/publishing-and-managing-packages/about-github-packages#about-tokens")
		}

		return err
	}

	packageVersionId, resp, err := r.gitHubApi.getPackageVersionId(owner, project, packageName, packageVersion, r.token)
	if err != nil {
		return processError(resp, err)
	}

	if resp, err := r.gitHubApi.deletePackageVersion(packageVersionId, r.token); err != nil {
		return processError(resp, err)
	}

	return nil
}

func (r *gitHubPackages) DeleteRepo(reference string) error {
	return r.deleteRepo(reference)
}

func (r *gitHubPackages) deleteRepo(reference string) error {
	owner, project, packageName, err := r.parseReference(reference)
	if err != nil {
		return err
	}

	tags, err := r.Tags(reference)
	for _, tag := range tags {
		if err := r.deletePackageVersion(owner, project, packageName, tag); err != nil {
			return err
		}
	}

	return nil
}

func (r *gitHubPackages) ResolveRepoMode(registryOrRepositoryAddress, repoMode string) (string, error) {
	_, _, packageName, err := r.parseReference(registryOrRepositoryAddress)
	if err != nil {
		return "", err
	}

	switch repoMode {
	case MonorepoRepoMode:
		if packageName != "" {
			return MonorepoRepoMode, nil
		}

		return "", fmt.Errorf("docker registry implementation %[1]s and repo mode %[2]s cannot be used with %[4]s (add repository to address or use %[3]s repo mode)", r.String(), MonorepoRepoMode, MultirepoRepoMode, registryOrRepositoryAddress)
	case MultirepoRepoMode:
		if packageName == "" {
			return MultirepoRepoMode, nil
		}

		return "", fmt.Errorf("docker registry implementation %[1]s and repo mode %[3]s cannot be used with %[4]s (exclude repository from address or use %[2]s repo mode)", r.String(), MonorepoRepoMode, MultirepoRepoMode, registryOrRepositoryAddress)
	case "auto", "":
		if packageName == "" {
			return MultirepoRepoMode, nil
		} else {
			return MonorepoRepoMode, nil
		}
	default:
		return "", fmt.Errorf("docker registry implementation %s does not support repo mode %s", r.String(), repoMode)
	}
}

func (r *gitHubPackages) String() string {
	return GitHubPackagesImplementationName
}

func (r *gitHubPackages) parseReference(reference string) (string, string, string, error) {
	var owner, project, packageName string

	parsedReference, err := name.NewTag(reference)
	if err != nil {
		return "", "", "", err
	}

	repositoryParts := strings.Split(parsedReference.RepositoryStr(), "/")
	if len(repositoryParts) == 2 {
		owner = repositoryParts[0]
		project = repositoryParts[1]
	} else if len(repositoryParts) == 3 {
		owner = repositoryParts[0]
		project = repositoryParts[1]
		packageName = repositoryParts[2]
	} else {
		return "", "", "", fmt.Errorf("unexpeced reference %s", reference)
	}

	return owner, project, packageName, nil
}
