package config

import (
	"github.com/werf/werf/v2/pkg/sbom"
	"github.com/werf/werf/v2/pkg/sbom/cyclonedxutil/gost"
)

type MetaBuildSbom struct {
	Enable   bool
	Standard sbom.StandardType
	Gost     gost.Config
}
