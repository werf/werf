package bundles

import (
	"context"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/werf/logboek"
	"github.com/werf/nelm/v2/pkg/action"
	"github.com/werf/nelm/v2/pkg/helm/pkg/registry"
	"github.com/werf/werf/v3/pkg/ref"
)

type PublishOptions struct {
	// ChartPackOptions are the options to pack the chart with. The chart name and version are set
	// from the repo and the tag.
	ChartPackOptions action.ChartPackOptions
}

// Publish packs the chart and pushes it into repo under the chart version derived from tag and under
// tag itself, both with "+" encoded as "_" as Helm does.
func Publish(ctx context.Context, chartDir, repo, tag string, registryClient *registry.Client, opts PublishOptions) error {
	repo, err := normalizeRepo(repo)
	if err != nil {
		return err
	}

	packOpts := opts.ChartPackOptions
	packOpts.Name = path.Base(repo)
	packOpts.Version = action.ChartVersionFromTag(ctx, tag, time.Now())

	tags := chartTags(packOpts.Version, tag)

	r, err := ref.ParseReference(fmt.Sprintf("%s:%s", repo, tags[0]))
	if err != nil {
		return fmt.Errorf("error parsing bundle ref: %w", err)
	}

	if err := checkTagIsNotProvenance(tag); err != nil {
		return err
	}

	if err := checkChartTagsDoNotOverwriteImages(r.Repo, tags, packOpts.ExtraValues); err != nil {
		return err
	}

	var packResult *action.ChartPackResult
	if err := logboek.Context(ctx).Default().LogProcess("Packing bundle").DoError(func() error {
		packResult, err = action.ChartPack(ctx, chartDir, packOpts)
		if err != nil {
			return fmt.Errorf("pack chart %q: %w", chartDir, err)
		}

		return nil
	}); err != nil {
		return err
	}

	chartArchive, err := os.ReadFile(packResult.ArchivePath)
	if err != nil {
		return fmt.Errorf("read chart archive: %w", err)
	}

	var provenance []byte
	if packResult.ProvenancePath != "" {
		provenance, err = os.ReadFile(packResult.ProvenancePath)
		if err != nil {
			return fmt.Errorf("read chart provenance file: %w", err)
		}
	}

	var pushResult *registry.PushResult
	if err := logboek.Context(ctx).LogProcess("Pushing bundle %s with tags %s", repo, strings.Join(tags, ", ")).DoError(func() error {
		if pushResult, err = registryClient.PushContext(ctx, chartArchive, fmt.Sprintf("%s:%s", repo, packResult.ChartVersion), registry.PushOptProvData(provenance), registry.PushOptExtraTags(tags[1:])); err != nil {
			return fmt.Errorf("push chart: %w", err)
		}

		return nil
	}); err != nil {
		return err
	}

	logboek.Context(ctx).Default().LogF("Pushed bundle %s:%s (digest %s)\n", repo, tags[len(tags)-1], pushResult.Manifest.Digest)

	return nil
}

// checkChartTagsDoNotOverwriteImages fails if a chart tag equals the tag of a werf image from the
// service values in the same repository, since pushing the chart would replace the image.
func checkChartTagsDoNotOverwriteImages(repo string, tags []string, serviceValues map[string]interface{}) error {
	imageRefs := map[string]string{}

	werfVals, _ := serviceValues["werf"].(map[string]interface{})
	legacyImageVals, _ := werfVals["image"].(map[string]interface{})
	for imageName, val := range legacyImageVals {
		if imageRef, ok := val.(string); ok {
			imageRefs[imageName] = imageRef
		}
	}

	globalVals, _ := serviceValues["global"].(map[string]interface{})
	globalWerfVals, _ := globalVals["werf"].(map[string]interface{})
	imagesVals, _ := globalWerfVals["images"].(map[string]interface{})
	for imageName, val := range imagesVals {
		imageVals, _ := val.(map[string]interface{})
		if imageRef, ok := imageVals["ref_tag"].(string); ok {
			imageRefs[imageName] = imageRef
		}
	}

	imageNames := lo.Keys(imageRefs)
	sort.Strings(imageNames)

	for _, imageName := range imageNames {
		imageRef, err := ref.ParseReference(imageRefs[imageName])
		if err != nil || imageRef.Repo != repo || !lo.Contains(tags, imageRef.Tag) {
			continue
		}

		return fmt.Errorf("bundle tag %q would overwrite the image %q %s in the same repository: publish the bundle with another --tag or use another image tag", imageRef.Tag, imageName, imageRef.FullName())
	}

	return nil
}

// chartTags returns the tags to push a chart with chartVersion under: the version itself and the tag the user asked for, if it differs, both with "+", invalid in a tag, encoded as Helm does.
func chartTags(chartVersion, tag string) []string {
	versionTag := encodeTag(chartVersion)
	tag = encodeTag(tag)
	if tag == "" || tag == versionTag {
		return []string{versionTag}
	}

	return []string{versionTag, tag}
}

func checkTagIsNotProvenance(tag string) error {
	if strings.HasSuffix(tag, ".prov") {
		return fmt.Errorf("tag %q is not allowed: tags ending in \".prov\" address provenance files", tag)
	}

	return nil
}

func encodeTag(tag string) string {
	return strings.ReplaceAll(tag, "+", "_")
}
