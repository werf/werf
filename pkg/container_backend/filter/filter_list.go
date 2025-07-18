package filter

import (
	"slices"

	"github.com/samber/lo"

	"github.com/werf/common-go/pkg/util"
)

type FilterList []Filter

func (list *FilterList) Add(filter Filter) {
	*list = lo.Uniq(append(*list, filter))
}

func (list *FilterList) Remove(filter Filter) {
	*list = slices.DeleteFunc(*list, func(f Filter) bool {
		return f == filter
	})
}

func (list *FilterList) ToPairs() []util.Pair[string, string] {
	return lo.Map(*list, func(f Filter, _ int) util.Pair[string, string] {
		return f.ToPair()
	})
}
