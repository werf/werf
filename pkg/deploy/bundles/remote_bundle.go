package bundles

import (
	"context"
	"fmt"
	"strings"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"helm.sh/helm/v3/pkg/chart"
	"sigs.k8s.io/yaml"

	"github.com/werf/logboek"
	bundles_registry "github.com/werf/werf/pkg/deploy/bundles/registry"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/util"
)

type RemoteBundle struct {
	RegistryAddress       *RegistryAddress
	BundlesRegistryClient BundlesRegistryClient
	RegistryClient        docker_registry.Interface
}

func NewRemoteBundle(registryAddress *RegistryAddress, bundlesRegistryClient BundlesRegistryClient, registryClient docker_registry.Interface) *RemoteBundle {
	return &RemoteBundle{
		RegistryAddress:       registryAddress,
		BundlesRegistryClient: bundlesRegistryClient,
		RegistryClient:        registryClient,
	}
}

func (bundle *RemoteBundle) ReadChart(ctx context.Context) (*chart.Chart, error) {
	if err := logboek.Context(ctx).LogProcess("Pulling bundle %s", bundle.RegistryAddress.FullName()).DoError(func() error {
		if err := bundle.BundlesRegistryClient.PullChartToCache(bundle.RegistryAddress.Reference); err != nil {
			return fmt.Errorf("unable to pull bundle %s: %w", bundle.RegistryAddress.FullName(), err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	var ch *chart.Chart
	if err := logboek.Context(ctx).LogProcess("Loading bundle %s", bundle.RegistryAddress.FullName()).DoError(func() error {
		var err error
		ch, err = bundle.BundlesRegistryClient.LoadChart(bundle.RegistryAddress.Reference)
		if err != nil {
			return fmt.Errorf("unable to load pulled bundle %s: %w", bundle.RegistryAddress.FullName(), err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return ch, nil
}

func (bundle *RemoteBundle) WriteChart(ctx context.Context, ch *chart.Chart) error {
	if err := logboek.Context(ctx).LogProcess("Saving bundle %s", bundle.RegistryAddress.FullName()).DoError(func() error {
		if err := bundle.BundlesRegistryClient.SaveChart(ch, bundle.RegistryAddress.Reference); err != nil {
			return fmt.Errorf("unable to save bundle %s to the local chart helm cache: %w", bundle.RegistryAddress.FullName(), err)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := logboek.Context(ctx).LogProcess("Pushing bundle %s", bundle.RegistryAddress.FullName()).DoError(func() error {
		if err := bundle.BundlesRegistryClient.PushChart(bundle.RegistryAddress.Reference); err != nil {
			return fmt.Errorf("unable to push bundle %s: %w", bundle.RegistryAddress.FullName(), err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (bundle *RemoteBundle) CopyTo(ctx context.Context, to BundleAccessor, opts copyToOptions) error {
	return to.CopyFromRemote(ctx, bundle, opts)
}

func (bundle *RemoteBundle) CopyFromArchive(ctx context.Context, fromArchive *BundleArchive, opts copyToOptions) error {
	ch, err := fromArchive.ReadChart(ctx)
	if err != nil {
		return fmt.Errorf("unable to read chart from the bundle archive %q: %w", fromArchive.Reader.String(), err)
	}

	if err := logboek.Context(ctx).LogProcess("Copy images from bundle archive").DoError(func() error {
		if werfVals, ok := ch.Values["werf"].(map[string]interface{}); ok {
			if imageVals, ok := werfVals["image"].(map[string]interface{}); ok {
				newImageVals := make(map[string]interface{})

				for imageName, v := range imageVals {
					if imageRef, ok := v.(string); ok {
						ref, err := bundles_registry.ParseReference(imageRef)
						if err != nil {
							return fmt.Errorf("unable to parse bundle image %s: %w", imageRef, err)
						}
						ref.Repo = bundle.RegistryAddress.Repo

						if imageRef != ref.FullName() {
							logboek.Context(ctx).Default().LogFDetails("Image: %s\n", ref.FullName())

							imageArchiveOpener := fromArchive.GetImageArchiveOpener(ref.Tag)

							if err := bundle.RegistryClient.PushImageArchive(ctx, imageArchiveOpener, ref.FullName()); err != nil {
								return fmt.Errorf("error copying image from bundle archive %q into %q: %w", fromArchive.Reader.String(), ref.FullName(), err)
							}
						}

						newImageVals[imageName] = ref.FullName()
					} else {
						return fmt.Errorf("unexpected value .Values.werf.image.%s=%v", imageName, v)
					}
				}

				werfVals["image"] = newImageVals
			}

			werfVals["repo"] = bundle.RegistryAddress.Repo
		}

		return nil
	}); err != nil {
		return err
	}

	if err := SaveChartValues(ctx, ch); err != nil {
		return err
	}

	if nameOverwrite := GetChartNameOverwrite(bundle.RegistryAddress.Repo, opts.RenameChart, opts.HelmCompatibleChart); nameOverwrite != nil {
		ch.Metadata.Name = *nameOverwrite
	}

	sv, err := BundleTagToChartVersion(ctx, bundle.RegistryAddress.Tag, time.Now())
	if err != nil {
		return fmt.Errorf("unable to set chart version from bundle tag %q: %w", bundle.RegistryAddress.Tag, err)
	}
	ch.Metadata.Version = sv.String()

	if err := bundle.WriteChart(ctx, ch); err != nil {
		return fmt.Errorf("unable to write chart to remote bundle: %w", err)
	}

	return nil
}

func (bundle *RemoteBundle) CopyFromRemote(ctx context.Context, fromRemote *RemoteBundle, opts copyToOptions) error {
	ch, err := fromRemote.ReadChart(ctx)
	if err != nil {
		return fmt.Errorf("unable to read chart from source remote bundle: %w", err)
	}

	{
		d, err := yaml.Marshal(ch.Values)
		logboek.Context(ctx).Debug().LogF("Values before change (%v):\n%s\n---\n", err, d)
	}

	if err := logboek.Context(ctx).LogProcess("Copy images from remote bundle").DoError(func() error {
		if werfVals, ok := ch.Values["werf"].(map[string]interface{}); ok {
			if imageVals, ok := werfVals["image"].(map[string]interface{}); ok {
				newImageVals := make(map[string]interface{})

				for imageName, v := range imageVals {
					if image, ok := v.(string); ok {
						ref, err := bundles_registry.ParseReference(image)
						if err != nil {
							return fmt.Errorf("unable to parse bundle image %s: %w", image, err)
						}

						ref.Repo = bundle.RegistryAddress.Repo

						// TODO: copy images in parallel
						if image != ref.FullName() {
							logboek.Context(ctx).Default().LogFDetails("Source: %s\n", image)
							logboek.Context(ctx).Default().LogFDetails("Destination: %s\n", ref.FullName())

							if err := fromRemote.RegistryClient.MutateAndPushImage(ctx, image, ref.FullName(), func(cfg v1.Config) (v1.Config, error) { return cfg, nil }); err != nil {
								return fmt.Errorf("error copying image %s into %s: %w", image, ref.FullName(), err)
							}
						}

						newImageVals[imageName] = ref.FullName()
					} else {
						return fmt.Errorf("unexpected value .Values.werf.image.%s=%v", imageName, v)
					}
				}

				werfVals["image"] = newImageVals
			}

			werfVals["repo"] = bundle.RegistryAddress.Repo
		}
		return nil
	}); err != nil {
		return err
	}

	if err := SaveChartValues(ctx, ch); err != nil {
		return err
	}

	ch.Metadata.Name = util.Reverse(strings.SplitN(util.Reverse(bundle.RegistryAddress.Repo), "/", 2)[0])

	sv, err := BundleTagToChartVersion(ctx, bundle.RegistryAddress.Tag, time.Now())
	if err != nil {
		return fmt.Errorf("unable to set chart version from bundle tag %q: %w", bundle.RegistryAddress.Tag, err)
	}
	ch.Metadata.Version = sv.String()

	if err := bundle.WriteChart(ctx, ch); err != nil {
		return fmt.Errorf("unable to write chart to destination remote bundle: %w", err)
	}

	return nil
}
