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
	Helm   helm   `json:"helm"`
}

type config struct {
	AllowUncommitted    bool                `json:"allowUncommitted"`
	GoTemplateRendering goTemplateRendering `json:"goTemplateRendering"`
	Stapel              stapel              `json:"stapel"`
	Dockerfile          dockerfile          `json:"dockerfile"`
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
	AllowFromLatest bool  `json:"allowFromLatest"`
	Mount           mount `json:"mount"`
}

type mount struct {
	AllowBuildDir  bool     `json:"allowBuildDir"`
	AllowFromPaths []string `json:"allowFromPaths"`
}

func (m mount) IsFromPathAccepted(path string) (bool, error) {
	return isPathMatched(m.AllowFromPaths, path, true)
}

type dockerfile struct {
	AllowUncommitted                  []string `json:"allowUncommitted"`
	AllowUncommittedDockerignoreFiles []string `json:"allowUncommittedDockerignoreFiles"`
	AllowContextAddFile               []string `json:"allowContextAddFile"`
}

func (d dockerfile) IsContextAddFileAccepted(path string) (bool, error) {
	return isPathMatched(d.AllowContextAddFile, path, true)
}

func (d dockerfile) IsUncommittedAccepted(path string) (bool, error) {
	return isPathMatched(d.AllowUncommitted, path, true)
}

func (d dockerfile) IsUncommittedDockerignoreAccepted(path string) (bool, error) {
	return isPathMatched(d.AllowUncommittedDockerignoreFiles, path, true)
}

type helm struct {
	AllowUncommittedFiles []string `json:"allowUncommittedFiles"`
}

func isPathMatched(patterns []string, path string, withGlobs bool) (bool, error) {
	path = filepath.ToSlash(path)
	for _, pattern := range patterns {
		pattern = filepath.ToSlash(pattern)

		var matchFunc func(string, string) (bool, error)
		if strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/") && withGlobs {
			pattern = pattern[1 : len(pattern)-1]
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
