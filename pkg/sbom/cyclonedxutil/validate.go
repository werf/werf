package cyclonedxutil

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/xeipuuv/gojsonschema"
)

var (
	bom16Schema          *gojsonschema.Schema
	bom16SchemaOnce      sync.Once
	errCycloneDX16Schema error
)

// preloadCycloneDX16Schema ensures offline usage because of the network restrictions.
func preloadCycloneDX16Schema() (*gojsonschema.Schema, error) {
	sl := gojsonschema.NewSchemaLoader()

	// Add JSF schema to the loader to avoid network requests.
	// CycloneDX 1.6 schema $refs this schema.
	jsfLoader := gojsonschema.NewStringLoader(jsf_0_82_SchemaValue)
	if err := sl.AddSchemas(jsfLoader); err != nil {
		return nil, fmt.Errorf("failed to add JSF schema: %w", err)
	}

	loader := gojsonschema.NewStringLoader(bom_1_6_SchemaValue)
	// We use sl.Compile directly. It will automatically detect the $id from the JSON content
	// and register it. sync.Once ensures we don't do this more than once per process.
	return sl.Compile(loader)
}

// ValidateCycloneDX16Schema validates the given JSON bytes against the CycloneDX 1.6 JSON Schema.
func ValidateCycloneDX16Schema(jsonBytes []byte) error {
	bom16SchemaOnce.Do(func() {
		bom16Schema, errCycloneDX16Schema = preloadCycloneDX16Schema()
	})

	if errCycloneDX16Schema != nil {
		return fmt.Errorf("failed to load CycloneDX 1.6 schema: %w", errCycloneDX16Schema)
	}

	documentLoader := gojsonschema.NewBytesLoader(jsonBytes)
	result, err := bom16Schema.Validate(documentLoader)
	if err != nil {
		return fmt.Errorf("failed to execute schema validation: %w", err)
	}

	if !result.Valid() {
		var errs []string
		for _, desc := range result.Errors() {
			errs = append(errs, desc.String())
		}
		return fmt.Errorf("cyclonedx: schema validation errors: %s", strings.Join(errs, "; "))
	}

	return nil
}

// ValidateBOM validates the given CycloneDX BOM against the CycloneDX 1.6 JSON Schema.
func ValidateBOM(bom *cdx.BOM) error {
	if bom == nil {
		return nil
	}

	jsonBytes, err := json.Marshal(bom)
	if err != nil {
		return fmt.Errorf("encode BOM for validation: %w", err)
	}

	return ValidateCycloneDX16Schema(jsonBytes)
}
