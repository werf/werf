package cyclonedxutil

import (
	"encoding/json"
	"fmt"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/google/uuid"
	jsonpatch "gopkg.in/evanphx/json-patch.v5"
	syaml "sigs.k8s.io/yaml"
)

// BuildCycloneDX16BOMFromYAMLFragment builds a CycloneDX BOM from a YAML document that can be:
//   - a full CycloneDX BOM document, or
//   - a partial "fragment" (e.g. only `components:`, `metadata:`, etc).
//
// Process:
//   - convert YAML -> JSON (strict) as-is
//   - create an empty BOM document based on the requested standard and encode it as JSON
//   - merge fragment JSON into the base JSON using JSON Merge Patch (RFC 7396)
//   - decode merged JSON into the resulting BOM document
func BuildCycloneDX16BOMFromYAMLFragment(fragmentYAML []byte) (*cdx.BOM, error) {
	jsonFromYAML, err := syaml.YAMLToJSONStrict(fragmentYAML)
	if err != nil {
		return nil, fmt.Errorf("sbom: invalid YAML fragment: %w", err)
	}

	// Create a base (empty) BOM for the requested standard.
	baseBOM := NewBOM()

	baseJSON, err := json.Marshal(baseBOM)
	if err != nil {
		return nil, fmt.Errorf("sbom: failed to build base CycloneDX document: %w", err)
	}

	// Merge fragment into base using JSON Merge Patch.
	mergedJSON, err := jsonpatch.MergePatch(baseJSON, jsonFromYAML)
	if err != nil {
		return nil, fmt.Errorf("sbom: failed to merge YAML fragment into base document: %w", err)
	}

	return BuildCycloneDX16BOMFromJSON(mergedJSON)
}

// BuildCycloneDX16BOMFromJSON builds a CycloneDX BOM document from JSON bytes and validates it.
// Currently, only CycloneDX@1.6 is supported.
//
// External SBOMs may contain duplicate entries in arrays with "uniqueItems: true"
// (e.g. components). To handle this gracefully, the function first deserializes
// and deduplicates the BOM, then validates the cleaned result against the schema.
func BuildCycloneDX16BOMFromJSON(bomJSON []byte) (*cdx.BOM, error) {
	var bom cdx.BOM
	if err := json.Unmarshal(bomJSON, &bom); err != nil {
		return nil, fmt.Errorf("cyclonedxutil: failed to decode CycloneDX document: %w", err)
	}

	DedupBOM(&bom)

	cleanJSON, err := json.Marshal(bom)
	if err != nil {
		return nil, fmt.Errorf("cyclonedxutil: failed to re-encode BOM after dedup: %w", err)
	}

	if err := ValidateCycloneDX16Schema(cleanJSON); err != nil {
		return nil, fmt.Errorf("cyclonedxutil: validation failed: %w", err)
	}

	// As a "hard-ish" check, ensure cyclonedx-go can encode the resulting BOM as JSON 1.6.
	if _, err := ToJSON(&bom); err != nil {
		return nil, fmt.Errorf("cyclonedxutil: failed to encode BOM for specVersion 1.6: %w", err)
	}

	return &bom, nil
}

// NewBOM creates a new CycloneDX 1.6 BOM with a unique serial number.
func NewBOM() *cdx.BOM {
	bom := cdx.NewBOM()
	bom.SerialNumber = "urn:uuid:" + uuid.New().String()
	return bom
}
