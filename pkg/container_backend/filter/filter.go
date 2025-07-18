package filter

import (
	"github.com/werf/common-go/pkg/util"
)

type Filter util.Pair[string, string]

func NewFilter(key, value string) Filter {
	return Filter{
		First:  key,
		Second: value,
	}
}

func (f *Filter) ToPair() util.Pair[string, string] {
	return util.NewPair(f.First, f.Second)
}
