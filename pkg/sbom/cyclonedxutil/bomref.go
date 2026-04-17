package cyclonedxutil

import (
	"crypto/sha256"
	"fmt"

	cdx "github.com/CycloneDX/cyclonedx-go"
	packageurl "github.com/package-url/packageurl-go"
)

func packageID(serial string, index int) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", serial, index)))
	return fmt.Sprintf("%x", h[:8])
}

func deriveBomRef(purl, serial string, index int) string {
	id := packageID(serial, index)

	parsed, err := packageurl.FromString(purl)
	if err != nil {
		return id
	}

	parsed.Qualifiers = append(parsed.Qualifiers, packageurl.Qualifier{
		Key:   "package-id",
		Value: id,
	})

	return parsed.ToString()
}

func assignNewRef(oldRef, purl, serial string, index int, refMap map[string]string) string {
	newRef := deriveBomRef(purl, serial, index)
	if oldRef != newRef {
		refMap[oldRef] = newRef
	}

	return newRef
}

func ensureUniqueBOMRefs(bom *cdx.BOM) {
	if bom == nil {
		return
	}

	refMap := map[string]string{}
	serial := bom.SerialNumber
	index := 0

	if bom.Components != nil {
		comps := *bom.Components
		for i := range comps {
			if comps[i].BOMRef != "" {
				comps[i].BOMRef = assignNewRef(comps[i].BOMRef, comps[i].PackageURL, serial, index, refMap)
			}
			index++
		}
	}

	if bom.Services != nil {
		svcs := *bom.Services
		for i := range svcs {
			if svcs[i].BOMRef != "" {
				svcs[i].BOMRef = assignNewRef(svcs[i].BOMRef, "", serial, index, refMap)
			}
			index++
		}
	}

	rewriteAllRefs(bom, refMap)
}

func remapRef(ref string, refMap map[string]string) string {
	if newRef, ok := refMap[ref]; ok {
		return newRef
	}

	return ref
}

func remapStringSlice(ss *[]string, refMap map[string]string) {
	if ss == nil {
		return
	}

	for i := range *ss {
		(*ss)[i] = remapRef((*ss)[i], refMap)
	}
}

func remapBOMReferenceSlice(refs *[]cdx.BOMReference, refMap map[string]string) {
	if refs == nil {
		return
	}

	for i := range *refs {
		(*refs)[i] = cdx.BOMReference(remapRef(string((*refs)[i]), refMap))
	}
}

func rewriteAllRefs(bom *cdx.BOM, refMap map[string]string) {
	if len(refMap) == 0 {
		return
	}

	rewriteDependencyRefs(bom.Dependencies, refMap)
	rewriteVulnerabilityRefs(bom.Vulnerabilities, refMap)
	rewriteCompositionRefs(bom.Compositions, refMap)
	rewriteAnnotationRefs(bom.Annotations, refMap)
	rewriteDeclarationRefs(bom.Declarations, refMap)
}

func rewriteDependencyRefs(deps *[]cdx.Dependency, refMap map[string]string) {
	if deps == nil {
		return
	}

	for i := range *deps {
		(*deps)[i].Ref = remapRef((*deps)[i].Ref, refMap)
		remapStringSlice((*deps)[i].Dependencies, refMap)
		remapStringSlice((*deps)[i].Provides, refMap)
	}
}

func rewriteVulnerabilityRefs(vulns *[]cdx.Vulnerability, refMap map[string]string) {
	if vulns == nil {
		return
	}

	for i := range *vulns {
		if (*vulns)[i].Affects == nil {
			continue
		}

		for j := range *(*vulns)[i].Affects {
			(*(*vulns)[i].Affects)[j].Ref = remapRef((*(*vulns)[i].Affects)[j].Ref, refMap)
		}
	}
}

func rewriteCompositionRefs(compositions *[]cdx.Composition, refMap map[string]string) {
	if compositions == nil {
		return
	}

	for i := range *compositions {
		remapBOMReferenceSlice((*compositions)[i].Assemblies, refMap)
		remapBOMReferenceSlice((*compositions)[i].Dependencies, refMap)
		remapBOMReferenceSlice((*compositions)[i].Vulnerabilities, refMap)
	}
}

func rewriteAnnotationRefs(annotations *[]cdx.Annotation, refMap map[string]string) {
	if annotations == nil {
		return
	}

	for i := range *annotations {
		remapBOMReferenceSlice((*annotations)[i].Subjects, refMap)
	}
}

func rewriteDeclarationRefs(declarations *cdx.Declarations, refMap map[string]string) {
	if declarations == nil {
		return
	}

	if declarations.Assessors != nil {
		for i := range *declarations.Assessors {
			(*declarations.Assessors)[i].BOMRef = cdx.BOMReference(remapRef(string((*declarations.Assessors)[i].BOMRef), refMap))
		}
	}

	if declarations.Claims != nil {
		for i := range *declarations.Claims {
			(*declarations.Claims)[i].BOMRef = remapRef((*declarations.Claims)[i].BOMRef, refMap)
		}
	}

	if declarations.Evidence != nil {
		for i := range *declarations.Evidence {
			(*declarations.Evidence)[i].BOMRef = remapRef((*declarations.Evidence)[i].BOMRef, refMap)
		}
	}
}
