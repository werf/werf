package bundles

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/samber/lo"
	"sigs.k8s.io/yaml"

	"github.com/werf/logboek"
	"github.com/werf/nelm/v2/pkg/chartyaml"
	nelmcommon "github.com/werf/nelm/v2/pkg/common"
	chartcommon "github.com/werf/nelm/v2/pkg/helm/pkg/chart/common"
	chart "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2"
	chartv2util "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2/util"
	"github.com/werf/werf/v3/pkg/docker_registry"
	bundles_registry "github.com/werf/werf/v3/pkg/ref"
)

var _ BundleAccessor = (*RemoteBundle)(nil)

type RemoteBundle struct {
	RegistryAddress       *bundles_registry.RegistryAddress
	BundlesRegistryClient BundlesRegistryClient
	RegistryClient        docker_registry.Interface
}

func NewRemoteBundle(registryAddress *bundles_registry.RegistryAddress, bundlesRegistryClient BundlesRegistryClient, registryClient docker_registry.Interface) *RemoteBundle {
	return &RemoteBundle{
		RegistryAddress:       registryAddress,
		BundlesRegistryClient: bundlesRegistryClient,
		RegistryClient:        registryClient,
	}
}

func (bundle *RemoteBundle) ReadChart(ctx context.Context, opts nelmcommon.HelmOptions) (*chart.Chart, error) {
	if err := logboek.Context(ctx).LogProcess("Pulling bundle %s", bundle.RegistryAddress.FullName()).DoError(func() error {
		if err := bundle.BundlesRegistryClient.PullChartToCache(ctx, bundle.RegistryAddress.Reference, opts); err != nil {
			return fmt.Errorf("unable to pull bundle %s: %w", bundle.RegistryAddress.FullName(), err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	var ch *chart.Chart
	if err := logboek.Context(ctx).LogProcess("Loading bundle %s", bundle.RegistryAddress.FullName()).DoError(func() error {
		var err error
		ch, err = bundle.BundlesRegistryClient.LoadChart(ctx, bundle.RegistryAddress.Reference, opts)
		if err != nil {
			return fmt.Errorf("unable to load pulled bundle %s: %w", bundle.RegistryAddress.FullName(), err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return ch, nil
}

func (bundle *RemoteBundle) WriteChart(ctx context.Context, ch *chart.Chart, opts nelmcommon.HelmOptions) error {
	tags := chartTags(ch.Metadata.Version, bundle.RegistryAddress.Tag)
	cacheRef := &bundles_registry.Reference{Repo: bundle.RegistryAddress.Repo, Tag: tags[0]}

	chartArchive, err := ChartToBytes(ch)
	if err != nil {
		return fmt.Errorf("unable to dump chart to archive: %w", err)
	}

	if err := logboek.Context(ctx).LogProcess("Saving bundle %s", cacheRef.FullName()).DoError(func() error {
		if err := bundle.BundlesRegistryClient.SaveChartArchive(ctx, chartArchive, nil, cacheRef, opts); err != nil {
			return fmt.Errorf("unable to save bundle %s to the local chart helm cache: %w", cacheRef.FullName(), err)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := logboek.Context(ctx).LogProcess("Pushing bundle %s with tags %s", bundle.RegistryAddress.Repo, strings.Join(tags, ", ")).DoError(func() error {
		if err := bundle.BundlesRegistryClient.PushChart(ctx, cacheRef, tags, opts); err != nil {
			return fmt.Errorf("unable to push bundle %s: %w", cacheRef.FullName(), err)
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
	ch, err := fromArchive.ReadChart(ctx, opts.HelmOptions)
	if err != nil {
		return fmt.Errorf("unable to read chart from the bundle archive %q: %w", fromArchive.Reader.String(), err)
	}

	if err := logboek.Context(ctx).LogProcess("Copy images from bundle archive").DoError(func() error {
		return retargetChart(ctx, ch, bundle.RegistryAddress.Repo, func(imageRef string, ref *bundles_registry.Reference) error {
			if imageRef == ref.FullName() {
				return nil
			}

			logboek.Context(ctx).Default().LogFDetails("Image: %s\n", ref.FullName())

			imageArchiveOpener := fromArchive.GetImageArchiveOpener(ref.Tag)

			if err := bundle.RegistryClient.PushImageArchive(ctx, imageArchiveOpener, ref.FullName()); err != nil {
				return fmt.Errorf("error copying image from bundle archive %q into %q: %w", fromArchive.Reader.String(), ref.FullName(), err)
			}

			return nil
		})
	}); err != nil {
		return err
	}

	if bundle.RegistryAddress.Tag != "" {
		sv, err := BundleTagToChartVersion(ctx, bundle.RegistryAddress.Tag, time.Now())
		if err != nil {
			return fmt.Errorf("unable to set chart version from bundle tag %q: %w", bundle.RegistryAddress.Tag, err)
		}

		if err := editChartMetadata(ctx, ch, chartyaml.EditOptions{Version: sv.String()}); err != nil {
			return err
		}
	}

	if err := bundle.WriteChart(ctx, ch, opts.HelmOptions); err != nil {
		return fmt.Errorf("unable to write chart to remote bundle: %w", err)
	}

	return nil
}

func (bundle *RemoteBundle) CopyFromRemote(ctx context.Context, fromRemote *RemoteBundle, opts copyToOptions) error {
	ch, err := fromRemote.ReadChart(ctx, opts.HelmOptions)
	if err != nil {
		return fmt.Errorf("unable to read chart from source remote bundle: %w", err)
	}

	sourceFingerprint, err := chartFingerprint(ch)
	if err != nil {
		return err
	}

	{
		d, err := yaml.Marshal(ch.Values)
		logboek.Context(ctx).Debug().LogF("Values before change (%v):\n%s\n---\n", err, d)
	}

	if err := logboek.Context(ctx).LogProcess("Copy images from remote bundle").DoError(func() error {
		return retargetChart(ctx, ch, bundle.RegistryAddress.Repo, func(imageRef string, ref *bundles_registry.Reference) error {
			// TODO: copy images in parallel
			if imageRef == ref.FullName() {
				return nil
			}

			logboek.Context(ctx).Default().LogFDetails("Source: %s\n", imageRef)
			logboek.Context(ctx).Default().LogFDetails("Destination: %s\n", ref.FullName())

			if err := fromRemote.RegistryClient.CopyImage(ctx, imageRef, ref.FullName(), docker_registry.CopyImageOptions{}); err != nil {
				return fmt.Errorf("error copying image %s into %s: %w", imageRef, ref.FullName(), err)
			}

			return nil
		})
	}); err != nil {
		return err
	}

	destination := bundle
	if bundle.RegistryAddress.Tag == "" {
		destination = NewRemoteBundle(&bundles_registry.RegistryAddress{
			Reference: &bundles_registry.Reference{Repo: bundle.RegistryAddress.Repo, Tag: fromRemote.RegistryAddress.Tag},
		}, bundle.BundlesRegistryClient, bundle.RegistryClient)
	} else {
		sv, err := BundleTagToChartVersion(ctx, bundle.RegistryAddress.Tag, time.Now())
		if err != nil {
			return fmt.Errorf("unable to set chart version from bundle tag %q: %w", bundle.RegistryAddress.Tag, err)
		}

		if err := editChartMetadata(ctx, ch, chartyaml.EditOptions{Version: sv.String()}); err != nil {
			return err
		}
	}

	fingerprint, err := chartFingerprint(ch)
	if err != nil {
		return err
	}

	if bytes.Equal(fingerprint, sourceFingerprint) {
		if err := destination.writeCachedChart(ctx, fromRemote.RegistryAddress.Reference, ch.Metadata.Version, opts.HelmOptions); err != nil {
			return fmt.Errorf("unable to write unchanged chart to destination remote bundle: %w", err)
		}

		return nil
	}

	if err := warnIfProvenanceDropped(ctx, fromRemote, "the copy changes the chart", opts.HelmOptions); err != nil {
		return err
	}

	if err := destination.WriteChart(ctx, ch, opts.HelmOptions); err != nil {
		return fmt.Errorf("unable to write chart to destination remote bundle: %w", err)
	}

	return nil
}

// retargetChart copies the werf images of the chart into repo with copyImage, given the source
// image ref and its ref in repo, and points the image values, the helm.sh/images entries of the
// werf images and the chart name at repo.
func retargetChart(ctx context.Context, ch *chart.Chart, repo string, copyImage func(imageRef string, ref *bundles_registry.Reference) error) error {
	vals, werfData, err := chartImageValues(ch)
	if err != nil {
		return err
	}

	sourceValsJSON, err := json.Marshal(vals)
	if err != nil {
		return fmt.Errorf("marshal image values: %w", err)
	}

	newImageRefs := make(map[string]string)

	if werfVals, ok := vals["werf"].(map[string]interface{}); ok {
		if imageVals, ok := werfVals["image"].(map[string]interface{}); ok {
			newImageVals := make(map[string]interface{})

			for imageName, v := range imageVals {
				imageRef, ok := v.(string)
				if !ok {
					return fmt.Errorf("unexpected value .Values.werf.image.%s=%v", imageName, v)
				}

				ref, err := bundles_registry.ParseReference(imageRef)
				if err != nil {
					return err
				}

				ref.Repo = repo

				if err := copyImage(imageRef, ref); err != nil {
					return err
				}

				newImageVals[imageName] = ref.FullName()
				newImageRefs[imageName] = ref.FullName()
			}

			werfVals["image"] = newImageVals
		}

		werfVals["repo"] = repo
	}

	if err := updateGlobalWerfValues(vals, repo, newImageRefs); err != nil {
		return err
	}

	if err := saveChartImageValues(ctx, ch, vals, werfData, sourceValsJSON); err != nil {
		return err
	}

	return editChartMetadata(ctx, ch, chartyaml.EditOptions{Name: chartNameFromRepo(repo)})
}

// chartFingerprint returns what a copy can change in the chart.
func chartFingerprint(ch *chart.Chart) ([]byte, error) {
	chartYAML := lo.Filter(ch.Raw, func(f *chartcommon.File, _ int) bool { return f.Name == chartv2util.ChartfileName })

	fingerprint, err := json.Marshal([]interface{}{ch.Metadata.Name, ch.Metadata.Version, ch.Values, ch.WerfDataFiles, chartYAML})
	if err != nil {
		return nil, fmt.Errorf("marshal chart fingerprint: %w", err)
	}

	return fingerprint, nil
}

func (bundle *RemoteBundle) writeCachedChart(ctx context.Context, from *bundles_registry.Reference, chartVersion string, opts nelmcommon.HelmOptions) error {
	tags := chartTags(chartVersion, bundle.RegistryAddress.Tag)
	cacheRef := &bundles_registry.Reference{Repo: bundle.RegistryAddress.Repo, Tag: tags[0]}

	if err := bundle.BundlesRegistryClient.TagCachedChart(ctx, from, cacheRef, opts); err != nil {
		return fmt.Errorf("unable to save bundle %s to the local chart helm cache: %w", cacheRef.FullName(), err)
	}

	if err := logboek.Context(ctx).LogProcess("Pushing bundle %s with tags %s", bundle.RegistryAddress.Repo, strings.Join(tags, ", ")).DoError(func() error {
		if err := bundle.BundlesRegistryClient.PushChart(ctx, cacheRef, tags, opts); err != nil {
			return fmt.Errorf("unable to push bundle %s: %w", cacheRef.FullName(), err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func warnIfProvenanceDropped(ctx context.Context, fromRemote *RemoteBundle, reason string, opts nelmcommon.HelmOptions) error {
	signed, err := fromRemote.BundlesRegistryClient.IsChartSigned(ctx, fromRemote.RegistryAddress.Reference, opts)
	if err != nil {
		return fmt.Errorf("check whether bundle %s is signed: %w", fromRemote.RegistryAddress.FullName(), err)
	}

	if signed {
		logboek.Context(ctx).Warn().LogF("Bundle %s is signed, but its copy is not: %s\n", fromRemote.RegistryAddress.FullName(), reason)
	}

	return nil
}
