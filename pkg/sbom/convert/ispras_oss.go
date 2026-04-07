package convert

import (
	"context"
	"fmt"

	cdx "github.com/CycloneDX/cyclonedx-go"

	"github.com/werf/werf/v2/pkg/sbom/cyclonedxutil"
)

var _ Assembler = (*ISPRASOSSAssembler)(nil)

type ISPRASOSSAssembler struct{}

func (a *ISPRASOSSAssembler) Assemble(_ context.Context, images []*ImageSBOM, meta ProductMeta) (*cdx.BOM, error) {
	result, err := cyclonedxutil.MergeBOMs(nil, cyclonedxutil.MergeOpts{
		ImportBOMs: imageBOMs(images),
	})
	if err != nil {
		return nil, fmt.Errorf("merge image BOMs: %w", err)
	}

	result.Metadata = buildProductMetadata(meta, aggregateImageGOST(images))

	return result, nil
}
