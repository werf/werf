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

const (
	SelectelImplementationName          = "selectel"
	SelectelRepositoryNotFoundErrPrefix = "Selectel repository not found: "
)

var selectelPatterns = []string{"^cr.selcloud.ru"}

type SelectelRepositoryNotFoundErr apiError

func NewSelectelRepositoryNotFoundErr(err error) SelectelRepositoryNotFoundErr {
	return SelectelRepositoryNotFoundErr{
		error: errors.New(SelectelRepositoryNotFoundErrPrefix + err.Error()),
	}
}

func IsSelectelRepositoryNotFoundErr(err error) bool {
	return strings.Contains(err.Error(), SelectelRepositoryNotFoundErrPrefix)
}

type selectel struct {
	*defaultImplementation
	selectelApi
	selectelCredentials
	selectelRegistryId string
}

type selectelOptions struct {
	defaultImplementationOptions
	selectelCredentials
}

type selectelCredentials struct {
	username string
	password string
	account  string
	vpc      string
	token    string
}

func newSelectel(options selectelOptions) (*selectel, error) {
	d, err := newDefaultAPIForImplementation(SelectelImplementationName, options.defaultImplementationOptions)
	if err != nil {
		return nil, err
	}

	Selectel := &selectel{
		defaultImplementation: d,
		selectelCredentials:   options.selectelCredentials,
		selectelApi:           newSelectelApi(),
	}

	return Selectel, nil
}

func (r *selectel) DeleteRepo(ctx context.Context, reference string) error {
	return r.deleteRepo(ctx, reference)
}

func (r *selectel) DeleteRepoImage(ctx context.Context, repoImage *image.Info) error {
	return r.deleteRepoImage(ctx, repoImage)
}

func (r *selectel) Tags(ctx context.Context, reference string, _ ...Option) ([]string, error) {
	return r.tags(ctx, reference)
}

func (r *selectel) deleteRepoImage(ctx context.Context, repoImage *image.Info) error {
	token, err := r.getToken(ctx)
	if err != nil {
		return err
	}

	registryID, err := r.getRegistryId(ctx, repoImage.Repository)
	if err != nil {
		return err
	}

	hostname, _, repository, err := r.parseReference(repoImage.Repository)
	if err != nil {
		return err
	}

	resp, err := r.selectelApi.deleteReference(ctx, hostname, registryID, repository, repoImage.RepoDigest, token)
	if err != nil {
		return r.handleFailedApiResponse(resp, err)
	}

	return nil
}

func (r *selectel) deleteRepo(ctx context.Context, reference string) error {
	token, err := r.getToken(ctx)
	if err != nil {
		return err
	}

	registryID, err := r.getRegistryId(ctx, reference)
	if err != nil {
		return err
	}

	hostname, _, repository, err := r.parseReference(reference)
	if err != nil {
		return err
	}

	resp, err := r.selectelApi.deleteRepository(ctx, hostname, registryID, repository, token)
	if err != nil {
		return r.handleFailedApiResponse(resp, err)
	}

	return nil
}

func (r *selectel) tags(ctx context.Context, reference string) ([]string, error) {
	token, err := r.getToken(ctx)
	if err != nil {
		return nil, err
	}

	registryID, err := r.getRegistryId(ctx, reference)
	if err != nil {
		return nil, err
	}

	hostname, _, repository, err := r.parseReference(reference)
	if err != nil {
		return nil, err
	}

	tags, resp, err := r.selectelApi.getTags(ctx, hostname, registryID, repository, token)
	if err != nil {
		return nil, r.handleFailedApiResponse(resp, err)
	}

	return tags, nil
}

func (r *selectel) getToken(ctx context.Context) (string, error) {
	if r.selectelCredentials.token != "" {
		return r.selectelCredentials.token, nil
	}

	token, err := r.selectelApi.getToken(ctx, r.selectelCredentials.username, r.selectelCredentials.password, r.selectelCredentials.account, r.selectelCredentials.vpc)
	if err != nil {
		return "", err
	}

	r.selectelCredentials.token = token

	return r.selectelCredentials.token, nil
}

func (r *selectel) getRegistryId(ctx context.Context, reference string) (string, error) {
	if r.selectelRegistryId != "" {
		return r.selectelRegistryId, nil
	}

	token, err := r.getToken(ctx)
	if err != nil {
		return "", err
	}

	hostname, registry, _, err := r.parseReference(reference)
	if err != nil {
		return "", err
	}

	registryId, resp, err := r.selectelApi.getRegistryId(ctx, hostname, registry, token)
	if err != nil {
		return "", r.handleFailedApiResponse(resp, err)
	}

	r.selectelRegistryId = registryId

	return r.selectelRegistryId, nil
}

func (r *selectel) handleFailedApiResponse(resp *http.Response, err error) error {
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return NewDockerHubUnauthorizedErr(err)
	case http.StatusNotFound:
		return NewDockerHubRepositoryNotFoundErr(err)
	}

	return nil
}

func (r *selectel) String() string {
	return DockerHubImplementationName
}

func (r *selectel) parseReference(reference string) (string, string, string, error) {
	parsedReference, err := name.NewRepository(reference)
	if err != nil {
		return "", "", "", err
	}

	repositoryParts := strings.Split(parsedReference.RepositoryStr(), "/")

	if len(repositoryParts) == 0 {
		return "", "", "", fmt.Errorf("unexpected reference %s", reference)
	}

	repository := strings.Join(repositoryParts[1:], "/")
	return parsedReference.RegistryStr(), repositoryParts[0], repository, nil
}
