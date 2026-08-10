package bundles

import (
	"strings"

	"github.com/werf/common-go/pkg/util"
)

func GetChartNameOverwrite(targetRepo, renameChart string, helmCompatibleChart bool) *string {
	switch {
	case renameChart != "":
		ret := new(string)
		*ret = renameChart
		return ret
	case helmCompatibleChart:
		ret := new(string)
		*ret = util.Reverse(strings.SplitN(util.Reverse(targetRepo), "/", 2)[0])
		return ret
	}

	return nil
}
