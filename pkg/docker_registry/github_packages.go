package docker_registry

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"

	"github.com/werf/werf/pkg/image"
)

const GitHubPackagesImplementationName = "github"
const gitHubPackagesMetaTag = "docker-base-layer"

type GitHubPackagesUnauthorizedError apiError

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

func (r *gitHubPackages) Tags(ctx context.Context, reference string) ([]string, error) {
	tags, err := r.api.Tags(ctx, reference)
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

func (r *gitHubPackages) DeleteRepoImage(ctx context.Context, repoImage *image.Info) error {
	owner, project, packageName, err := r.parseReference(repoImage.Repository)
	if err != nil {
		return err
	}

	err = r.deletePackageVersion(ctx, owner, project, packageName, repoImage.Tag)
	if err != nil {
		return err
	}

	return nil
}

func (r *gitHubPackages) deletePackageVersion(ctx context.Context, owner, project, packageName, packageVersion string) error {
	processError := func(resp *http.Response, err error) error {
		if resp != nil && resp.StatusCode == http.StatusUnauthorized {
			return GitHubPackagesUnauthorizedError{error: err}
		}

		return err
	}

	packageVersionId, resp, err := r.gitHubApi.getPackageVersionId(ctx, owner, project, packageName, packageVersion, r.token)
	if err != nil {
		return processError(resp, err)
	}

	if resp, err := r.gitHubApi.deletePackageVersion(ctx, packageVersionId, r.token); err != nil {
		return processError(resp, err)
	}

	return nil
}

func (r *gitHubPackages) DeleteRepo(ctx context.Context, reference string) error {
	return r.deleteRepo(ctx, reference)
}

func (r *gitHubPackages) deleteRepo(ctx context.Context, reference string) error {
	owner, project, packageName, err := r.parseReference(reference)
	if err != nil {
		return err
	}

	tags, err := r.Tags(ctx, reference)
	for _, tag := range tags {
		if err := r.deletePackageVersion(ctx, owner, project, packageName, tag); err != nil {
			return err
		}
	}

	return nil
}

func (r *gitHubPackages) String() string {
	return GitHubPackagesImplementationName
}

func (r *gitHubPackages) parseReference(reference string) (string, string, string, error) {
	var owner, project, packageName string

	parsedReference, err := name.NewRepository(reference)
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
