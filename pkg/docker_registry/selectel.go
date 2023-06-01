package docker_registry

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"

	"github.com/werf/werf/pkg/image"
)

const SelectelImplementationName = "selectel"

const (
	selectelUnauthorizedErrPrefix       = "Selectel registry unauthorized: "
	selectelRepositoryNotFoundErrPrefix = "Selectel repository not found: "
)

var selectelPatterns = []string{"^cr.selcloud.ru"}

type (
	SelectelUnauthorizedErr       apiError
	SelectelRepositoryNotFoundErr apiError
)

func NewSelectelRepositoryNotFoundErr(err error) SelectelRepositoryNotFoundErr {
	return SelectelRepositoryNotFoundErr{
		error: errors.New(selectelRepositoryNotFoundErrPrefix + err.Error()),
	}
}

func IsSelectelRepositoryNotFoundErr(err error) bool {
	return strings.Contains(err.Error(), selectelRepositoryNotFoundErrPrefix)
}

func NewSelectelUnauthorizedErr(err error) SelectelUnauthorizedErr {
	return SelectelUnauthorizedErr{
		error: errors.New(selectelUnauthorizedErrPrefix + err.Error()),
	}
}

func IsSelectelUnauthorizedErr(err error) bool {
	return strings.Contains(err.Error(), selectelUnauthorizedErrPrefix)
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
	vpcID    string
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
	if r.hasExtraCredentials() {
		return r.tags(ctx, reference)
	}

	return r.api.tags(ctx, reference)
}

func (r *selectel) deleteRepoImage(ctx context.Context, repoImage *image.Info) error {
	token, registryID, err := r.getCredentials(ctx, repoImage.Repository)
	if err != nil {
		return err
	}

	hostname, _, repository, err := r.parseReference(repoImage.Repository)
	if err != nil {
		return err
	}

	//nolint:bodyclose
	// TODO: close response body
	resp, err := r.selectelApi.deleteReference(ctx, hostname, registryID, repository, repoImage.GetDigest(), token)
	if err != nil {
		return r.handleFailedApiResponse(resp, err)
	}

	return nil
}

func (r *selectel) deleteRepo(ctx context.Context, reference string) error {
	token, registryID, err := r.getCredentials(ctx, reference)
	if err != nil {
		return err
	}

	hostname, _, repository, err := r.parseReference(reference)
	if err != nil {
		return err
	}

	//nolint:bodyclose
	// TODO: close response body
	resp, err := r.selectelApi.deleteRepository(ctx, hostname, registryID, repository, token)
	if err != nil {
		return r.handleFailedApiResponse(resp, err)
	}

	return nil
}

func (r *selectel) tags(ctx context.Context, reference string) ([]string, error) {
	token, registryID, err := r.getCredentials(ctx, reference)
	if err != nil {
		return nil, err
	}

	hostname, _, repository, err := r.parseReference(reference)
	if err != nil {
		return nil, err
	}

	//nolint:bodyclose
	// TODO: close response body
	tags, resp, err := r.selectelApi.getTags(ctx, hostname, registryID, repository, token)
	if err != nil {
		return nil, r.handleFailedApiResponse(resp, err)
	}

	return tags, nil
}

func (r *selectel) getCredentials(ctx context.Context, reference string) (string, string, error) {
	token, err := r.getToken(ctx)
	if err != nil {
		return "", "", err
	}

	registryID, err := r.getRegistryId(ctx, token, reference)
	if err != nil {
		return "", "", err
	}

	return token, registryID, nil
}

func (r *selectel) getToken(ctx context.Context) (string, error) {
	if r.selectelCredentials.token != "" {
		return r.selectelCredentials.token, nil
	}

	token, err := r.selectelApi.getToken(ctx, r.selectelCredentials.username, r.selectelCredentials.password, r.selectelCredentials.account, r.selectelCredentials.vpc, r.selectelCredentials.vpcID)
	if err != nil {
		return "", err
	}

	r.selectelCredentials.token = token

	return r.selectelCredentials.token, nil
}

func (r *selectel) getRegistryId(ctx context.Context, token, reference string) (string, error) {
	if r.selectelRegistryId != "" {
		return r.selectelRegistryId, nil
	}

	hostname, registry, _, err := r.parseReference(reference)
	if err != nil {
		return "", err
	}

	//nolint:bodyclose
	// TODO: close response body
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
		return NewSelectelUnauthorizedErr(err)
	case http.StatusNotFound:
		return NewSelectelRepositoryNotFoundErr(err)
	}

	return nil
}

func (r *selectel) String() string {
	return SelectelImplementationName
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

	repository := url.QueryEscape(strings.Join(repositoryParts[1:], "/"))
	return parsedReference.RegistryStr(), repositoryParts[0], repository, nil
}

func (r *selectel) hasExtraCredentials() bool {
	credentials := selectelCredentials(r.selectelCredentials)
	if credentials.username != "" && credentials.password != "" && credentials.account != "" && (credentials.vpc != "" || credentials.vpcID != "") {
		return true
	}
	return false
}
