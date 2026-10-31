package bundles

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/samber/lo"
	"sigs.k8s.io/yaml"

	"github.com/werf/nelm/v2/pkg/chartyaml"
	chart "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2"
	chartv2util "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2/util"
	"github.com/werf/nelm/v2/pkg/werfdata"
)

// ReadWerfDataMetadata returns the .werf-data metadata of the chart directory. The returned bool is
// false, and the metadata empty, for a chart without .werf-data, e.g. a bundle published by werf <= 3.12.
func ReadWerfDataMetadata(ctx context.Context, chartDir string) (*werfdata.Metadata, bool, error) {
	metadataPath := filepath.Join(chartDir, werfdata.DirName, werfdata.MetadataFileName)

	data, err := os.ReadFile(metadataPath)
	if errors.Is(err, fs.ErrNotExist) {
		return &werfdata.Metadata{}, false, nil
	} else if err != nil {
		return nil, false, fmt.Errorf("read %q: %w", metadataPath, err)
	}

	// A newer apiVersion may add fields, so check it before parsing strictly.
	var versioned struct {
		APIVersion string `json:"apiVersion"`
	}
	if err := yaml.Unmarshal(data, &versioned); err != nil {
		return nil, false, fmt.Errorf("parse %q: %w", metadataPath, err)
	}

	if err := werfdata.CheckAPIVersion(versioned.APIVersion); err != nil {
		return nil, false, err
	}

	metadata := &werfdata.Metadata{}
	if err := yaml.UnmarshalStrict(data, metadata); err != nil {
		return nil, false, fmt.Errorf("parse %q: %w", metadataPath, err)
	}

	return metadata, true, nil
}

// ReleaseDefaults returns the release name and namespace that werf bundle publish resolved, or empty
// ones when env is set and differs from the environment they were resolved for, as converge would
// resolve them differently.
func ReleaseDefaults(ctx context.Context, metadata *werfdata.Metadata, env string) werfdata.ReleaseMetadata {
	if metadata.Release == nil || (env != "" && env != metadata.Release.Env) {
		return werfdata.ReleaseMetadata{}
	}

	return *metadata.Release
}

// ChartImagesFromServiceValues returns the helm.sh/images entries of the werf images in the service
// values.
func ChartImagesFromServiceValues(ctx context.Context, serviceValues map[string]interface{}) []chartyaml.Image {
	globalVals, _ := serviceValues["global"].(map[string]interface{})
	werfVals, _ := globalVals["werf"].(map[string]interface{})
	imagesVals, _ := werfVals["images"].(map[string]interface{})

	names := lo.Keys(imagesVals)
	sort.Strings(names)

	images := []chartyaml.Image{}
	for _, name := range names {
		imageVals, ok := imagesVals[name].(map[string]interface{})
		if !ok {
			continue
		}

		imageRef, _ := imageVals["ref"].(string)
		if digest, _ := imageVals["digest"].(string); digest == "" {
			imageRef, _ = imageVals["ref_tag"].(string)
		}

		images = append(images, chartyaml.Image{Name: name, Image: imageRef})
	}

	return images
}

// chartImageValues returns the values with the refs of the werf images: .werf-data/extra-values.yaml
// of a chart with .werf-data, the chart values of a bundle published by werf <= 3.12. The
// .werf-data is nil for the latter.
func chartImageValues(ch *chart.Chart) (map[string]interface{}, *werfdata.Data, error) {
	if len(ch.WerfDataFiles) == 0 {
		return ch.Values, nil, nil
	}

	data, err := werfdata.Parse(ch.WerfDataFiles)
	if err != nil {
		return nil, nil, fmt.Errorf("parse %s: %w", werfdata.DirName, err)
	}

	return data.ExtraValues, data, nil
}

// saveChartImageValues writes the image values returned by chartImageValues back into the chart.
// For a chart with .werf-data the helm.sh/images entries of the werf images are rewritten too,
// unless the values are unchanged.
func saveChartImageValues(ctx context.Context, ch *chart.Chart, vals map[string]interface{}, data *werfdata.Data, sourceValsJSON []byte) error {
	if data == nil {
		return SaveChartValues(ctx, ch)
	}

	valsJSON, err := json.Marshal(vals)
	if err != nil {
		return fmt.Errorf("marshal image values: %w", err)
	}

	if string(valsJSON) == string(sourceValsJSON) {
		return nil
	}

	images := ChartImagesFromServiceValues(ctx, vals)

	imagesAnnotation, err := chartyaml.UpdateImages(ctx, ch.Metadata.Annotations[chartyaml.ImagesAnnotation], data.Metadata.Images, images)
	if err != nil {
		return fmt.Errorf("update %s annotation: %w", chartyaml.ImagesAnnotation, err)
	}

	data.ExtraValues = vals
	data.Metadata.Images = lo.Map(images, func(image chartyaml.Image, _ int) string { return image.Name })

	ch.WerfDataFiles, err = data.ChartFiles()
	if err != nil {
		return fmt.Errorf("build %s files: %w", werfdata.DirName, err)
	}

	edit := chartyaml.EditOptions{}
	if imagesAnnotation == "" {
		edit.DeleteAnnotations = []string{chartyaml.ImagesAnnotation}
	} else {
		edit.SetAnnotations = map[string]string{chartyaml.ImagesAnnotation: imagesAnnotation}
	}

	return editChartMetadata(ctx, ch, edit)
}

// editChartMetadata applies edit to the chart metadata and, for a chart with .werf-data, to its
// original Chart.yaml in place, which ChartToBytes then writes as is.
func editChartMetadata(ctx context.Context, ch *chart.Chart, edit chartyaml.EditOptions) error {
	if edit.Name == ch.Metadata.Name {
		edit.Name = ""
	}

	if edit.Version == ch.Metadata.Version {
		edit.Version = ""
	}

	ch.Metadata.Name = lo.CoalesceOrEmpty(edit.Name, ch.Metadata.Name)
	ch.Metadata.Version = lo.CoalesceOrEmpty(edit.Version, ch.Metadata.Version)

	if len(edit.SetAnnotations) > 0 || len(edit.DeleteAnnotations) > 0 {
		ch.Metadata.Annotations = lo.OmitByKeys(lo.Assign(ch.Metadata.Annotations, edit.SetAnnotations), edit.DeleteAnnotations)
	}

	if len(ch.WerfDataFiles) == 0 {
		return nil
	}

	for _, f := range ch.Raw {
		if f.Name != chartv2util.ChartfileName {
			continue
		}

		chartYAML, err := chartyaml.Edit(ctx, f.Data, edit)
		if err != nil {
			return fmt.Errorf("edit Chart.yaml: %w", err)
		}

		f.Data = chartYAML

		return nil
	}

	return fmt.Errorf("%s not found in chart %q", chartv2util.ChartfileName, ch.Name())
}
