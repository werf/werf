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
	"strings"

	"github.com/samber/lo"

	"github.com/werf/logboek"
	"github.com/werf/nelm/v2/pkg/helm/pkg/chart/chartyaml"
	chartcommon "github.com/werf/nelm/v2/pkg/helm/pkg/chart/common"
	chart "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2"
	chartv2util "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2/util"
	"github.com/werf/nelm/v2/pkg/werfdata"
)

// ReadWerfData returns the .werf-data of the chart directory. The returned bool is false, and the
// .werf-data empty, for a chart without .werf-data, e.g. a bundle published by werf <= 3.12.
func ReadWerfData(ctx context.Context, chartDir string) (*werfdata.Data, bool, error) {
	werfDataDir := filepath.Join(chartDir, werfdata.DirName)

	var files []*chartcommon.File
	if err := filepath.WalkDir(werfDataDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %q: %w", path, err)
		}

		relPath, err := filepath.Rel(chartDir, path)
		if err != nil {
			return fmt.Errorf("get relative path of %q: %w", path, err)
		}

		files = append(files, &chartcommon.File{Name: filepath.ToSlash(relPath), Data: data})

		return nil
	}); errors.Is(err, fs.ErrNotExist) {
		return &werfdata.Data{}, false, nil
	} else if err != nil {
		return nil, false, fmt.Errorf("read %q: %w", werfDataDir, err)
	}

	data, err := werfdata.Parse(files)
	if err != nil {
		return nil, false, fmt.Errorf("parse %s: %w", werfdata.DirName, err)
	}

	return data, true, nil
}

// ReleaseFlags are the release name, namespace and environment given to a bundle command.
type ReleaseFlags struct {
	Release   string
	Namespace string
	Env       string
}

// ResolveRelease fills in the release name and namespace not given in flags with the ones werf bundle
// publish recorded. They are recorded for the environment of the publication, werf.env of the
// extra values, and converge would resolve them differently for another one, so with another
// non-empty --env both must be given. Without recorded ones, the flags are returned as is.
func ResolveRelease(ctx context.Context, data *werfdata.Data, flags ReleaseFlags) (ReleaseFlags, error) {
	recorded := data.Metadata.Release
	if recorded == nil {
		return flags, nil
	}

	werfVals, _ := data.ExtraValues["werf"].(map[string]any)
	recordedEnv, _ := werfVals["env"].(string)

	if flags.Env != "" && flags.Env != recordedEnv {
		var missing []string
		if flags.Release == "" {
			missing = append(missing, "--release")
		}

		if flags.Namespace == "" {
			missing = append(missing, "--namespace")
		}

		if len(missing) > 0 {
			return ReleaseFlags{}, fmt.Errorf("release defaults recorded for env %q are not used with --env %q: pass %s", recordedEnv, flags.Env, strings.Join(missing, " and "))
		}

		return flags, nil
	}

	if flags.Release != "" && flags.Namespace != "" {
		return flags, nil
	}

	var used []string
	if flags.Release == "" && recorded.Name != "" {
		flags.Release = recorded.Name
		used = append(used, "release "+recorded.Name)
	}

	if flags.Namespace == "" && recorded.Namespace != "" {
		flags.Namespace = recorded.Namespace
		used = append(used, "namespace "+recorded.Namespace)
	}

	if len(used) > 0 {
		logboek.Context(ctx).Default().LogF("Using %s recorded in %s for env %q\n", strings.Join(used, " and "), werfdata.DirName, recordedEnv)
	}

	return flags, nil
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
