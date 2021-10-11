package storage

import "github.com/werf/werf/pkg/image"

type CustomTagMetadata struct {
	StageID string
	Tag     string
}

func newCustomTagMetadata(stageID, tag string) *CustomTagMetadata {
	return &CustomTagMetadata{
		StageID: stageID,
		Tag:     tag,
	}
}

func (m *CustomTagMetadata) ToLabels() map[string]string {
	return map[string]string{
		image.WerfCustomTagMetadataStageIDLabel: m.StageID,
		image.WerfCustomTagMetadataTag:          m.Tag,
	}
}

func newCustomTagMetadataFromLabels(labels map[string]string) *CustomTagMetadata {
	return &CustomTagMetadata{
		StageID: labels[image.WerfCustomTagMetadataStageIDLabel],
		Tag:     labels[image.WerfCustomTagMetadataTag],
	}
}
