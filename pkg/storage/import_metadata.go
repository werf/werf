package storage

import (
	"fmt"

	"github.com/werf/werf/v2/pkg/image"
)

type ImportMetadata struct {
	ImportSourceID string
	SourceStageID  string
	Checksum       string
}

func (m *ImportMetadata) ToLabels() []string {
	return []string{
		fmt.Sprintf("%s=%s", image.WerfImportMetadataImportSourceIDLabel, m.ImportSourceID),
		fmt.Sprintf("%s=%s", image.WerfImportMetadataSourceStageIDLabel, m.SourceStageID),
		fmt.Sprintf("%s=%s", image.WerfImportMetadataChecksumLabel, m.Checksum),
	}
}

func (m *ImportMetadata) ToLabelsMap() map[string]string {
	return map[string]string{
		image.WerfImportMetadataImportSourceIDLabel: m.ImportSourceID,
		image.WerfImportMetadataSourceStageIDLabel:  m.SourceStageID,
		image.WerfImportMetadataChecksumLabel:       m.Checksum,
	}
}

func newImportMetadataFromLabels(labels map[string]string) *ImportMetadata {
	return &ImportMetadata{
		ImportSourceID: labels[image.WerfImportMetadataImportSourceIDLabel],
		SourceStageID:  labels[image.WerfImportMetadataSourceStageIDLabel],
		Checksum:       labels[image.WerfImportMetadataChecksumLabel],
	}
}
