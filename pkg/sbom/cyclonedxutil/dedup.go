package cyclonedxutil

import (
	"crypto/sha256"
	"encoding/json"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

// dedupJSONSlice removes duplicate items from a slice by comparing their JSON
// representations. This matches the "uniqueItems" semantics of JSON Schema
// (deep equality over serialized form). First occurrence wins; order is preserved.
//
// Uses SHA-256 hashing to keep memory usage constant per entry (~32 bytes)
// regardless of item size, which matters for large SBOMs (20k+ components).
func dedupJSONSlice[T any](items []T) []T {
	if len(items) == 0 {
		return items
	}

	seen := make(map[[sha256.Size]byte]struct{}, len(items))
	result := make([]T, 0, len(items))

	for i := range items {
		jsonBytes, err := json.Marshal(items[i])
		if err != nil {
			result = append(result, items[i])
			continue
		}

		hash := sha256.Sum256(jsonBytes)
		if _, exists := seen[hash]; exists {
			continue
		}

		seen[hash] = struct{}{}
		result = append(result, items[i])
	}

	return result
}

// dedupPtrSlice deduplicates the contents of a pointer-to-slice using JSON
// deep equality. Returns nil when the input is nil or the result is empty.
func dedupPtrSlice[T any](items *[]T) *[]T {
	if items == nil {
		return nil
	}

	deduped := dedupJSONSlice(*items)
	if len(deduped) == 0 {
		return nil
	}

	return &deduped
}

// DedupBOM removes duplicate entries from all top-level BOM arrays that have
// "uniqueItems: true" in the CycloneDX 1.6 JSON Schema: components, services,
// dependencies, compositions, vulnerabilities, annotations, and formulation.
func DedupBOM(bom *cdx.BOM) {
	if bom == nil {
		return
	}

	bom.Components = dedupPtrSlice(bom.Components)
	bom.Services = dedupPtrSlice(bom.Services)
	bom.Dependencies = dedupPtrSlice(bom.Dependencies)
	bom.Compositions = dedupPtrSlice(bom.Compositions)
	bom.Vulnerabilities = dedupPtrSlice(bom.Vulnerabilities)
	bom.Annotations = dedupPtrSlice(bom.Annotations)
	bom.Formulation = dedupPtrSlice(bom.Formulation)
}
