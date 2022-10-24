package dockerfile

import (
	"path"
	"strings"
)

func NormalizeCopyAddSourcesForPathMatcher(wildcards []string) []string {
	var result []string
	for _, wildcard := range wildcards {
		normalizedWildcard := path.Clean(wildcard)
		if normalizedWildcard == "/" {
			normalizedWildcard = "."
		}
		normalizedWildcard = strings.TrimPrefix(normalizedWildcard, "/")

		result = append(result, normalizedWildcard)
	}

	return result
}
