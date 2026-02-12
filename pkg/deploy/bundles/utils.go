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
		// TODO(major): HelmCompatibleChart mode is enabled by default for the 'werf bundle copy', but disabled for the 'werf bundle publish'. We need to decide whether compatibility mode will be enabled or disabled by default for all bundle commands.
		ret := new(string)
		*ret = util.Reverse(strings.SplitN(util.Reverse(targetRepo), "/", 2)[0])
		return ret
	}

	return nil
}
