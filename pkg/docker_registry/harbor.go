package docker_registry

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"

	"github.com/werf/werf/pkg/image"
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

func (r *harbor) Tags(ctx context.Context, reference string, _ ...Option) ([]string, error) {
	tags, err := r.defaultImplementation.Tags(ctx, reference)
	if err != nil {
		if IsHarborNotFoundError(err) {
			return []string{}, nil
		}
		return nil, err
	}

	return tags, nil
}

func (r *harbor) TryGetRepoImage(ctx context.Context, reference string) (*image.Info, error) {
	res, err := r.api.TryGetRepoImage(ctx, reference)
	if err != nil {
		if IsHarborNotFoundError(err) {
			return nil, nil
		}

		return nil, err
	}

	return res, err
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

func IsHarborNotFoundError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "NOT_FOUND")
}

func IsHarbor404Error(err error) bool {
	if err == nil {
		return false
	}

	// Example error:
	// GET https://domain/harbor/s3/object/name/prefix/docker/registry/v2/blobs/sha256/2d/3d8c68cd9df32f1beb4392298a123eac58aba1433a15b3258b2f3728bad4b7d1/data?X-Amz-Algorithm=REDACTED&X-Amz-Credential=REDACTED&X-Amz-Date=REDACTED&X-Amz-Expires=REDACTED&X-Amz-Signature=REDACTED&X-Amz-SignedHeaders=REDACTED: unsupported status code 404; body: <?xml version="1.0" encoding="UTF-8"?>
	// <Error><Code>NoSuchKey</Code><Message>The specified key does not exist.</Message><Resource>/harbor/s3/object/name/prefix/docker/registry/v2/blobs/sha256/3d/3d8c68cd9df32f1beb4392298a123eac58aba1433a15b3258b2f3728bad4b7d1/data</Resource><RequestId>c5bb943c-1e85-5930-b455-c3e8edbbaccd</RequestId></Error>

	return strings.Contains(err.Error(), "unsupported status code 404")
}
