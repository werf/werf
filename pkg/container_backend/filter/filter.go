package filter

import (
	"fmt"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/container_backend/label"
)

type Filter util.Pair[string, string]

func NewFilter(key, value string) Filter {
	return Filter{
		First:  key,
		Second: value,
	}
}

func (f Filter) String() string {
	return fmt.Sprintf("%s=%s", f.First, f.Second)
}

func (f Filter) ToPair() util.Pair[string, string] {
	return util.NewPair(f.First, f.Second)
}

func NewFilterFromLabel(l label.Label) Filter {
	if l.Second == "" {
		return NewFilter(LabelPrefix, l.First)
	}

	return NewFilter(LabelPrefix, l.String())
}
