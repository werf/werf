package docker_registry

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"

	"github.com/werf/werf/pkg/image"
)

const gitHubPackagesOldMetaTag = "docker-base-layer"

type gitHubPackagesOld struct {
	*gitHubPackagesBase
	gitHubGraphqlApi
}

func newGitHubPackagesOld(options gitHubPackagesOptions) (*gitHubPackagesOld, error) {
	base, err := newGitHubPackagesBase(options)
	if err != nil {
		return nil, err
	}

	gitHub := &gitHubPackagesOld{
		gitHubPackagesBase: base,
		gitHubGraphqlApi:   newGitHubGraphqlApi(),
	}

	return gitHub, nil
}

func (r *gitHubPackagesOld) Tags(ctx context.Context, reference string) ([]string, error) {
	tags, err := r.gitHubPackagesBase.Tags(ctx, reference)
	if err != nil {
		return nil, err
	}

	return r.exceptMetaTag(tags), nil
}

func (r *gitHubPackagesOld) exceptMetaTag(tags []string) []string {
	var result []string

	for _, tag := range tags {
		if tag == gitHubPackagesOldMetaTag {
			continue
		}

		result = append(result, tag)
	}

	return result
}

func (r *gitHubPackagesOld) DeleteRepoImage(ctx context.Context, repoImage *image.Info) error {
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

func (r *gitHubPackagesOld) deletePackageVersion(ctx context.Context, owner, project, packageName, packageVersion string) error {
	processError := func(resp *http.Response, err error) error {
		if resp != nil && resp.StatusCode == http.StatusUnauthorized {
			return GitHubPackagesUnauthorizedError{error: err}
		}

		return err
	}

	packageVersionId, resp, err := r.gitHubGraphqlApi.getPackageVersionId(ctx, owner, project, packageName, packageVersion, r.token)
	if err != nil {
		return processError(resp, err)
	}

	if resp, err = r.gitHubGraphqlApi.deletePackageVersion(ctx, packageVersionId, r.token); err != nil {
		return processError(resp, err)
	}

	return nil
}

func (r *gitHubPackagesOld) DeleteRepo(ctx context.Context, reference string) error {
	return r.deleteRepo(ctx, reference)
}

func (r *gitHubPackagesOld) deleteRepo(ctx context.Context, reference string) error {
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

func (r *gitHubPackagesOld) parseReference(reference string) (string, string, string, error) {
	parsedReference, err := name.NewTag(reference)
	if err != nil {
		return "", "", "", err
	}

	var owner, project, packageName string
	repositoryParts := strings.Split(parsedReference.RepositoryStr(), "/")
	if len(repositoryParts) == 2 {
		owner = repositoryParts[0]
		project = repositoryParts[1]
	} else if len(repositoryParts) == 3 {
		owner = repositoryParts[0]
		project = repositoryParts[1]
		packageName = repositoryParts[2]
	} else {
		return "", "", "", fmt.Errorf("unexpected reference %s", reference)
	}

	return owner, project, packageName, nil
}
