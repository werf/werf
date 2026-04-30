package stage

import (
	"context"

	"github.com/werf/werf/v2/pkg/build/import_server"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/giterminism_manager"
)

type Conveyor interface {
	GetImageContextTagStageID(targetPlatform, imageName string) string
	GetImageContextTagDigest(targetPlatform, imageName string) string

	GetImportServer(ctx context.Context, targetPlatform, imageName string, fromExternalImage bool) (import_server.ImportServer, error)

	GiterminismManager() giterminism_manager.Interface

	UseLegacyStapelBuilder(cb container_backend.ContainerBackend) bool
}
