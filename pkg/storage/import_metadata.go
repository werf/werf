package storage

import (
	"github.com/werf/werf/pkg/image"
)

type ImportMetadata struct {
	ImportSourceID string
	SourceImageID  string
	Checksum       string
}

func (m *ImportMetadata) ToLabels() map[string]string {
	return map[string]string{
		image.WerfImportMetadataImportSourceIDLabel: m.ImportSourceID,
		image.WerfImportMetadataSourceImageIDLabel:  m.SourceImageID,
		image.WerfImportMetadataChecksumLabel:       m.Checksum,
	}
}

func newImportMetadataFromLabels(labels map[string]string) *ImportMetadata {
	return &ImportMetadata{
		ImportSourceID: labels[image.WerfImportMetadataImportSourceIDLabel],
		SourceImageID:  labels[image.WerfImportMetadataSourceImageIDLabel],
		Checksum:       labels[image.WerfImportMetadataChecksumLabel],
	}
}
