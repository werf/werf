package docker_registry

import (
	"context"
	"fmt"
	"net/url"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"

	registry_api "github.com/werf/werf/v2/pkg/docker_registry/api"
	"github.com/werf/werf/v2/pkg/image"
)

type genericApi struct {
	commonApi *api
	mirrors   []string
}

func newGenericApi(_ context.Context, options apiOptions) (*genericApi, error) {
	d := &genericApi{}
	d.commonApi = newAPI(options)
	d.mirrors = options.RegistryMirrors
	return d, nil
}

func (api *genericApi) MutateAndPushImage(ctx context.Context, sourceReference, destinationReference string, opts ...registry_api.MutateOption) error {
	return api.commonApi.MutateAndPushImage(ctx, sourceReference, destinationReference, opts...)
}

func (api *genericApi) GetRepoImageConfigFile(ctx context.Context, reference string) (*v1.ConfigFile, error) {
	mirrorReferenceList, err := api.mirrorReferenceList(ctx, reference)
	if err != nil {
		return nil, fmt.Errorf("unable to prepare mirror reference list: %w", err)
	}

	for _, mirrorReference := range mirrorReferenceList {
		config, err := api.getRepoImageConfigFile(ctx, mirrorReference)
		if err != nil {
			if IsStatusNotFoundErr(err) || IsImageNotFoundError(err) || IsBrokenImageError(err) {
				continue
			}

			return nil, err
		}

		return config, nil
	}

	return api.getRepoImageConfigFile(ctx, reference)
}

func (api *genericApi) getRepoImageConfigFile(ctx context.Context, reference string) (*v1.ConfigFile, error) {
	desc, _, err := api.commonApi.getImageDesc(ctx, reference)
	if err != nil {
		return nil, err
	}

	img, err := desc.Image()
	if err != nil {
		return nil, err
	}

	return img.ConfigFile()
}

func (api *genericApi) GetRepoImage(ctx context.Context, reference string) (*image.Info, error) {
	mirrorReferenceList, err := api.mirrorReferenceList(ctx, reference)
	if err != nil {
		return nil, fmt.Errorf("unable to prepare mirror reference list: %w", err)
	}

	for _, mirrorReference := range mirrorReferenceList {
		info, err := api.commonApi.TryGetRepoImage(ctx, mirrorReference)
		if err != nil {
			return nil, fmt.Errorf("unable to try getting mirror repo image %q: %w", mirrorReference, err)
		}
		if info != nil {
			return info, nil
		}
	}

	return api.commonApi.GetRepoImage(ctx, reference)
}

func (api *genericApi) mirrorReferenceList(ctx context.Context, reference string) ([]string, error) {
	var referenceList []string

	referenceParts, err := api.commonApi.parseReferenceParts(reference)
	if err != nil {
		return nil, fmt.Errorf("unable to parse reference %q: %w", reference, err)
	}

	// nothing if container registry is not Docker Hub
	if referenceParts.registry != name.DefaultRegistry {
		return nil, nil
	}

	for _, mirrorRegistry := range api.mirrors {
		mirrorRegistryUrl, err := url.Parse(mirrorRegistry)
		if err != nil {
			return nil, fmt.Errorf("unable to parse mirror registry url %q: %w", mirrorRegistry, err)
		}

		mirrorReference := mirrorRegistryUrl.Host
		mirrorReference += "/" + referenceParts.repository
		mirrorReference += ":" + referenceParts.tag

		if referenceParts.digest != "" {
			mirrorReference += "@" + referenceParts.digest
		}

		referenceList = append(referenceList, mirrorReference)
	}

	return referenceList, nil
}
