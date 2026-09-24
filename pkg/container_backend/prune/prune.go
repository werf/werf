package prune

import (
	"github.com/werf/werf/v3/pkg/container_backend/filter"
)

type Options struct {
	Filters filter.FilterList
}

type Report struct {
	ItemsDeleted   []string
	SpaceReclaimed uint64
}
