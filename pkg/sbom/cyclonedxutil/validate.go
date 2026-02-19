package cyclonedxutil

import (
	"fmt"
	"strings"
	"sync"

	"github.com/xeipuuv/gojsonschema"
)

var (
	cycloneDX16Schema     *gojsonschema.Schema
	cycloneDX16SchemaOnce sync.Once
	errCycloneDX16Schema  error
)

// preloadCycloneDX16Schema ensures offline usage because of the network restrictions.
func preloadCycloneDX16Schema() (*gojsonschema.Schema, error) {
	sl := gojsonschema.NewSchemaLoader()

	loader := gojsonschema.NewStringLoader(cycloneDX16SchemaValue)
	// We use sl.Compile directly. It will automatically detect the $id from the JSON content
	// and register it. sync.Once ensures we don't do this more than once per process.
	return sl.Compile(loader)
}

// ValidateCycloneDX16Schema validates the given JSON bytes against the CycloneDX 1.6 JSON Schema.
func ValidateCycloneDX16Schema(jsonBytes []byte) error {
	cycloneDX16SchemaOnce.Do(func() {
		cycloneDX16Schema, errCycloneDX16Schema = preloadCycloneDX16Schema()
	})

	if errCycloneDX16Schema != nil {
		return fmt.Errorf("failed to load CycloneDX 1.6 schema: %w", errCycloneDX16Schema)
	}

	documentLoader := gojsonschema.NewBytesLoader(jsonBytes)
	result, err := cycloneDX16Schema.Validate(documentLoader)
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
