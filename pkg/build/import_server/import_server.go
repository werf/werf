package import_server

import "github.com/werf/werf/pkg/config"

type ImportServer interface {
	GetCopyCommand(importConfig *config.Import) string
}
