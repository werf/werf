package gost

import (
	"fmt"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/samber/lo"
)

// Validate checks if the BOM and all its components (metadata and direct ones)
// contain the mandatory GOST properties with valid values.
func Validate(bom *cdx.BOM) error {
	if bom == nil {
		return fmt.Errorf("BOM is required")
	}

	if bom.SpecVersion != cdx.SpecVersion1_6 {
		return fmt.Errorf("GOST validation requires CycloneDX version 1.6, got %s", bom.SpecVersion)
	}

	if bom.Metadata != nil && bom.Metadata.Component != nil {
		if err := ValidateComponent(bom.Metadata.Component); err != nil {
			return fmt.Errorf("metadata component %q: %w", bom.Metadata.Component.Name, err)
		}
	}

	for _, comp := range lo.FromPtr(bom.Components) {
		if err := ValidateComponent(&comp); err != nil {
			return fmt.Errorf("component %q: %w", comp.Name, err)
		}
	}

	return nil
}

// ValidateComponent checks if a single component has the mandatory GOST properties.
func ValidateComponent(comp *cdx.Component) error {
	a := newAccessor(comp)
	var missing []string

	as, asOk := a.GetAttackSurface()
	if !asOk {
		missing = append(missing, PropertyAttackSurface)
	} else if !IsValidGostValue(as.String()) {
		return fmt.Errorf("invalid value for %s: %q (expected 'yes', 'no' or 'inherit')", PropertyAttackSurface, as)
	}

	sf, sfOk := a.GetSecurityFunction()
	if !sfOk {
		missing = append(missing, PropertySecurityFunction)
	} else if !IsValidGostValue(sf.String()) {
		return fmt.Errorf("invalid value for %s: %q (expected 'yes', 'no' or 'inherit')", PropertySecurityFunction, sf)
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing mandatory GOST properties: %v", missing)
	}

	return nil
}
