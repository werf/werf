package filter

import (
	"slices"

	"github.com/samber/lo"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/container_backend/label"
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

func (list *FilterList) ToStringSlice() []string {
	outputList := make([]string, len(*list))
	for i, l := range *list {
		outputList[i] = l.String()
	}
	return outputList
}

func NewFilterListFromLabelList(inputList label.LabelList) *FilterList {
	outputList := make(FilterList, 0, len(inputList))
	for _, l := range inputList {
		outputList.Add(NewFilterFromLabel(l))
	}
	return &outputList
}
