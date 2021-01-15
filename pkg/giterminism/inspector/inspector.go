package inspector

import (
	"github.com/werf/werf/pkg/giterminism"
)

type Inspector struct {
	manager giterminism.Manager
}

func NewInspector(manager giterminism.Manager) Inspector {
	return Inspector{manager}
}
