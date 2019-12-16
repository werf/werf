// +build integration integration_k8s

package releaseserver_test

import "github.com/flant/werf/integration/utils/liveexec"

func werfDeploy(dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(dir, werfBinPath, opts, append([]string{"deploy", "--env", "dev"}, extraArgs...)...)
}

func werfDismiss(dir string, opts liveexec.ExecCommandOptions) error {
	return liveexec.ExecCommand(dir, werfBinPath, opts, "dismiss", "--env", "dev", "--with-namespace")
}
