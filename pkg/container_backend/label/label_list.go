package label

import (
	"slices"

	"github.com/samber/lo"
)

type LabelList []Label

func (list *LabelList) Add(label Label) {
	*list = lo.Uniq(append(*list, label))
}

func (list *LabelList) Remove(label Label) {
	*list = slices.DeleteFunc(*list, func(l Label) bool {
		return l == label
	})
}

func (list *LabelList) ToStringSlice() []string {
	s1 := make([]string, len(*list))
	for i, l := range *list {
		s1[i] = l.String()
	}
	return s1
}

func NewLabelListFromMap(m map[string]string) *LabelList {
	list := new(LabelList)
	for k, v := range m {
		list.Add(NewLabel(k, v))
	}
	return list
}
