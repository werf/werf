package git_repo

import "os"

func debug() bool {
	return os.Getenv("WERF_GIT_REPO_DEBUG") == "1"
}
