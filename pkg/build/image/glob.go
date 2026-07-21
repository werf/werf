package image

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// extendedGlob is a port of github.com/containers/buildah/copier extendedGlob (Apache-2.0):
// filepath.Glob with "**" path components expanding to any number of subdirectories, preserved
// verbatim because stage digests depend on the exact set of matched paths.
func extendedGlob(pattern string) ([]string, error) {
	subdirs := func(dir string) []string {
		var subdirectories []string
		if err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				if rel, err := filepath.Rel(dir, path); err == nil {
					subdirectories = append(subdirectories, rel)
				}
			}
			return nil
		}); err != nil {
			subdirectories = []string{"."}
		}
		return subdirectories
	}
	expandPatterns := func(pattern string) []string {
		components := []string{}
		dir := pattern
		file := ""
		for dir != filepath.VolumeName(dir) && dir != string(os.PathSeparator) {
			dir, file = filepath.Split(dir)
			if file != "" {
				components = append([]string{file}, components...)
			}
			dir = strings.TrimSuffix(dir, string(os.PathSeparator))
		}
		patterns := []string{filepath.VolumeName(dir) + string(os.PathSeparator)}
		for i := range components {
			var nextPatterns []string
			if components[i] == "**" {
				for _, parent := range patterns {
					nextSubdirs := subdirs(parent)
					for _, nextSubdir := range nextSubdirs {
						nextPatterns = append(nextPatterns, filepath.Join(parent, nextSubdir))
					}
				}
			} else {
				for _, parent := range patterns {
					nextPattern := filepath.Join(parent, components[i])
					nextPatterns = append(nextPatterns, nextPattern)
				}
			}
			patterns = []string{}
			seen := map[string]struct{}{}
			for _, nextPattern := range nextPatterns {
				if _, seen2 := seen[nextPattern]; seen2 {
					continue
				}
				patterns = append(patterns, nextPattern)
				seen[nextPattern] = struct{}{}
			}
		}
		return patterns
	}
	patterns := expandPatterns(pattern)
	var matches []string
	for _, pattern := range patterns {
		theseMatches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		matches = append(matches, theseMatches...)
	}
	sort.Strings(matches)
	return matches, nil
}
