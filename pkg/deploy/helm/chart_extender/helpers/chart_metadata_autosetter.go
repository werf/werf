package helpers

import (
	"fmt"

	"helm.sh/helm/v3/pkg/chart"
)

var ErrMetadataIsMissing = fmt.Errorf("chart metadata is missing")

type GetHelmChartMetadataOptions struct {
	OverrideName   string
	DefaultName    string
	DefaultVersion string
}

func AutosetChartMetadata(metadataIn *chart.Metadata, opts GetHelmChartMetadataOptions) (*chart.Metadata, error) {
	if metadataIn == nil {
		return nil, ErrMetadataIsMissing
	}

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

	if metadata.Version == "" {
		metadata.Version = opts.DefaultVersion
	}

	return metadata, nil
}
