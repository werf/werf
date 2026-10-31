package bundles

import (
	"path"
)

func chartNameFromRepo(repo string) string {
	return path.Base(repo)
}
