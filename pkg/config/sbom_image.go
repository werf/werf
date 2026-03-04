package config

import (
	"fmt"
	"strings"

	sbomPkg "github.com/werf/werf/v2/pkg/sbom"
	"github.com/werf/werf/v2/pkg/sbom/cyclonedxutil"
)

// buildImageSbom builds image-level SBOM configuration based on meta build settings.
func buildImageSbom(meta *Meta, raw *rawSbom, d *doc) (*Sbom, error) {
	if d == nil {
		// Fallback: avoid panics in error formatting in unexpected edge cases.
		d = &doc{Content: []byte{}}
	}

	if meta == nil {
		return nil, newDetailedConfigError("internal error: meta is not set while building image sbom", nil, d)
	}

	metaSbom := meta.Build.Sbom
	metaEnabled := metaSbom != nil && metaSbom.Enable

	if !metaEnabled {
		if raw != nil {
			return nil, newDetailedConfigError("`sbom` is specified for the image, but `build.sbom.enable` is false", nil, d)
		}
		return nil, nil
	}

	// Determine GOST config (fallback meta -> image).
	gostConfig := metaSbom.Gost
	if raw != nil && raw.Gost != nil {
		gostConfig = gostConfig.Merge(raw.Gost.toConfig())
	}

	// If no image-level configs are provided, we return early with the inherited GOST configuration.
	if raw == nil {
		return &Sbom{
			Standard: sbomPkg.StandardTypeCycloneDX16,
			Gost:     gostConfig,
		}, nil
	}

	// Defensive check: meta-level validation currently allows only CycloneDX@1.6.
	if metaSbom.Standard != sbomPkg.StandardTypeCycloneDX16 {
		return nil, newDetailedConfigError(
			fmt.Sprintf(
				"unsupported sbom standard %q for image sbom (only %q is supported)",
				metaSbom.Standard.String(),
				sbomPkg.StandardTypeCycloneDX16.String(),
			),
			nil,
			d,
		)
	}

	// If fragment is not specified, we return the configuration with GOST only.
	if raw.Fragment == nil {
		return &Sbom{
			Standard: sbomPkg.StandardTypeCycloneDX16,
			Gost:     gostConfig,
		}, nil
	}

	fragment := strings.TrimSpace(*raw.Fragment)
	if fragment == "" {
		return nil, newDetailedConfigError("`sbom.fragment` must not be empty if specified", nil, d)
	}

	bom, err := cyclonedxutil.BuildCycloneDX16BOMFromYAMLFragment([]byte(fragment))
	if err != nil {
		return nil, newDetailedConfigError(fmt.Sprintf("invalid `sbom.fragment`: %v", err), nil, d)
	}

	return &Sbom{
		Standard: sbomPkg.StandardTypeCycloneDX16,
		Document: bom,
		Gost:     gostConfig,
	}, nil
}
