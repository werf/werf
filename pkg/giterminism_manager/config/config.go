package config

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/werf/werf/pkg/path_matcher"
)

func NewConfig(ctx context.Context, fileReader fileReader, configRelPath string) (c Config, err error) {
	exist, err := fileReader.IsGiterminismConfigExistAnywhere(ctx, configRelPath)
	if err != nil {
		return c, err
	}

	if !exist {
		return Config{}, nil
	}

	data, err := fileReader.ReadGiterminismConfig(ctx, configRelPath)
	if err != nil {
		return c, err
	}

	err = processWithOpenAPISchema(&data)
	if err != nil {
		return c, fmt.Errorf("the giterminism config validation failed: %w", err)
	}

	if err := json.Unmarshal(data, &c); err != nil {
		panic(fmt.Sprint("unexpected error: ", err))
	}

	return c, err
}

type fileReader interface {
	IsGiterminismConfigExistAnywhere(ctx context.Context, relPath string) (bool, error)
	ReadGiterminismConfig(ctx context.Context, relPath string) ([]byte, error)
}

type Config struct {
	Cli    cli    `json:"cli"`
	Config config `json:"config"`
	Helm   helm   `json:"helm"`
}

func (c Config) IsCustomTagsAccepted() bool {
	return c.Cli.AllowCustomTags
}

func (c Config) IsUncommittedConfigAccepted() bool {
	return c.Config.AllowUncommitted
}

func (c Config) UncommittedConfigTemplateFilePathMatcher() path_matcher.PathMatcher {
	return c.Config.UncommittedTemplateFilePathMatcher()
}

func (c Config) UncommittedConfigGoTemplateRenderingFilePathMatcher() path_matcher.PathMatcher {
	return c.Config.GoTemplateRendering.UncommittedFilePathMatcher()
}

func (c Config) IsConfigGoTemplateRenderingEnvNameAccepted(envName string) (bool, error) {
	return c.Config.GoTemplateRendering.IsEnvNameAccepted(envName)
}

func (c Config) IsConfigStapelFromLatestAccepted() bool {
	return c.Config.Stapel.AllowFromLatest
}

func (c Config) IsConfigStapelGitBranchAccepted() bool {
	return c.Config.Stapel.Git.AllowBranch
}

func (c Config) IsConfigStapelMountBuildDirAccepted() bool {
	return c.Config.Stapel.Mount.AllowBuildDir
}

func (c Config) IsConfigStapelMountFromPathAccepted(fromPath string) bool {
	return c.Config.Stapel.Mount.IsFromPathAccepted(fromPath)
}

func (c Config) IsConfigDockerfileContextAddFileAccepted(relPath string) bool {
	return c.Config.Dockerfile.IsContextAddFileAccepted(relPath)
}

func (c Config) IsUncommittedDockerfileAccepted(relPath string) bool {
	return c.Config.Dockerfile.IsUncommittedAccepted(relPath)
}

func (c Config) IsUncommittedDockerignoreAccepted(relPath string) bool {
	return c.Config.Dockerfile.IsUncommittedDockerignoreAccepted(relPath)
}

func (c Config) UncommittedHelmFilePathMatcher() path_matcher.PathMatcher {
	return c.Helm.UncommittedHelmFilePathMatcher()
}

type cli struct {
	AllowCustomTags bool `json:"allowCustomTags"`
}

type config struct {
	AllowUncommitted          bool                `json:"allowUncommitted"`
	AllowUncommittedTemplates []string            `json:"allowUncommittedTemplates"`
	GoTemplateRendering       goTemplateRendering `json:"goTemplateRendering"`
	Stapel                    stapel              `json:"stapel"`
	Dockerfile                dockerfile          `json:"dockerfile"`
}

func (c config) UncommittedTemplateFilePathMatcher() path_matcher.PathMatcher {
	return pathMatcher(c.AllowUncommittedTemplates)
}

type goTemplateRendering struct {
	AllowEnvVariables     []string `json:"allowEnvVariables"`
	AllowUncommittedFiles []string `json:"allowUncommittedFiles"`
}

func (r goTemplateRendering) IsEnvNameAccepted(name string) (bool, error) {
	for _, pattern := range r.AllowEnvVariables {
		match, err := func() (bool, error) {
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
		}()
		if err != nil {
			return false, err
		}

		if match {
			return true, nil
		}
	}

	return false, nil
}

func (r goTemplateRendering) UncommittedFilePathMatcher() path_matcher.PathMatcher {
	return pathMatcher(r.AllowUncommittedFiles)
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

func (m mount) IsFromPathAccepted(path string) bool {
	return isPathMatched(m.AllowFromPaths, path)
}

type dockerfile struct {
	AllowUncommitted                  []string `json:"allowUncommitted"`
	AllowUncommittedDockerignoreFiles []string `json:"allowUncommittedDockerignoreFiles"`
	AllowContextAddFiles              []string `json:"allowContextAddFiles"`
}

func (d dockerfile) IsContextAddFileAccepted(path string) bool {
	return isPathMatched(d.AllowContextAddFiles, path)
}

func (d dockerfile) IsUncommittedAccepted(path string) bool {
	return isPathMatched(d.AllowUncommitted, path)
}

func (d dockerfile) IsUncommittedDockerignoreAccepted(path string) bool {
	return isPathMatched(d.AllowUncommittedDockerignoreFiles, path)
}

type helm struct {
	AllowUncommittedFiles []string `json:"allowUncommittedFiles"`
}

func (h helm) UncommittedHelmFilePathMatcher() path_matcher.PathMatcher {
	return pathMatcher(h.AllowUncommittedFiles)
}

func isPathMatched(patterns []string, p string) bool {
	return pathMatcher(patterns).IsPathMatched(p)
}

func pathMatcher(patterns []string) path_matcher.PathMatcher {
	if len(patterns) != 0 {
		return path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{IncludeGlobs: patterns})
	} else {
		return path_matcher.NewFalsePathMatcher()
	}
}
