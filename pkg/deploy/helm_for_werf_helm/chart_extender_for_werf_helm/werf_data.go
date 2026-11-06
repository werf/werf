package chart_extender_for_werf_helm

import (
	"fmt"

	"sigs.k8s.io/yaml"

	chart "github.com/werf/3p-helm-for-werf-helm/pkg/chart"
)

const werfDataMetadataFileName = ".werf-data/metadata.yaml"

// refuseWerfData rejects charts published by werf v3 and Nelm v2 with a .werf-data directory:
// werf v2 would deploy them without their publish-time values and annotations.
func refuseWerfData(files []*chart.ChartExtenderBufferedFile) error {
	for _, f := range files {
		if f.Name != werfDataMetadataFileName {
			continue
		}

		var metadata struct {
			APIVersion string `json:"apiVersion"`
		}
		if err := yaml.Unmarshal(f.Data, &metadata); err != nil {
			return fmt.Errorf("parse %s: %w", werfDataMetadataFileName, err)
		}

		if metadata.APIVersion == "" {
			continue
		}

		return fmt.Errorf("this chart has %s with apiVersion %q: it requires werf v3 or Nelm v2, or newer", werfDataMetadataFileName, metadata.APIVersion)
	}

	return nil
}
