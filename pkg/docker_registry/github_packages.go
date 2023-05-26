package docker_registry

import (
	"context"
	"errors"
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

	gitHubPackagesUnauthorizedErrPrefix = "gitHub packages unauthorized: "
	gitHubPackagesForbiddenErrPrefix    = "gitHub packages forbidden: "
)

var gitHubPackagesPatterns = []string{"^ghcr\\.io", "^docker\\.pkg\\.github\\.com"}

type (
	GitHubPackagesUnauthorizedErr apiError
	GitHubPackagesForbiddenErr    apiError
)

func NewGitHubPackagesUnauthorizedErr(err error) GitHubPackagesUnauthorizedErr {
	return GitHubPackagesUnauthorizedErr{
		error: errors.New(gitHubPackagesUnauthorizedErrPrefix + err.Error()),
	}
}

func IsGitHubPackagesUnauthorizedErr(err error) bool {
	return strings.Contains(err.Error(), gitHubPackagesUnauthorizedErrPrefix)
}

func NewGitHubPackagesForbiddenErr(err error) GitHubPackagesForbiddenErr {
	return GitHubPackagesForbiddenErr{
		error: errors.New(gitHubPackagesForbiddenErrPrefix + err.Error()),
	}
}

func IsGitHubPackagesForbiddenErr(err error) bool {
	return strings.Contains(err.Error(), gitHubPackagesForbiddenErrPrefix)
}

type gitHubPackages struct {
	*defaultImplementation
	gitHubCredentials
	gitHubApi

	isUserCache                sync.Map
	tagIDPackageVersionIDCache map[string]string
	getTagPackageVersionIDLock sync.Mutex
}

type gitHubCredentials struct {
	token string
}

type gitHubPackagesOptions struct {
	defaultImplementationOptions
	gitHubCredentials
}

func newGitHubPackages(options gitHubPackagesOptions) (*gitHubPackages, error) {
	d, err := newDefaultAPIForImplementation(GitHubPackagesImplementationName, options.defaultImplementationOptions)
	if err != nil {
		return nil, err
	}

	gitHub := &gitHubPackages{
		defaultImplementation:      d,
		gitHubCredentials:          options.gitHubCredentials,
		gitHubApi:                  newGitHubApi(),
		isUserCache:                sync.Map{},
		tagIDPackageVersionIDCache: map[string]string{},
		getTagPackageVersionIDLock: sync.Mutex{},
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

	packageVersionId, err := r.getTagPackageVersionID(ctx, orgOrUserName, packageName, repoImage.Tag)
	if err != nil {
		return err
	}

	if isUser {
		//nolint:bodyclose
		// TODO: close response body
		if resp, err := r.gitHubApi.deleteUserContainerPackageVersion(ctx, packageName, packageVersionId, r.token); err != nil {
			return r.handleFailedApiResponse(resp, err)
		}

		return nil
	}

	//nolint:bodyclose
	// TODO: close response body
	if resp, err := r.gitHubApi.deleteOrgContainerPackageVersion(ctx, orgOrUserName, packageName, packageVersionId, r.token); err != nil {
		return r.handleFailedApiResponse(resp, err)
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
		//nolint:bodyclose
		// TODO: close response body
		if resp, err := r.gitHubApi.deleteUserContainerPackage(ctx, packageName, r.token); err != nil {
			return r.handleFailedApiResponse(resp, err)
		}

		return nil
	}

	//nolint:bodyclose
	// TODO: close response body
	if resp, err := r.gitHubApi.deleteOrgContainerPackage(ctx, orgOrUserName, packageName, r.token); err != nil {
		return r.handleFailedApiResponse(resp, err)
	}

	return nil
}

func (r *gitHubPackages) getTagPackageVersionID(ctx context.Context, orgOrUserName, packageName, tag string) (string, error) {
	r.getTagPackageVersionIDLock.Lock()
	defer r.getTagPackageVersionIDLock.Unlock()

	if versionID, ok := r.tagIDPackageVersionIDCache[r.tagID(orgOrUserName, packageName, tag)]; ok {
		return versionID, nil
	}

	if err := r.populateTagIDPackageVersionIDCache(ctx, orgOrUserName, packageName); err != nil {
		return "", err
	}

	if versionID, ok := r.tagIDPackageVersionIDCache[r.tagID(orgOrUserName, packageName, tag)]; ok {
		return versionID, nil
	}

	return "", fmt.Errorf("container package version id for tag %q not found", tag)
}

func (r *gitHubPackages) populateTagIDPackageVersionIDCache(ctx context.Context, orgOrUserName, packageName string) error {
	isUser, err := r.isUser(ctx, orgOrUserName)
	if err != nil {
		return err
	}

	handleFunc := func(versionList []githubApiVersion) error {
		for _, version := range versionList {
			for _, versionTag := range version.Metadata.Container.Tags {
				r.tagIDPackageVersionIDCache[r.tagID(orgOrUserName, packageName, versionTag)] = fmt.Sprintf("%d", version.Id)
			}
		}

		return nil
	}

	if isUser {
		//nolint:bodyclose
		// TODO: close response body
		if resp, err := r.gitHubApi.getUserContainerPackageVersionsInBatches(ctx, packageName, r.token, handleFunc); err != nil {
			return r.handleFailedApiResponse(resp, err)
		}

		return nil
	}

	//nolint:bodyclose
	// TODO: close response body
	if resp, err := r.gitHubApi.getOrgContainerPackageVersionsInBatches(ctx, orgOrUserName, packageName, r.token, handleFunc); err != nil {
		return r.handleFailedApiResponse(resp, err)
	}

	return nil
}

func (r *gitHubPackages) tagID(orgOrUserName, packageName, tag string) string {
	return strings.Join([]string{orgOrUserName, packageName, tag}, "-")
}

func (r *gitHubPackages) isUser(ctx context.Context, orgOrUserName string) (bool, error) {
	isUser, ok := r.isUserCache.Load(orgOrUserName)
	if ok {
		return isUser.(bool), nil
	}

	//nolint:bodyclose
	// TODO: close response body
	user, resp, err := r.gitHubApi.getUser(ctx, orgOrUserName, r.token)
	if err != nil {
		return false, r.handleFailedApiResponse(resp, err)
	}

	isUser = user.Type == "User"
	r.isUserCache.Store(orgOrUserName, isUser)

	return isUser.(bool), nil
}

func (r *gitHubPackages) handleFailedApiResponse(resp *http.Response, err error) error {
	if resp != nil {
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return NewGitHubPackagesUnauthorizedErr(err)
		case http.StatusForbidden:
			return NewGitHubPackagesForbiddenErr(err)
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
		return "", "", fmt.Errorf("unable to parse reference %q: %w", reference, err)
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
