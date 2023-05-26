package docker_registry

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"

	"github.com/werf/werf/pkg/image"
)

const DockerHubImplementationName = "dockerhub"

const (
	dockerHubUnauthorizedErrPrefix       = "docker hub unauthorized: "
	dockerHubRepositoryNotFoundErrPrefix = "docker hub repository not found: "
)

type (
	DockerHubUnauthorizedErr       apiError
	DockerHubRepositoryNotFoundErr apiError
)

func NewDockerHubUnauthorizedErr(err error) DockerHubUnauthorizedErr {
	return DockerHubUnauthorizedErr{
		error: errors.New(dockerHubUnauthorizedErrPrefix + err.Error()),
	}
}

func IsDockerHubUnauthorizedErr(err error) bool {
	return strings.Contains(err.Error(), dockerHubUnauthorizedErrPrefix)
}

func NewDockerHubRepositoryNotFoundErr(err error) DockerHubRepositoryNotFoundErr {
	return DockerHubRepositoryNotFoundErr{
		error: errors.New(dockerHubRepositoryNotFoundErrPrefix + err.Error()),
	}
}

func IsDockerHubRepositoryNotFoundErr(err error) bool {
	return strings.Contains(err.Error(), dockerHubRepositoryNotFoundErrPrefix)
}

var dockerHubPatterns = []string{"^index\\.docker\\.io", "^registry\\.hub\\.docker\\.com"}

type dockerHub struct {
	*defaultImplementation
	dockerHubApi
	dockerHubCredentials
}

type dockerHubOptions struct {
	defaultImplementationOptions
	dockerHubCredentials
}

type dockerHubCredentials struct {
	username string
	password string
	token    string
}

func newDockerHub(options dockerHubOptions) (*dockerHub, error) {
	d, err := newDefaultAPIForImplementation(DockerHubImplementationName, options.defaultImplementationOptions)
	if err != nil {
		return nil, err
	}

	dockerHub := &dockerHub{
		defaultImplementation: d,
		dockerHubApi:          newDockerHubApi(),
		dockerHubCredentials:  options.dockerHubCredentials,
	}

	return dockerHub, nil
}

func (r *dockerHub) DeleteRepo(ctx context.Context, reference string) error {
	return r.deleteRepo(ctx, reference)
}

func (r *dockerHub) DeleteRepoImage(ctx context.Context, repoImage *image.Info) error {
	token, err := r.getToken(ctx)
	if err != nil {
		return err
	}

	account, project, err := r.parseReference(repoImage.Repository)
	if err != nil {
		return err
	}

	//nolint:bodyclose
	// TODO: close response body
	resp, err := r.dockerHubApi.deleteTag(ctx, account, project, repoImage.Tag, token)
	if err != nil {
		return r.handleFailedApiResponse(resp, err)
	}

	return nil
}

func (r *dockerHub) deleteRepo(ctx context.Context, reference string) error {
	token, err := r.getToken(ctx)
	if err != nil {
		return err
	}

	account, project, err := r.parseReference(reference)
	if err != nil {
		return err
	}

	//nolint:bodyclose
	// TODO: close response body
	resp, err := r.dockerHubApi.deleteRepository(ctx, account, project, token)
	if err != nil {
		return r.handleFailedApiResponse(resp, err)
	}

	return nil
}

func (r *dockerHub) getToken(ctx context.Context) (string, error) {
	if r.dockerHubCredentials.token != "" {
		return r.dockerHubCredentials.token, nil
	}

	//nolint:bodyclose
	// TODO: close response body
	token, resp, err := r.dockerHubApi.getToken(ctx, r.dockerHubCredentials.username, r.dockerHubCredentials.password)
	if err != nil {
		return "", r.handleFailedApiResponse(resp, err)
	}

	r.dockerHubCredentials.token = token

	return r.dockerHubCredentials.token, nil
}

func (r *dockerHub) handleFailedApiResponse(resp *http.Response, err error) error {
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return NewDockerHubUnauthorizedErr(err)
	case http.StatusNotFound:
		return NewDockerHubRepositoryNotFoundErr(err)
	}

	return nil
}

func (r *dockerHub) String() string {
	return DockerHubImplementationName
}

func (r *dockerHub) parseReference(reference string) (string, string, error) {
	parsedReference, err := name.NewRepository(reference)
	if err != nil {
		return "", "", err
	}

	repositoryParts := strings.Split(parsedReference.RepositoryStr(), "/")

	if len(repositoryParts) != 2 {
		return "", "", fmt.Errorf("unexpected reference %s", reference)
	}

	account := repositoryParts[0]
	project := repositoryParts[1]
	return account, project, nil
}
