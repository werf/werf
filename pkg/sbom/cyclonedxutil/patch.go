package cyclonedxutil

import (
	cdx "github.com/CycloneDX/cyclonedx-go"
)

func PatchComponents(bom *cdx.BOM, match func(*cdx.Component) bool, patch func(*cdx.Component)) {
	if bom == nil || bom.Components == nil {
		return
	}

	for i := range *bom.Components {
		c := &(*bom.Components)[i]
		if match(c) {
			patch(c)
		}
	}
}
