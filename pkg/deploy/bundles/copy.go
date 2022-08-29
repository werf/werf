package bundles

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/otiai10/copy"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"sigs.k8s.io/yaml"

	"github.com/werf/logboek"
	bundles_registry "github.com/werf/werf/pkg/deploy/bundles/registry"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/util"
)

type CopyOptions struct {
	BundlesRegistryClient    *bundles_registry.Client
	FromRegistry, ToRegistry docker_registry.Interface
}

func Copy(ctx context.Context, fromAddr, toAddr *Addr, opts CopyOptions) error {
	switch {
	case fromAddr.RegistryAddress != nil && toAddr.ArchiveAddress != nil:
		return copyFromRegistryToArchive(ctx, fromAddr.RegistryAddress, toAddr.ArchiveAddress, opts.BundlesRegistryClient, opts.FromRegistry)
	case fromAddr.ArchiveAddress != nil && toAddr.RegistryAddress != nil:
		return copyFromArchiveToRegistry(ctx, fromAddr.ArchiveAddress, toAddr.RegistryAddress, opts.BundlesRegistryClient, opts.ToRegistry)
	case fromAddr.RegistryAddress != nil && toAddr.RegistryAddress != nil:
		return copyFromRegistryToRegistry(ctx, fromAddr.RegistryAddress, toAddr.RegistryAddress, opts.BundlesRegistryClient, opts.FromRegistry, opts.ToRegistry)
	case fromAddr.ArchiveAddress != nil && toAddr.ArchiveAddress != nil:
		return copyFromArchiveToArchive(ctx, fromAddr.ArchiveAddress, toAddr.ArchiveAddress)
	default:
		panic(fmt.Sprintf("unexpected from %v and to %v", fromAddr, toAddr))
	}
}

func copyFromArchiveToArchive(ctx context.Context, from, to *ArchiveAddress) error {
	logboek.Context(ctx).Debug().LogF("-- copyFromArchiveToArchive\n")
	if err := copy.Copy(from.Path, to.Path); err != nil {
		return err
	}
	return nil
}

func copyFromArchiveToRegistry(ctx context.Context, from *ArchiveAddress, to *RegistryAddress, bundlesRegistryClient *bundles_registry.Client, toRegistry docker_registry.Interface) error {
	logboek.Context(ctx).Debug().LogF("-- copyFromArchiveToRegistry\n")

	bundleArchive := NewBundleArchive(from.Path)

	chartBytes, err := bundleArchive.ReadChartArchive()
	if err != nil {
		return fmt.Errorf("unable to read chart from the bundle archive at %q: %w", bundleArchive.Path, err)
	}

	ch, err := BytesToChart(chartBytes)
	if err != nil {
		return fmt.Errorf("unable to read chart from the bundle archive %q: %w", bundleArchive.Path, err)
	}

	if werfVals, ok := ch.Values["werf"].(map[string]interface{}); ok {
		if imageVals, ok := werfVals["image"].(map[string]interface{}); ok {
			newImageVals := make(map[string]interface{})

			for imageName, v := range imageVals {
				if imageRef, ok := v.(string); ok {
					ref, err := bundles_registry.ParseReference(imageRef)
					if err != nil {
						return fmt.Errorf("unable to parse bundle image %s: %w", imageRef, err)
					}
					ref.Repo = to.Repo

					if imageRef != ref.FullName() {
						if err := logboek.Context(ctx).LogProcess("Copy image from bundle archive").DoError(func() error {
							logboek.Context(ctx).Default().LogFDetails("Destination: %s\n", ref.FullName())

							imageArchiveOpener := bundleArchive.GetImageArchiveOpener(ref.Tag)

							if err := toRegistry.PushImageArchive(ctx, imageArchiveOpener, ref.FullName()); err != nil {
								return fmt.Errorf("error copying image from bundle archive %q into %q: %w", bundleArchive.Path, ref.FullName(), err)
							}

							return nil
						}); err != nil {
							return err
						}
					}

					newImageVals[imageName] = ref.FullName()
				} else {
					return fmt.Errorf("unexpected value .Values.werf.image.%s=%v", imageName, v)
				}
			}

			werfVals["image"] = newImageVals
		}

		werfVals["repo"] = to.Repo
	}

	ch.Metadata.Name = util.Reverse(strings.SplitN(util.Reverse(to.Repo), "/", 2)[0])

	if err := logboek.Context(ctx).LogProcess("Saving bundle %s", to.FullName()).DoError(func() error {
		if err := bundlesRegistryClient.SaveChart(ch, to.Reference); err != nil {
			return fmt.Errorf("unable to save bundle %s to the local chart helm cache: %w", to.FullName(), err)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := logboek.Context(ctx).LogProcess("Pushing bundle %s", to.FullName()).DoError(func() error {
		if err := bundlesRegistryClient.PushChart(to.Reference); err != nil {
			return fmt.Errorf("unable to push bundle %s: %w", to.FullName(), err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func copyFromRegistryToArchive(ctx context.Context, from *RegistryAddress, to *ArchiveAddress, bundlesRegistryClient *bundles_registry.Client, fromRegistry docker_registry.Interface) error {
	logboek.Context(ctx).Debug().LogF("-- copyFromRegistryToArchive\n")

	if err := logboek.Context(ctx).LogProcess("Pulling bundle %s", from.FullName()).DoError(func() error {
		if err := bundlesRegistryClient.PullChartToCache(from.Reference); err != nil {
			return fmt.Errorf("unable to pull bundle %s: %w", from.FullName(), err)
		}
		return nil
	}); err != nil {
		return err
	}

	var ch *chart.Chart
	if err := logboek.Context(ctx).LogProcess("Loading bundle %s", from.FullName()).DoError(func() error {
		var err error
		ch, err = bundlesRegistryClient.LoadChart(from.Reference)
		if err != nil {
			return fmt.Errorf("unable to load pulled bundle %s: %w", from.FullName(), err)
		}
		return nil
	}); err != nil {
		return err
	}

	b := NewBundleArchive(to.Path)

	if err := b.Open(); err != nil {
		return fmt.Errorf("unable to open target bundle archive: %w", err)
	}

	if err := logboek.Context(ctx).LogProcess("Saving bundle %s into archive", from.FullName()).DoError(func() error {
		chartBytes, err := ChartToBytes(ch)
		if err != nil {
			return fmt.Errorf("uanble to dump chart to bytes: %w", err)
		}

		if err := b.WriteChartArchive(chartBytes); err != nil {
			return fmt.Errorf("unable to write chart archive into bundle archive: %w", err)
		}

		return nil
	}); err != nil {
		return err
	}

	if werfVals, ok := ch.Values["werf"].(map[string]interface{}); ok {
		if imageVals, ok := werfVals["image"].(map[string]interface{}); ok {
			for imageName, v := range imageVals {
				if imageRef, ok := v.(string); ok {
					logboek.Context(ctx).Default().LogFDetails("Saving image %s\n", imageRef)

					_, tag := image.ParseRepositoryAndTag(imageRef)

					// TODO: maybe save into tmp file archive OR read resulting image size from the registry before pulling
					imageBytes := bytes.NewBuffer(nil)
					zipper := gzip.NewWriter(imageBytes)

					if err := fromRegistry.PullImageArchive(ctx, zipper, imageRef); err != nil {
						return fmt.Errorf("error pulling image %q archive: %w", imageRef, err)
					}

					if err := zipper.Close(); err != nil {
						return fmt.Errorf("unable to close gzip writer: %w", err)
					}

					if err := b.WriteImageArchive(tag, imageBytes.Bytes()); err != nil {
						return fmt.Errorf("error writing image %q into bundle archive: %w", imageRef, err)
					}
				} else {
					return fmt.Errorf("unexpected value .Values.werf.image.%s=%v", imageName, v)
				}
			}
		}
	}

	if err := b.Save(); err != nil {
		return fmt.Errorf("error saving destination bundle archive: %w", err)
	}

	return nil
}

func copyFromRegistryToRegistry(ctx context.Context, from, to *RegistryAddress, bundlesRegistryClient *bundles_registry.Client, fromRegistry, toRegistry docker_registry.Interface) error {
	logboek.Context(ctx).Debug().LogF("-- copyFromRegistryToRegistry\n")

	if err := logboek.Context(ctx).LogProcess("Pulling bundle %s", from.FullName()).DoError(func() error {
		if err := bundlesRegistryClient.PullChartToCache(from.Reference); err != nil {
			return fmt.Errorf("unable to pull bundle %s: %w", from.FullName(), err)
		}
		return nil
	}); err != nil {
		return err
	}

	var ch *chart.Chart
	if err := logboek.Context(ctx).LogProcess("Loading bundle %s", from.FullName()).DoError(func() error {
		var err error
		ch, err = bundlesRegistryClient.LoadChart(from.Reference)
		if err != nil {
			return fmt.Errorf("unable to load pulled bundle %s: %w", from.FullName(), err)
		}
		return nil
	}); err != nil {
		return err
	}

	{
		d, err := yaml.Marshal(ch.Values)
		logboek.Context(ctx).Debug().LogF("Values before change (%v):\n%s\n---\n", err, d)
	}

	if werfVals, ok := ch.Values["werf"].(map[string]interface{}); ok {
		if imageVals, ok := werfVals["image"].(map[string]interface{}); ok {
			newImageVals := make(map[string]interface{})

			for imageName, v := range imageVals {
				if image, ok := v.(string); ok {
					ref, err := bundles_registry.ParseReference(image)
					if err != nil {
						return fmt.Errorf("unable to parse bundle image %s: %w", image, err)
					}

					ref.Repo = to.Repo

					// TODO: copy images in parallel
					if image != ref.FullName() {
						if err := logboek.Context(ctx).LogProcess("Copy image").DoError(func() error {
							logboek.Context(ctx).Default().LogFDetails("Source: %s\n", image)
							logboek.Context(ctx).Default().LogFDetails("Destination: %s\n", ref.FullName())

							if err := fromRegistry.MutateAndPushImage(ctx, image, ref.FullName(), func(cfg v1.Config) (v1.Config, error) { return cfg, nil }); err != nil {
								return fmt.Errorf("error copying image %s into %s: %w", image, ref.FullName(), err)
							}
							return nil
						}); err != nil {
							return err
						}
					}

					newImageVals[imageName] = ref.FullName()
				} else {
					return fmt.Errorf("unexpected value .Values.werf.image.%s=%v", imageName, v)
				}
			}

			werfVals["image"] = newImageVals
		}

		werfVals["repo"] = to.Repo
	}

	valuesRaw, err := yaml.Marshal(ch.Values)
	if err != nil {
		return fmt.Errorf("unable to marshal changed bundle values: %w", err)
	}
	logboek.Context(ctx).Debug().LogF("Values after change (%v):\n%s\n---\n", err, valuesRaw)

	for _, f := range ch.Raw {
		if f.Name == chartutil.ValuesfileName {
			f.Data = valuesRaw
			break
		}
	}

	ch.Metadata.Name = util.Reverse(strings.SplitN(util.Reverse(to.Repo), "/", 2)[0])

	if err := logboek.Context(ctx).LogProcess("Saving bundle %s", to.FullName()).DoError(func() error {
		if err := bundlesRegistryClient.SaveChart(ch, to.Reference); err != nil {
			return fmt.Errorf("unable to save bundle %s to the local chart helm cache: %w", to.FullName(), err)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := logboek.Context(ctx).LogProcess("Pushing bundle %s", to.FullName()).DoError(func() error {
		if err := bundlesRegistryClient.PushChart(to.Reference); err != nil {
			return fmt.Errorf("unable to push bundle %s: %w", to.FullName(), err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}
