package docker_registry

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
)

const (
	HarborImplementationName          = "harbor"
	harborRepositoryNotFoundErrPrefix = "harbor repository not found: "
)

var harborPatterns = []string{"^harbor\\..*"}

type HarborRepositoryNotFoundErr apiError

func NewHarborRepositoryNotFoundErr(err error) HarborRepositoryNotFoundErr {
	return HarborRepositoryNotFoundErr{
		error: errors.New(harborRepositoryNotFoundErrPrefix + err.Error()),
	}
}

func IsHarborRepositoryNotFoundErr(err error) bool {
	return strings.Contains(err.Error(), harborRepositoryNotFoundErrPrefix)
}

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
	d, err := newDefaultAPIForImplementation(HarborImplementationName, options.defaultImplementationOptions)
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

func (r *harbor) DeleteRepo(ctx context.Context, reference string) error {
	return r.deleteRepo(ctx, reference)
}

func (r *harbor) deleteRepo(ctx context.Context, reference string) error {
	hostname, repository, err := r.parseReference(reference)
	if err != nil {
		return err
	}

	resp, err := r.harborApi.DeleteRepository(ctx, hostname, repository, r.harborCredentials.username, r.harborCredentials.password)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return NewHarborRepositoryNotFoundErr(err)
		}

		return err
	}
	defer resp.Body.Close()

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
