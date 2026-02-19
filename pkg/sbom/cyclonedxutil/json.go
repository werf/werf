package cyclonedxutil

import (
	"bytes"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

// ToJSON encodes the BOM to JSON using CycloneDX 1.6 specification.
func ToJSON(bom *cdx.BOM) ([]byte, error) {
	var buf bytes.Buffer
	if err := cdx.NewBOMEncoder(&buf, cdx.BOMFileFormatJSON).EncodeVersion(bom, cdx.SpecVersion1_6); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
