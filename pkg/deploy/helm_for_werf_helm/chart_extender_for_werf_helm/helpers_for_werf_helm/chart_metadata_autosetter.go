package helpers_for_werf_helm

import (
	chart "github.com/werf/3p-helm-for-werf-helm/pkg/chart"
)

type GetHelmChartMetadataOptions struct {
	OverrideAppVersion string
	OverrideName       string
	DefaultName        string
	DefaultVersion     string
}

func AutosetChartMetadata(metadataIn *chart.Metadata, opts GetHelmChartMetadataOptions) *chart.Metadata {
	var metadata *chart.Metadata
	if metadataIn == nil {
		metadata = &chart.Metadata{
			APIVersion: chart.APIVersionV2,
		}
	} else {
		metadata = metadataIn
	}

	if opts.OverrideName != "" {
		metadata.Name = opts.OverrideName
	} else if metadata.Name == "" {
		metadata.Name = opts.DefaultName
	}

	if opts.OverrideAppVersion != "" {
		metadata.AppVersion = opts.OverrideAppVersion
	}

	if metadata.Version == "" {
		metadata.Version = opts.DefaultVersion
	}

	return metadata
}
