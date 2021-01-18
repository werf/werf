package config

import (
	"fmt"
	"github.com/bmatcuk/doublestar"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

type GiterminismConfig struct {
	Config config `json:"config"`
	Helm   helm   `json:"helm"`
}

type config struct {
	AllowUncommitted          bool                `json:"allowUncommitted"`
	AllowUncommittedTemplates []string            `json:"allowUncommittedTemplates"`
	GoTemplateRendering       goTemplateRendering `json:"goTemplateRendering"`
	Stapel                    stapel              `json:"stapel"`
	Dockerfile                dockerfile          `json:"dockerfile"`
}

func (c config) IsUncommittedTemplateFileAccepted(path string) (bool, error) {
	return isPathMatched(c.AllowUncommittedTemplates, path)
}

type goTemplateRendering struct {
	AllowEnvVariables     []string `json:"allowEnvVariables"`
	AllowUncommittedFiles []string `json:"allowUncommittedFiles"`
}

func (r goTemplateRendering) IsEnvNameAccepted(name string) (bool, error) {
	for _, pattern := range r.AllowEnvVariables {
		if strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/") {
			expr := fmt.Sprintf("^%s$", pattern[1:len(pattern)-1])
			r, err := regexp.Compile(expr)
			if err != nil {
				return false, err
			}

			return r.MatchString(name), nil
		} else {
			return pattern == name, nil
		}
	}

	return false, nil
}

func (r goTemplateRendering) IsUncommittedFileAccepted(path string) (bool, error) {
	return isPathMatched(r.AllowUncommittedFiles, path)
}

type stapel struct {
	AllowFromLatest bool  `json:"allowFromLatest"`
	Git             git   `json:"git"`
	Mount           mount `json:"mount"`
}

type git struct {
	AllowBranch bool `json:"allowBranch"`
}

type mount struct {
	AllowBuildDir  bool     `json:"allowBuildDir"`
	AllowFromPaths []string `json:"allowFromPaths"`
}

func (m mount) IsFromPathAccepted(path string) (bool, error) {
	return isPathMatched(m.AllowFromPaths, path)
}

type dockerfile struct {
	AllowUncommitted                  []string `json:"allowUncommitted"`
	AllowUncommittedDockerignoreFiles []string `json:"allowUncommittedDockerignoreFiles"`
	AllowContextAddFiles              []string `json:"allowContextAddFiles"`
}

func (d dockerfile) IsContextAddFileAccepted(path string) (bool, error) {
	return isPathMatched(d.AllowContextAddFiles, path)
}

func (d dockerfile) IsUncommittedAccepted(path string) (bool, error) {
	return isPathMatched(d.AllowUncommitted, path)
}

func (d dockerfile) IsUncommittedDockerignoreAccepted(path string) (bool, error) {
	return isPathMatched(d.AllowUncommittedDockerignoreFiles, path)
}

type helm struct {
	AllowUncommittedFiles []string `json:"allowUncommittedFiles"`
}

func (h helm) IsUncommittedHelmFileAccepted(path string) (bool, error) {
	return isPathMatched(h.AllowUncommittedFiles, path)
}

func isPathMatched(patterns []string, p string) (bool, error) {
	p = filepath.ToSlash(p)
	for _, pattern := range patterns {
		pattern = filepath.ToSlash(pattern)

		matchFunc := func() (bool, error) {
			exist, err := doublestar.PathMatch(pattern, p)
			if err != nil {
				return false, err
			}

			if exist {
				return true, nil
			}

			return doublestar.PathMatch(path.Join(pattern, "**", "*"), p)
		}

		if matched, err := matchFunc(); err != nil {
			return false, fmt.Errorf("unable to match path (pattern: %s, path %s): %s", pattern, p, err)
		} else if matched {
			return true, nil
		}
	}

	return false, nil
}
