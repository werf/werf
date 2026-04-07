package convert

import (
	"context"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

type Assembler interface {
	Assemble(ctx context.Context, images []*ImageSBOM, meta ProductMeta) (*cdx.BOM, error)
}

type Converter struct {
	Assembler Assembler
}

func (c *Converter) Convert(ctx context.Context, images []*ImageSBOM, meta ProductMeta) (*cdx.BOM, error) {
	return c.Assembler.Assemble(ctx, images, meta)
}

func buildProductMetadata(meta ProductMeta, gostValues GOSTValues) *cdx.Metadata {
	metaComponent := &cdx.Component{
		Type:    cdx.ComponentTypeApplication,
		Name:    meta.AppName,
		Version: meta.AppVersion,
		Manufacturer: &cdx.OrganizationalEntity{
			Name: meta.Manufacturer,
		},
	}

	return &cdx.Metadata{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Component: metaComponent,
	}
}

func imageBOMs(images []*ImageSBOM) []*cdx.BOM {
	boms := make([]*cdx.BOM, len(images))
	for i, img := range images {
		boms[i] = img.BOM
	}
	return boms
}
