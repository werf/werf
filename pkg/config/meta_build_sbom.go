package config

import "github.com/werf/werf/v2/pkg/sbom"

type MetaBuildSbom struct {
	Enable   bool
	Standard sbom.StandardType
}
