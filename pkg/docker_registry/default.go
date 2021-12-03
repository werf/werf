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

func IsHarbor404Error(err error) bool {
	if err == nil {
		return false
	}

	// Example error:
	// GET https://domain/harbor/s3/object/name/prefix/docker/registry/v2/blobs/sha256/2d/3d8c68cd9df32f1beb4392298a123eac58aba1433a15b3258b2f3728bad4b7d1/data?X-Amz-Algorithm=REDACTED&X-Amz-Credential=REDACTED&X-Amz-Date=REDACTED&X-Amz-Expires=REDACTED&X-Amz-Signature=REDACTED&X-Amz-SignedHeaders=REDACTED: unsupported status code 404; body: <?xml version="1.0" encoding="UTF-8"?>
	// <Error><Code>NoSuchKey</Code><Message>The specified key does not exist.</Message><Resource>/harbor/s3/object/name/prefix/docker/registry/v2/blobs/sha256/3d/3d8c68cd9df32f1beb4392298a123eac58aba1433a15b3258b2f3728bad4b7d1/data</Resource><RequestId>c5bb943c-1e85-5930-b455-c3e8edbbaccd</RequestId></Error>

	return strings.Contains(err.Error(), "unsupported status code 404")
}
