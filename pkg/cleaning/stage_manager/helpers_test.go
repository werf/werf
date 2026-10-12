package stage_manager

import (
	"context"

	"github.com/werf/werf/v2/pkg/storage"
)

var (
	_ storage.PrimaryStagesStorage = (*metadataRoutingStorage)(nil)
	_ GitRepo                      = metadataRoutingGitRepo{}
)

type metadataRoutingStorage struct {
	storage.PrimaryStagesStorage
	images     map[string]map[string][]string
	customTags map[string]*storage.CustomTagMetadata
}

func (s *metadataRoutingStorage) GetAllAndGroupImageMetadataByImageName(_ context.Context, _ string, _ []string, _ ...storage.Option) (map[string]map[string][]string, map[string]map[string][]string, error) {
	return s.images, nil, nil
}

func (s *metadataRoutingStorage) GetStageCustomTagMetadataIDs(_ context.Context, _ ...storage.Option) ([]string, error) {
	var ids []string
	for id := range s.customTags {
		ids = append(ids, id)
	}
	return ids, nil
}

func (s *metadataRoutingStorage) GetStageCustomTagMetadata(_ context.Context, id string) (*storage.CustomTagMetadata, error) {
	metadata, ok := s.customTags[id]
	if !ok {
		return nil, storage.ErrCustomTagMetadataNotFound
	}
	return metadata, nil
}

type metadataRoutingGitRepo struct{}

func (metadataRoutingGitRepo) IsCommitExists(_ context.Context, _ string) (bool, error) {
	return true, nil
}
