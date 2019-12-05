// +build integration

package ansible_test

import "github.com/flant/werf/integration/utils/werfexec"

func werfBuild(dir string, opts werfexec.CommandOptions, extraArgs ...string) error {
	return werfexec.ExecWerfCommand(dir, werfBinPath, opts, append([]string{"build", "--stages-storage", ":local"}, extraArgs...)...)
}
