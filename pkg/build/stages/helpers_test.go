package stages

import (
	"context"

	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/storage"
)

type importMetadataRegistryStub struct {
	docker_registry.Interface
	info *image.Info
}

var _ docker_registry.Interface = (*importMetadataRegistryStub)(nil)

func (r *importMetadataRegistryStub) GetRepoImage(_ context.Context, _ string) (*image.Info, error) {
	return r.info, nil
}

type importMetadataStorageStub struct {
	storage.PrimaryStagesStorage
	ids       []string
	listCalls int
	metadata  map[string]*storage.ImportMetadata
	errors    map[string]error
}

var _ storage.PrimaryStagesStorage = (*importMetadataStorageStub)(nil)

func (s *importMetadataStorageStub) GetImportMetadataIDs(_ context.Context, _ string, _ ...storage.Option) ([]string, error) {
	s.listCalls++
	return s.ids, nil
}

func (s *importMetadataStorageStub) GetImportMetadata(_ context.Context, _, id string) (*storage.ImportMetadata, error) {
	return s.metadata[id], s.errors[id]
}
