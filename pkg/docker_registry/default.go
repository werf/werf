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

func (r *defaultImplementation) CheckRepoImageCustomTag(ctx context.Context, repoImage *image.Info, tag string) error {
	tagReference := strings.Join([]string{repoImage.Repository, tag}, ":")
	tagRepoImage, err := r.api.TryGetRepoImage(ctx, tagReference)
	if err != nil {
		return err
	}

	if tagRepoImage == nil {
		return fmt.Errorf("the custom tag %q not found", tag)
	}

	if repoImage.ID != tagRepoImage.ID {
		return fmt.Errorf("the custom tag %q image must be the same as related content-based tag %q image", tag, repoImage.Tag)
	}

	return nil
}

func (r *defaultImplementation) TagRepoImage(_ context.Context, repoImage *image.Info, tag string) error {
	return r.api.tagImage(strings.Join([]string{repoImage.Repository, repoImage.Tag}, ":"), tag)
}

func (r *defaultImplementation) DeleteRepoImage(_ context.Context, repoImage *image.Info) error {
	return r.api.deleteImageByReference(strings.Join([]string{repoImage.Repository, repoImage.RepoDigest}, "@"))
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
