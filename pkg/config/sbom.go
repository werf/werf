package config

import (
	cyclonedx "github.com/CycloneDX/cyclonedx-go"

	"github.com/werf/werf/v2/pkg/sbom"
	"github.com/werf/werf/v2/pkg/sbom/cyclonedxutil/gost"
)

// Sbom represents an SBOM directive attached to a specific image configuration.
//
// NOTE: At the config layer, this structure stores the final, normalized SBOM document
// (e.g. CycloneDX v1.6). Build pipeline behavior (generation/merge) is handled elsewhere.
type Sbom struct {
	// Standard is the SBOM standard for this document. Currently only CycloneDX@1.6 is supported.
	Standard sbom.StandardType

	// Document is the full SBOM document assembled from configuration (and, in the future, possibly merged
	// with scanner-generated output).
	Document *cyclonedx.BOM

	// Gost contains GOST-specific security properties configuration.
	Gost gost.Config
}
