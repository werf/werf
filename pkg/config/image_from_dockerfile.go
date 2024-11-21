package config

import (
	"fmt"
	"path/filepath"

	"github.com/werf/werf/v2/pkg/giterminism_manager"
	"github.com/werf/werf/v2/pkg/util"
)

type ImageFromDockerfile struct {
	Name            string
	Dockerfile      string
	Context         string
	ContextAddFiles []string
	Target          string
	Args            map[string]interface{}
	AddHost         []string
	Network         string
	SSH             string
	Dependencies    []*Dependency
	Staged          bool
	Secrets         []string

	platform []string
	final    bool
	raw      *rawImageFromDockerfile
}

func (c *ImageFromDockerfile) validate(giterminismManager giterminism_manager.Interface) error {
	switch {
	case !isRelativePath(c.Context):
		return newDetailedConfigError("`context: PATH` should be relative to project directory!", nil, c.raw.doc)
	case c.Dockerfile != "" && !isRelativePath(c.Dockerfile):
		return newDetailedConfigError("`dockerfile: PATH` required and should be relative to context!", nil, c.raw.doc)
	case !allRelativePaths(c.ContextAddFiles):
		return newDetailedConfigError("`contextAddFiles: [PATH, ...]|PATH` each path should be relative to context!", nil, c.raw.doc)
	case len(c.ContextAddFiles) != 0:
		for _, contextAddFile := range c.ContextAddFiles {
			if err := giterminismManager.Inspector().InspectConfigDockerfileContextAddFile(filepath.Join(c.Context, contextAddFile)); err != nil {
				return newDetailedConfigError(err.Error(), nil, c.raw.doc)
			}
		}
	}

	if len(c.Args) > 0 {
		for _, dep := range c.Dependencies {
			for _, depImport := range dep.Imports {
				if _, ok := c.Args[depImport.TargetBuildArg]; ok {
					return newDetailedConfigError(fmt.Sprintf("dockerfile `args:` build arg %q already defined in dependency import `targetBuildArg:` directive. This is not allowed, avoid duplicated build args!", depImport.TargetBuildArg), nil, c.raw.doc)
				}
			}
		}
	}

	return nil
}

func (c *ImageFromDockerfile) GetName() string {
	return c.Name
}

func (c *ImageFromDockerfile) IsStapel() bool {
	return false
}

func (c *ImageFromDockerfile) IsFinal() bool {
	return c.final
}

func (c *ImageFromDockerfile) Platform() []string {
	return c.platform
}

func (c *ImageFromDockerfile) dependsOn() DependsOn {
	var dependsOn DependsOn

	for _, dep := range c.Dependencies {
		dependsOn.Dependencies = append(dependsOn.Dependencies, dep.ImageName)
	}
	dependsOn.Dependencies = util.UniqStrings(dependsOn.Dependencies)

	return dependsOn
}

func (c *ImageFromDockerfile) rawDoc() *doc {
	return c.raw.doc
}
