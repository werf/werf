package convert

import (
	"fmt"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/samber/lo"
)

func namespaceBOMRefs(bom *cdx.BOM, prefix string) {
	namespaceComponentBOMRefs(lo.FromPtr(bom.Components), prefix)

	for i, dep := range lo.FromPtr(bom.Dependencies) {
		(*bom.Dependencies)[i].Ref = namespacedRef(dep.Ref, prefix)

		newDeps := make([]string, len(lo.FromPtr(dep.Dependencies)))
		for j, d := range lo.FromPtr(dep.Dependencies) {
			newDeps[j] = namespacedRef(d, prefix)
		}
		if len(newDeps) > 0 {
			(*bom.Dependencies)[i].Dependencies = &newDeps
		}
	}
}

func namespaceComponentBOMRefs(components []cdx.Component, prefix string) {
	for i := range components {
		if components[i].BOMRef != "" {
			components[i].BOMRef = namespacedRef(components[i].BOMRef, prefix)
		}
		namespaceComponentBOMRefs(lo.FromPtr(components[i].Components), prefix)
	}
}

func namespacedRef(ref, prefix string) string {
	if ref == "" {
		return ""
	}

	return fmt.Sprintf("%s/%s", prefix, ref)
}
