package instruction

import (
	"path/filepath"
	"strings"
)

// contextRelativeExcludes rebases --exclude patterns from source paths to the context dir:
// buildkit matches every pattern against each source path, buildah matches them against the
// context dir, so a pattern has to be repeated for every source path it may apply to.
func contextRelativeExcludes(sourcePaths, excludePatterns []string) []string {
	if len(excludePatterns) == 0 {
		return nil
	}

	var res []string
	for _, sourcePath := range sourcePaths {
		if strings.HasPrefix(sourcePath, "http://") || strings.HasPrefix(sourcePath, "https://") {
			continue
		}

		prefix := strings.TrimPrefix(filepath.Clean(sourcePath), string(filepath.Separator))

		for _, pattern := range excludePatterns {
			if negated := strings.HasPrefix(pattern, "!"); negated {
				res = append(res, "!"+filepath.Join(prefix, pattern[1:]))
				continue
			}
			res = append(res, filepath.Join(prefix, pattern))
		}
	}

	return res
}
