package config

import (
	"fmt"
	"path/filepath"

	"github.com/bmatcuk/doublestar"
)

type GiterminismConfig struct {
	Config config `json:"config"`
}

type config struct {
	AllowUncommitted bool   `json:"allowUncommitted"`
	Stapel           stapel `json:"stapel"`
}

type stapel struct {
	Mount mount `json:"mount"`
}

type mount struct {
	AllowBuildDir  bool     `json:"allowBuildDir"`
	AllowFromPaths []string `json:"allowFromPaths"`
}

func (m mount) IsFromPathAccepted(path string) (bool, error) {
	return isPathMatched(m.AllowFromPaths, path, true)
}

func isPathMatched(patterns []string, path string, withGlobs bool) (bool, error) {
	path = filepath.ToSlash(path)
	for _, pattern := range patterns {
		pattern = filepath.ToSlash(pattern)
		var matchFunc func(string, string) (bool, error)
		if withGlobs {
			matchFunc = doublestar.Match
		} else {
			matchFunc = func(pattern string, path string) (bool, error) {
				return pattern == path, nil
			}
		}

		if matched, err := matchFunc(pattern, path); err != nil {
			return false, fmt.Errorf("unable to match path (pattern: %s, path %s): %s", pattern, path, err)
		} else if matched {
			return true, nil
		}
	}

	return false, nil
}
