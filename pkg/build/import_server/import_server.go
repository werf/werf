package import_server

import (
	"context"

	"github.com/werf/werf/pkg/config"
)

type ImportServer interface {
	GetCopyCommand(ctx context.Context, importConfig *config.Import) string
}
