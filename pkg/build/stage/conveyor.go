package stage

import (
	"context"

	"github.com/werf/werf/pkg/build/import_server"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/giterminism_manager"
	"github.com/werf/werf/pkg/storage"
)

type Conveyor interface {
	GetImportMetadata(ctx context.Context, projectName, id string) (*storage.ImportMetadata, error)
	PutImportMetadata(ctx context.Context, projectName string, metadata *storage.ImportMetadata) error
	RmImportMetadata(ctx context.Context, projectName, id string) error

	GetImageStageContentDigest(imageName, stageName string) string
	GetImageContentDigest(imageName string) string

	FetchImageStage(ctx context.Context, imageName, stageName string) error
	FetchLastNonEmptyImageStage(ctx context.Context, imageName string) error
	GetImageNameForLastImageStage(imageName string) string
	GetImageIDForLastImageStage(imageName string) string

	GetImageNameForImageStage(imageName, stageName string) string
	GetImageIDForImageStage(imageName, stageName string) string

	GetImportServer(ctx context.Context, imageName, stageName string) (import_server.ImportServer, error)
	GetLocalGitRepoVirtualMergeOptions() VirtualMergeOptions

	GiterminismManager() giterminism_manager.Interface

	UseLegacyStapelBuilder(cb container_backend.ContainerBackend) bool
}

type VirtualMergeOptions struct {
	VirtualMerge bool
}
