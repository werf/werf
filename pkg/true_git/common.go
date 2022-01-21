package true_git

import (
	"os"
	"os/exec"
)

var command = exec.Command
var cmd exec.Cmd

func getCommonGitOptions() []string {
	return []string{"-c", "core.autocrlf=false", "-c", "gc.auto=0"}
}

func debug() bool {
	return os.Getenv("WERF_DEBUG_TRUE_GIT") == "1"
}
