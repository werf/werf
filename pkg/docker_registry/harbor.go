package docker_registry

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/google/go-containerregistry/pkg/name"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/image"
)

const (
	HarborImplementationName          = "harbor"
	harborRepositoryNotFoundErrPrefix = "harbor repository not found: "
)

var harborPatterns = []string{"^harbor\\..*", "demo\\.goharbor\\.io"}

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

func (r *harbor) GetRepoImage(ctx context.Context, reference string) (*image.Info, error) {
	var info *image.Info
	operation := func() error {
		var err error
		info, err = r.api.GetRepoImage(ctx, reference)
		if err != nil {
			if strings.Contains(err.Error(), "PROJECTPOLICYVIOLATION") {
				return err
			}

			// Do not retry on other errors.
			return backoff.Permanent(err)
		}
		return nil
	}

	notify := func(err error, duration time.Duration) {
		logboek.Context(ctx).Warn().LogF("WARNING: %s. Retrying in %v...\n", err.Error(), duration)
	}

	eb := backoff.NewExponentialBackOff()
	eb.InitialInterval = 2 * time.Second
	eb.MaxElapsedTime = 2 * time.Minute // Maximum time for all retries.

	err := backoff.RetryNotify(operation, eb, notify)
	if err != nil {
		return nil, err
	}

	return info, nil
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
