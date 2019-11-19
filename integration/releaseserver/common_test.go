package releaseserver_test

import "github.com/flant/werf/integration/utils/werfexec"

func werfDeploy(dir string, opts werfexec.CommandOptions, extraArgs ...string) error {
	return werfexec.ExecWerfCommand(dir, werfBinPath, opts, append([]string{"deploy", "--env", "dev"}, extraArgs...)...)
}

func werfDismiss(dir string, opts werfexec.CommandOptions) error {
	return werfexec.ExecWerfCommand(dir, werfBinPath, opts, "dismiss", "--env", "dev", "--with-namespace")
}
