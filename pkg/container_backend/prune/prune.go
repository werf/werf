package prune

import "github.com/werf/common-go/pkg/util"

type Options struct {
	Filters []util.Pair[string, string]
}

type Report struct {
	ItemsDeleted   []string
	SpaceReclaimed uint64
}
