package docker_registry

import (
	"net/http"

	"github.com/google/go-containerregistry/pkg/name"
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

func (r *harbor) DeleteRepo(reference string) error {
	return r.deleteRepo(reference)
}

func (r *harbor) deleteRepo(reference string) error {
	hostname, repository, err := r.parseReference(reference)
	if err != nil {
		return err
	}

	resp, err := r.harborApi.DeleteRepository(hostname, repository, r.harborCredentials.username, r.harborCredentials.password)
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
