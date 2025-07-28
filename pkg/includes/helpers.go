package includes

import (
	"path"
	"path/filepath"
	"strings"
)

func sliceContainsSubstring(s string, substrings []string) bool {
	for _, sub := range substrings {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func prepareRelPath(fileName, add, to string) string {
	addClean := strings.TrimPrefix(filepath.Clean(add), string(filepath.Separator))
	relPath := strings.TrimPrefix(fileName, addClean)
	relPath = strings.TrimPrefix(relPath, string(filepath.Separator))
	newPath := path.Join(to, relPath)
	newPath = strings.TrimPrefix(newPath, string(filepath.Separator))
	return newPath
}
