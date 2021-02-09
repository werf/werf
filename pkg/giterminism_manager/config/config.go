package config

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bmatcuk/doublestar"
)

func NewConfig(ctx context.Context, fileReader fileReader) (c Config, err error) {
	exist, err := fileReader.IsGiterminismConfigExistAnywhere(ctx)
	if err != nil {
		return c, err
	}

	if !exist {
		return Config{}, nil
	}

	data, err := fileReader.ReadGiterminismConfig(ctx)
	if err != nil {
		return c, err
	}

	err = processWithOpenAPISchema(&data)
	if err != nil {
		return c, fmt.Errorf("the giterminism config validation failed: %s", err)
	}

	if err := json.Unmarshal(data, &c); err != nil {
		panic(fmt.Sprint("unexpected error: ", err))
	}

	return c, err
}

type fileReader interface {
	IsGiterminismConfigExistAnywhere(ctx context.Context) (bool, error)
	ReadGiterminismConfig(ctx context.Context) ([]byte, error)
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

func (c Config) IsUncommittedConfigTemplateFileAccepted(path string) (bool, error) {
	return c.Config.IsUncommittedTemplateFileAccepted(path)
}

func (c Config) IsUncommittedConfigGoTemplateRenderingFileAccepted(path string) (bool, error) {
	return c.Config.GoTemplateRendering.IsUncommittedFileAccepted(path)
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

func (c Config) IsConfigStapelMountFromPathAccepted(fromPath string) (bool, error) {
	return c.Config.Stapel.Mount.IsFromPathAccepted(fromPath)
}

func (c Config) IsConfigDockerfileContextAddFileAccepted(relPath string) (bool, error) {
	return c.Config.Dockerfile.IsContextAddFileAccepted(relPath)
}

func (c Config) IsUncommittedDockerfileAccepted(relPath string) (bool, error) {
	return c.Config.Dockerfile.IsUncommittedAccepted(relPath)
}

func (c Config) IsUncommittedDockerignoreAccepted(relPath string) (bool, error) {
	return c.Config.Dockerfile.IsUncommittedDockerignoreAccepted(relPath)
}

func (c Config) IsUncommittedHelmFileAccepted(relPath string) (bool, error) {
	return c.Helm.IsUncommittedHelmFileAccepted(relPath)
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

func (c config) IsUncommittedTemplateFileAccepted(path string) (bool, error) {
	return isPathMatched(c.AllowUncommittedTemplates, path)
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
			exist, err := doublestar.Match(pattern, p)
			if err != nil {
				return false, err
			}

			if exist {
				return true, nil
			}

			return doublestar.Match(path.Join(pattern, "**", "*"), p)
		}

		if matched, err := matchFunc(); err != nil {
			return false, fmt.Errorf("unable to match path (pattern: %q, path %q): %s", pattern, p, err)
		} else if matched {
			return true, nil
		}
	}

	return false, nil
}
