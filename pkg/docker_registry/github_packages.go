package docker_registry

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/google/go-containerregistry/pkg/name"

	"github.com/werf/werf/pkg/image"
)

const (
	GitHubPackagesImplementationName = "github"
	GitHubPackagesRegistryAddress    = "ghcr.io"
)

var gitHubPackagesPatterns = []string{"^ghcr\\.io"}

type (
	GitHubPackagesUnauthorizedErr apiError
	GitHubPackagesForbiddenErr    apiError
)

type gitHubPackages struct {
	*defaultImplementation
	gitHubCredentials
	gitHubApi
	isUserCache sync.Map
}

type gitHubCredentials struct {
	token string
}

type gitHubPackagesOptions struct {
	defaultImplementationOptions
	gitHubCredentials
}

func newGitHubPackages(options gitHubPackagesOptions) (*gitHubPackages, error) {
	d, err := newDefaultImplementation(options.defaultImplementationOptions)
	if err != nil {
		return nil, err
	}

	gitHub := &gitHubPackages{
		defaultImplementation: d,
		gitHubCredentials:     options.gitHubCredentials,
		gitHubApi:             newGitHubApi(),
		isUserCache:           sync.Map{},
	}

	return gitHub, nil
}

func (r *gitHubPackages) DeleteRepoImage(ctx context.Context, repoImage *image.Info) error {
	orgOrUserName, packageName, err := r.parseReference(repoImage.Repository)
	if err != nil {
		return err
	}

	isUser, err := r.isUser(ctx, orgOrUserName)
	if err != nil {
		return err
	}

	if isUser {
		packageVersionId, resp, err := r.gitHubApi.getUserContainerPackageVersionId(ctx, packageName, repoImage.Tag, r.token)
		if err != nil {
			return r.handleApiErr(resp, err)
		}

		if resp, err = r.gitHubApi.deleteUserContainerPackageVersion(ctx, packageName, packageVersionId, r.token); err != nil {
			return r.handleApiErr(resp, err)
		}

		return nil
	}

	packageVersionId, resp, err := r.gitHubApi.getOrgContainerPackageVersionId(ctx, orgOrUserName, packageName, repoImage.Tag, r.token)
	if err != nil {
		return r.handleApiErr(resp, err)
	}

	if resp, err = r.gitHubApi.deleteOrgContainerPackageVersion(ctx, orgOrUserName, packageName, packageVersionId, r.token); err != nil {
		return r.handleApiErr(resp, err)
	}

	return nil
}

func (r *gitHubPackages) DeleteRepo(ctx context.Context, reference string) error {
	orgOrUserName, packageName, err := r.parseReference(reference)
	if err != nil {
		return err
	}

	isUser, err := r.isUser(ctx, orgOrUserName)
	if err != nil {
		return err
	}

	if isUser {
		if resp, err := r.gitHubApi.deleteUserContainerPackage(ctx, packageName, r.token); err != nil {
			return r.handleApiErr(resp, err)
		}

		return nil
	}

	if resp, err := r.gitHubApi.deleteOrgContainerPackage(ctx, orgOrUserName, packageName, r.token); err != nil {
		return r.handleApiErr(resp, err)
	}

	return nil
}

func (r *gitHubPackages) isUser(ctx context.Context, orgOrUserName string) (bool, error) {
	isUser, ok := r.isUserCache.Load(orgOrUserName)
	if ok {
		return isUser.(bool), nil
	}

	user, resp, err := r.gitHubApi.getUser(ctx, orgOrUserName, r.token)
	if err != nil {
		return false, r.handleApiErr(resp, err)
	}

	isUser = user.Type == "User"
	r.isUserCache.Store(orgOrUserName, isUser)

	return isUser.(bool), nil
}

func (r *gitHubPackages) handleApiErr(resp *http.Response, err error) error {
	if resp != nil {
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return GitHubPackagesUnauthorizedErr{error: err}
		case http.StatusForbidden:
			return GitHubPackagesForbiddenErr{error: err}
		}
	}

	return err
}

func (r *gitHubPackages) String() string {
	return GitHubPackagesImplementationName
}

func (r *gitHubPackages) parseReference(reference string) (string, string, error) {
	parsedReference, err := name.NewTag(reference)
	if err != nil {
		return "", "", fmt.Errorf("unable to parse reference %q: %s", reference, err)
	}

	repositoryWithoutRegistry := strings.TrimPrefix(parsedReference.RepositoryStr(), parsedReference.RegistryStr()+"/")
	orgOrUserNameAndPackageName := strings.SplitN(repositoryWithoutRegistry, "/", 2)
	orgOrUserName := orgOrUserNameAndPackageName[0]
	packageName := strings.ReplaceAll(orgOrUserNameAndPackageName[1], "/", "%2F")

	if orgOrUserName == "" || packageName == "" {
		return "", "", fmt.Errorf("unexpected reference %s: cannot parse organization and package name", reference)
	}

	return orgOrUserName, packageName, nil
}
