package label

import (
	"fmt"

	"github.com/werf/common-go/pkg/util"
)

type Label util.Pair[string, string]

func NewLabel(key, value string) Label {
	return Label{
		First:  key,
		Second: value,
	}
}

func (l Label) String() string {
	return fmt.Sprintf("%s=%s", l.First, l.Second)
}
