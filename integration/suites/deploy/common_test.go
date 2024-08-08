package deploy_test

import (
	"github.com/werf/werf/v2/test/pkg/utils/liveexec"
)

func werfConverge(dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(dir, SuiteData.WerfBinPath, opts, append([]string{"converge"}, extraArgs...)...)
}
