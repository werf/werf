package bundles

import (
	"context"
	"fmt"
	"sort"

	"github.com/samber/lo"

	"github.com/werf/nelm/v2/pkg/action"
	"github.com/werf/nelm/v2/pkg/chartyaml"
	nelmcommon "github.com/werf/nelm/v2/pkg/common"
	v3chart "github.com/werf/nelm/v2/pkg/helm/intern/chart/v3"
	helmchart "github.com/werf/nelm/v2/pkg/helm/pkg/chart"
	chartcommon "github.com/werf/nelm/v2/pkg/helm/pkg/chart/common"
	chart "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2"
	"github.com/werf/nelm/v2/pkg/werfdata"
	"github.com/werf/werf/v3/pkg/ref"
)

type CopyOptions struct {
	ChartRepoConnectionOptions nelmcommon.ChartRepoConnectionOptions
	RegistryCredentialsPath    string
	TempDirPath                string
}

// Copy copies the bundle at from, with its images, to to, both either a registry address
// "[docker://]REPO[:TAG]" or an archive "archive:PATH". The image values of a copy to a registry
// are pointed at the new repository.
func Copy(ctx context.Context, from, to string, opts CopyOptions) error {
	if err := action.ChartCopy(ctx, from, to, action.ChartCopyOptions{
		ChartRepoConnectionOptions: opts.ChartRepoConnectionOptions,
		ImagesFunc:                 bundleImages,
		RegistryCredentialsPath:    opts.RegistryCredentialsPath,
		RewriteChart:               rewriteImageValues,
		TempDirPath:                opts.TempDirPath,
	}); err != nil {
		return fmt.Errorf("copy chart: %w", err)
	}

	return nil
}

// bundleImages returns the werf images of a bundle published by werf <= 3.12, which lists them
// only in its werf.image values. A chart with .werf-data lists them in helm.sh/images.
func bundleImages(ctx context.Context, ch helmchart.Charter) ([]chartyaml.Image, error) {
	accessor, err := helmchart.NewAccessor(ch)
	if err != nil {
		return nil, fmt.Errorf("create chart accessor: %w", err)
	}

	if len(accessor.WerfDataFiles()) > 0 {
		return nil, nil
	}

	werfVals, _ := accessor.Values()["werf"].(map[string]interface{})
	imageVals, _ := werfVals["image"].(map[string]interface{})

	names := lo.Keys(imageVals)
	sort.Strings(names)

	images := make([]chartyaml.Image, 0, len(names))
	for _, name := range names {
		imageRef, ok := imageVals[name].(string)
		if !ok {
			return nil, fmt.Errorf("unexpected value .Values.werf.image.%s=%v", name, imageVals[name])
		}

		images = append(images, chartyaml.Image{Name: name, Image: imageRef})
	}

	return images, nil
}

// rewriteImageValues points the werf image and repo values at repo, where the images are copied
// keeping their tags: .werf-data/extra-values.yaml of a chart with .werf-data, the chart values of a
// bundle published by werf <= 3.12.
func rewriteImageValues(ctx context.Context, ch helmchart.Charter, repo string, _ map[string]string) error {
	accessor, err := helmchart.NewAccessor(ch)
	if err != nil {
		return fmt.Errorf("create chart accessor: %w", err)
	}

	if len(accessor.WerfDataFiles()) == 0 {
		return retargetImageValues(accessor.Values(), repo)
	}

	data, err := werfdata.Parse(accessor.WerfDataFiles())
	if err != nil {
		return fmt.Errorf("parse %s: %w", werfdata.DirName, err)
	}

	if err := retargetImageValues(data.ExtraValues, repo); err != nil {
		return err
	}

	files, err := data.ChartFiles()
	if err != nil {
		return fmt.Errorf("build %s files: %w", werfdata.DirName, err)
	}

	return setWerfDataFiles(ch, files)
}

func retargetImageValues(vals map[string]interface{}, repo string) error {
	newImageRefs := make(map[string]string)

	if werfVals, ok := vals["werf"].(map[string]interface{}); ok {
		if imageVals, ok := werfVals["image"].(map[string]interface{}); ok {
			for imageName, v := range imageVals {
				imageRef, ok := v.(string)
				if !ok {
					return fmt.Errorf("unexpected value .Values.werf.image.%s=%v", imageName, v)
				}

				imageReference, err := ref.ParseReference(imageRef)
				if err != nil {
					return fmt.Errorf("parse image reference %q: %w", imageRef, err)
				}

				imageReference.Repo = repo

				imageVals[imageName] = imageReference.FullName()
				newImageRefs[imageName] = imageReference.FullName()
			}
		}

		werfVals["repo"] = repo
	}

	return updateGlobalWerfValues(vals, repo, newImageRefs)
}

func setWerfDataFiles(ch helmchart.Charter, files []*chartcommon.File) error {
	switch c := ch.(type) {
	case *chart.Chart:
		c.WerfDataFiles = files
	case *v3chart.Chart:
		c.WerfDataFiles = files
	default:
		return fmt.Errorf("unsupported chart type %T", ch)
	}

	return nil
}
