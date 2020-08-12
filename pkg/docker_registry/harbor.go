package docker_registry

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"

	"github.com/werf/werf/pkg/image"
)

const HarborImplementationName = "harbor"

type HarborNotFoundError apiError

var harborPatterns = []string{"^harbor\\..*"}

type harbor struct {
	*defaultImplementation
	harborApi
	harborCredentials
}

type harborOptions struct {
	defaultImplementationOptions
	harborCredentials
}

type harborCredentials struct {
	username string
	password string
}

func newHarbor(options harborOptions) (*harbor, error) {
	d, err := newDefaultImplementation(options.defaultImplementationOptions)
	if err != nil {
		return nil, err
	}

	harbor := &harbor{
		defaultImplementation: d,
		harborCredentials:     options.harborCredentials,
		harborApi:             newHarborApi(),
	}

	return harbor, nil
}

func (r *harbor) Tags(ctx context.Context, reference string) ([]string, error) {
	tags, err := r.defaultImplementation.Tags(ctx, reference)
	if err != nil {
		if IsNotFoundError(err) {
			return []string{}, nil
		}
		return nil, err
	}

	return tags, nil
}

func (r *harbor) SelectRepoImageList(ctx context.Context, reference string, f func(string, *image.Info, error) (bool, error)) ([]*image.Info, error) {
	tags, err := r.Tags(ctx, reference)
	if err != nil {
		return nil, err
	}

	return r.selectRepoImageListByTags(ctx, reference, tags, f)
}

func (r *harbor) DeleteRepo(ctx context.Context, reference string) error {
	return r.deleteRepo(ctx, reference)
}

func (r *harbor) deleteRepo(ctx context.Context, reference string) error {
	hostname, repository, err := r.parseReference(reference)
	if err != nil {
		return err
	}

	resp, err := r.harborApi.DeleteRepository(ctx, hostname, repository, r.harborCredentials.username, r.harborCredentials.password)
	if resp != nil {
		if resp.StatusCode == http.StatusNotFound {
			return HarborNotFoundError{error: err}
		}
	}

	if err != nil {
		return err
	}

	return nil
}

func (r *harbor) String() string {
	return HarborImplementationName
}

func (r *harbor) parseReference(reference string) (string, string, error) {
	parsedReference, err := name.NewRepository(reference)
	if err != nil {
		return "", "", err
	}

	return parsedReference.RegistryStr(), parsedReference.RepositoryStr(), nil
}

func IsNotFoundError(err error) bool {
	return strings.Contains(err.Error(), "NOT_FOUND")
}
