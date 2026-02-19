package cyclonedxutil

import (
	"encoding/json"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"

	"github.com/werf/common-go/pkg/util"
)

type MergeOpts struct {
	BaseBOM     *cdx.BOM
	ImportBOMs  []*cdx.BOM
	FragmentBOM *cdx.BOM
}

func (o MergeOpts) IsEmpty() bool {
	return o.BaseBOM == nil && len(o.ImportBOMs) == 0 && o.FragmentBOM == nil
}

func (o MergeOpts) Checksum() string {
	var parts []string
	for _, bom := range append([]*cdx.BOM{o.BaseBOM, o.FragmentBOM}, o.ImportBOMs...) {
		if cs := BOMChecksum(bom); cs != "" {
			parts = append(parts, cs)
		}
	}
	return strings.Join(parts, "-")
}

func (o MergeOpts) mergeOrder(target *cdx.BOM) []*cdx.BOM {
	boms := make([]*cdx.BOM, 0, len(o.ImportBOMs)+3)
	boms = append(boms, o.BaseBOM)
	boms = append(boms, o.ImportBOMs...)
	boms = append(boms, o.FragmentBOM, target)
	return boms
}

func MergeBOMs(target *cdx.BOM, opts MergeOpts) *cdx.BOM {
	result := NewBOM()

	if target != nil && target.Metadata != nil {
		result.Metadata = target.Metadata
	}

	boms := opts.mergeOrder(target)

	result.Components = mergeComponents(boms)
	result.Services = mergeServices(boms)
	result.Vulnerabilities = mergeVulnerabilities(boms)
	result.ExternalReferences = mergeExternalReferences(boms)
	result.Compositions = mergeCompositions(boms)
	result.Properties = mergeProperties(boms)
	result.Annotations = mergeAnnotations(boms)
	result.Formulation = mergeFormulation(boms)

	return result
}

func mergeComponents(boms []*cdx.BOM) *[]cdx.Component {
	var components []cdx.Component
	for _, bom := range boms {
		components = appendBOMComponents(components, bom)
	}
	if len(components) > 0 {
		return &components
	}

	return nil
}

func mergeServices(boms []*cdx.BOM) *[]cdx.Service {
	var services []cdx.Service
	for _, bom := range boms {
		services = appendBOMServices(services, bom)
	}
	if len(services) > 0 {
		return &services
	}

	return nil
}

func mergeVulnerabilities(boms []*cdx.BOM) *[]cdx.Vulnerability {
	var vulnerabilities []cdx.Vulnerability
	for _, bom := range boms {
		vulnerabilities = appendBOMVulnerabilities(vulnerabilities, bom)
	}
	if len(vulnerabilities) > 0 {
		return &vulnerabilities
	}

	return nil
}

func mergeExternalReferences(boms []*cdx.BOM) *[]cdx.ExternalReference {
	var externalReferences []cdx.ExternalReference
	for _, bom := range boms {
		externalReferences = appendBOMExternalReferences(externalReferences, bom)
	}
	if len(externalReferences) > 0 {
		return &externalReferences
	}

	return nil
}

func mergeCompositions(boms []*cdx.BOM) *[]cdx.Composition {
	var compositions []cdx.Composition
	for _, bom := range boms {
		compositions = appendBOMCompositions(compositions, bom)
	}
	if len(compositions) > 0 {
		return &compositions
	}

	return nil
}

func mergeProperties(boms []*cdx.BOM) *[]cdx.Property {
	var properties []cdx.Property
	for _, bom := range boms {
		properties = appendBOMProperties(properties, bom)
	}
	if len(properties) > 0 {
		return &properties
	}

	return nil
}

func mergeAnnotations(boms []*cdx.BOM) *[]cdx.Annotation {
	var annotations []cdx.Annotation
	for _, bom := range boms {
		annotations = appendBOMAnnotations(annotations, bom)
	}
	if len(annotations) > 0 {
		return &annotations
	}

	return nil
}

func mergeFormulation(boms []*cdx.BOM) *[]cdx.Formula {
	var formulation []cdx.Formula
	for _, bom := range boms {
		formulation = appendBOMFormulation(formulation, bom)
	}
	if len(formulation) > 0 {
		return &formulation
	}

	return nil
}

func appendBOMComponents(dest []cdx.Component, bom *cdx.BOM) []cdx.Component {
	if bom != nil && bom.Components != nil {
		return append(dest, *bom.Components...)
	}

	return dest
}

func appendBOMServices(dest []cdx.Service, bom *cdx.BOM) []cdx.Service {
	if bom != nil && bom.Services != nil {
		return append(dest, *bom.Services...)
	}

	return dest
}

func appendBOMVulnerabilities(dest []cdx.Vulnerability, bom *cdx.BOM) []cdx.Vulnerability {
	if bom != nil && bom.Vulnerabilities != nil {
		return append(dest, *bom.Vulnerabilities...)
	}

	return dest
}

func appendBOMExternalReferences(dest []cdx.ExternalReference, bom *cdx.BOM) []cdx.ExternalReference {
	if bom != nil && bom.ExternalReferences != nil {
		return append(dest, *bom.ExternalReferences...)
	}

	return dest
}

func appendBOMCompositions(dest []cdx.Composition, bom *cdx.BOM) []cdx.Composition {
	if bom != nil && bom.Compositions != nil {
		return append(dest, *bom.Compositions...)
	}

	return dest
}

func appendBOMProperties(dest []cdx.Property, bom *cdx.BOM) []cdx.Property {
	if bom != nil && bom.Properties != nil {
		return append(dest, *bom.Properties...)
	}

	return dest
}

func appendBOMAnnotations(dest []cdx.Annotation, bom *cdx.BOM) []cdx.Annotation {
	if bom != nil && bom.Annotations != nil {
		return append(dest, *bom.Annotations...)
	}

	return dest
}

func appendBOMFormulation(dest []cdx.Formula, bom *cdx.BOM) []cdx.Formula {
	if bom != nil && bom.Formulation != nil {
		return append(dest, *bom.Formulation...)
	}

	return dest
}

func BOMChecksum(bom *cdx.BOM) string {
	if bom == nil {
		return ""
	}
	data, err := json.Marshal(bom)
	if err != nil {
		return ""
	}

	return util.Sha256Hash(string(data))
}
