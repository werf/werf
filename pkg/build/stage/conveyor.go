package stage

import (
	"context"

	"github.com/werf/werf/v2/pkg/build/import_server"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/giterminism_manager"
)

type Conveyor interface {
	GetImageContentDigest(targetPlatform, imageName string) string
	GetImageContextDigest(targetPlatform, imageName string) string

	FetchLastNonEmptyImageStage(ctx context.Context, targetPlatform, imageName string) error
	GetImageNameForLastImageStage(targetPlatform, imageName string) string
	GetStageIDForLastImageStage(targetPlatform, imageName string) string
	// TODO: remove this legacy logic in v3.
	GetImageIDForLastImageStage(targetPlatform, imageName string) string
	GetImageDigestForLastImageStage(targetPlatform, imageName string) string

	GetImportServer(ctx context.Context, targetPlatform, imageName string, fromExternalImage bool) (import_server.ImportServer, error)

	GiterminismManager() giterminism_manager.Interface

	UseLegacyStapelBuilder(cb container_backend.ContainerBackend) bool
}
