package config

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bmatcuk/doublestar"
)

type GiterminismConfig struct {
	Config config `json:"config"`
}

type config struct {
	AllowUncommitted    bool                `json:"allowUncommitted"`
	GoTemplateRendering goTemplateRendering `json:"goTemplateRendering"`
	Stapel              stapel              `json:"stapel"`
}

type goTemplateRendering struct {
	AllowEnvVariables []string `json:"allowEnvVariables"`
}

func (r goTemplateRendering) IsEnvNameAccepted(name string) (bool, error) {
	for _, pattern := range r.AllowEnvVariables {
		if strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/") {
			r, err := regexp.Compile(pattern[1 : len(pattern)-1])
			if err != nil {
				return false, err
			}

			if r.MatchString(name) {
				return true, nil
			}
		} else {
			if pattern == name {
				return true, nil
			}
		}
	}

	return false, nil
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
