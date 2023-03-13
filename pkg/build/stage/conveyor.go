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

	GetImageStageContentDigest(targetPlatform, imageName, stageName string) string
	GetImageContentDigest(targetPlatform, imageName string) string

	FetchImageStage(ctx context.Context, targetPlatform, imageName, stageName string) error
	FetchLastNonEmptyImageStage(ctx context.Context, targetPlatform, imageName string) error
	GetImageNameForLastImageStage(targetPlatform, imageName string) string
	GetImageIDForLastImageStage(targetPlatform, imageName string) string
	GetImageDigestForLastImageStage(targetPlatform, imageName string) string

	GetImageNameForImageStage(targetPlatform, imageName, stageName string) string
	GetImageIDForImageStage(targetPlatform, imageName, stageName string) string

	GetImportServer(ctx context.Context, targetPlatform, imageName, stageName string) (import_server.ImportServer, error)
	GetLocalGitRepoVirtualMergeOptions() VirtualMergeOptions

	GiterminismManager() giterminism_manager.Interface

	UseLegacyStapelBuilder(cb container_backend.ContainerBackend) bool
}

type VirtualMergeOptions struct {
	VirtualMerge bool
}
