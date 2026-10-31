package bundles

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/samber/lo"
	"sigs.k8s.io/yaml"

	"github.com/werf/logboek"
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

// ReleaseFlags are the release name, namespace and environment given to a bundle command.
type ReleaseFlags struct {
	Release   string
	Namespace string
	Env       string
}

// ResolveRelease fills in the release name and namespace not given in flags with the ones werf bundle
// publish recorded. They are recorded for the environment of the publication, and converge would
// resolve them differently for another one, so with another non-empty --env both must be given.
// Without recorded ones, the flags are returned as is.
func ResolveRelease(ctx context.Context, metadata *werfdata.Metadata, flags ReleaseFlags) (ReleaseFlags, error) {
	recorded := metadata.Release
	if recorded == nil {
		return flags, nil
	}

	if flags.Env != "" && flags.Env != recorded.Env {
		if flags.Release == "" || flags.Namespace == "" {
			return ReleaseFlags{}, fmt.Errorf("release defaults recorded for env %q are not used with --env %q: pass --release and --namespace", recorded.Env, flags.Env)
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
		logboek.Context(ctx).Default().LogF("Using %s recorded in %s for env %q\n", strings.Join(used, " and "), werfdata.DirName, recorded.Env)
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
