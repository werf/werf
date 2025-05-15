package true_git

import (
	"os"
)

func getCommonGitOptions() []string {
	return []string{"-c", "core.autocrlf=false", "-c", "gc.auto=0", "-c", "commit.gpgsign=false", "-c", "core.untrackedCache=false", "-c", "core.splitIndex=false"}
}

func debug() bool {
	return os.Getenv("WERF_DEBUG_TRUE_GIT") == "1"
}
