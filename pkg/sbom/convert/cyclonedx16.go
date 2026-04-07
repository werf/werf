package convert

import (
	"context"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/samber/lo"
)

func NewImageSBOMFromCycloneDX16(_ context.Context, name string, bom *cdx.BOM) (*ImageSBOM, error) {
	namespaceBOMRefs(bom, name)

	return &ImageSBOM{
		Name: name,
		BOM:  bom,
		GOST: aggregateGOST(lo.FromPtr(bom.Components)),
	}, nil
}
