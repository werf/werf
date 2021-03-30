package docker_registry

import (
	"context"
	"fmt"
	"strings"

	"github.com/werf/werf/pkg/image"
)

const DefaultImplementationName = "default"

type defaultImplementation struct {
	*api
}

type defaultImplementationOptions struct {
	apiOptions
}

func newDefaultImplementation(options defaultImplementationOptions) (*defaultImplementation, error) {
	d := &defaultImplementation{}
	d.api = newAPI(options.apiOptions)
	return d, nil
}

func (r *defaultImplementation) CreateRepo(_ context.Context, _ string) error {
	return fmt.Errorf("method is not implemented")
}

func (r *defaultImplementation) DeleteRepo(_ context.Context, _ string) error {
	return fmt.Errorf("method is not implemented")
}

func (r *defaultImplementation) DeleteRepoImage(_ context.Context, repoImage *image.Info) error {
	reference := strings.Join([]string{repoImage.Repository, repoImage.RepoDigest}, "@")
	return r.api.deleteImageByReference(reference)
}

func (r *defaultImplementation) String() string {
	return DefaultImplementationName
}

func IsManifestUnknownError(err error) bool {
	return (err != nil) && strings.Contains(err.Error(), "MANIFEST_UNKNOWN")
}

func IsBlobUnknownError(err error) bool {
	return (err != nil) && strings.Contains(err.Error(), "BLOB_UNKNOWN")
}

func IsNameUnknownError(err error) bool {
	return (err != nil) && strings.Contains(err.Error(), "NAME_UNKNOWN")
}
