package import_server

import (
	"context"

	"github.com/werf/werf/v2/pkg/config"
)

type ImportServer interface {
	GetCopyCommand(ctx context.Context, importConfig *config.Import) string
}
