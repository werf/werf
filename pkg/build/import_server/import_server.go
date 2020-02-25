package import_server

import "github.com/flant/werf/pkg/config"

type ImportServer interface {
	GetCopyCommand(importConfig *config.Import) string
}
