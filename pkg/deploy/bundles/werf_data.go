package bundles

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/samber/lo"
	"sigs.k8s.io/yaml"

	"github.com/werf/nelm/v2/pkg/chartyaml"
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

	metadata := &werfdata.Metadata{}
	if err := yaml.UnmarshalStrict(data, metadata); err != nil {
		return nil, false, fmt.Errorf("parse %q: %w", metadataPath, err)
	}

	if err := werfdata.CheckAPIVersion(metadata.APIVersion); err != nil {
		return nil, false, err
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
