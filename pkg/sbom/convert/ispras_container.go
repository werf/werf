package convert

import (
	"context"
	"fmt"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/samber/lo"

	"github.com/werf/werf/v2/pkg/sbom/cyclonedxutil"
)

var _ Assembler = (*ISPRASContainerAssembler)(nil)

type ISPRASContainerAssembler struct{}

func (a *ISPRASContainerAssembler) Assemble(_ context.Context, images []*ImageSBOM, meta ProductMeta) (*cdx.BOM, error) {
	result, err := cyclonedxutil.MergeBOMs(nil, cyclonedxutil.MergeOpts{
		ImportBOMs: imageBOMs(images),
	})
	if err != nil {
		return nil, fmt.Errorf("merge image BOMs: %w", err)
	}

	var containers []cdx.Component
	for _, img := range images {
		container := cdx.Component{BOMRef: img.Name, Type: cdx.ComponentTypeContainer, Name: img.Name}

		if img.BOM.Metadata != nil && img.BOM.Metadata.Component != nil {
			container = *img.BOM.Metadata.Component
			container.BOMRef = img.Name
			container.Type = cdx.ComponentTypeContainer
			container.Name = img.Name
		}

		container.ExternalReferences = img.BOM.ExternalReferences
		container.Properties = img.BOM.Properties

		setMissingGOSTOnComponent(&container, img.GOST)

		imgComponents := lo.FromPtr(img.BOM.Components)
		if len(imgComponents) > 0 {
			container.Components = &imgComponents
		}

		containers = append(containers, container)
	}
	if len(containers) > 0 {
		result.Components = &containers
	} else {
		result.Components = nil
	}

	result.Metadata = buildProductMetadata(meta, aggregateImageGOST(images))

	return result, nil
}
