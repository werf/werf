package bundles

import (
	"context"
	"fmt"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"sigs.k8s.io/yaml"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/deploy/bundles/registry"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/util"
)

func Copy(ctx context.Context, fromAddr, toAddr string, bundlesRegistryClient *registry.Client, fromRegistry docker_registry.Interface) error {
	fromRef, err := registry.ParseReference(fromAddr)
	if err != nil {
		return fmt.Errorf("unable to parse source address %q: %w", fromAddr, err)
	}
	toRef, err := registry.ParseReference(toAddr)
	if err != nil {
		return fmt.Errorf("unable to parse destination address %q: %w", toAddr, err)
	}

	if err := logboek.Context(ctx).LogProcess("Pulling bundle %s", fromRef.FullName()).DoError(func() error {
		if err := bundlesRegistryClient.PullChartToCache(fromRef); err != nil {
			return fmt.Errorf("unable to pull bundle %s: %w", fromRef.FullName(), err)
		}
		return nil
	}); err != nil {
		return err
	}

	var ch *chart.Chart
	if err := logboek.Context(ctx).LogProcess("Loading bundle %s", fromRef.FullName()).DoError(func() error {
		ch, err = bundlesRegistryClient.LoadChart(fromRef)
		if err != nil {
			return fmt.Errorf("unable to load pulled bundle %s: %w", fromRef.FullName(), err)
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
					ref, err := registry.ParseReference(image)
					if err != nil {
						return fmt.Errorf("unable to parse bundle image %s: %w", image, err)
					}

					ref.Repo = toRef.Repo

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

		werfVals["repo"] = toRef.Repo
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

	ch.Metadata.Name = util.Reverse(strings.SplitN(util.Reverse(toRef.Repo), "/", 2)[0])

	if err := logboek.Context(ctx).LogProcess("Saving bundle %s", toRef.FullName()).DoError(func() error {
		if err := bundlesRegistryClient.SaveChart(ch, toRef); err != nil {
			return fmt.Errorf("unable to save bundle %s to the local chart helm cache: %w", toRef.FullName(), err)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := logboek.Context(ctx).LogProcess("Pushing bundle %s", toRef.FullName()).DoError(func() error {
		if err := bundlesRegistryClient.PushChart(toRef); err != nil {
			return fmt.Errorf("unable to push bundle %s: %w", toRef.FullName(), err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}
