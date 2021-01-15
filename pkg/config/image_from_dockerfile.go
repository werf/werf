package config

import (
	"path/filepath"

	"github.com/werf/werf/pkg/giterminism"
)

type ImageFromDockerfile struct {
	Name           string
	Dockerfile     string
	Context        string
	ContextAddFile []string
	Target         string
	Args           map[string]interface{}
	AddHost        []string
	Network        string
	SSH            string

	raw *rawImageFromDockerfile
}

func (c *ImageFromDockerfile) validate(giterminismManager giterminism.Manager) error {
	if !isRelativePath(c.Context) {
		return newDetailedConfigError("`context: PATH` should be relative to project directory!", nil, c.raw.doc)
	} else if c.Dockerfile != "" && !isRelativePath(c.Dockerfile) {
		return newDetailedConfigError("`dockerfile: PATH` required and should be relative to context!", nil, c.raw.doc)
	} else if !allRelativePaths(c.ContextAddFile) {
		return newDetailedConfigError("`contextAddFile: [PATH, ...]|PATH` each path should be relative to context!", nil, c.raw.doc)
	}

	if len(c.ContextAddFile) != 0 {
		for _, contextAddFile := range c.ContextAddFile {
			if err := giterminismManager.Inspector().InspectConfigDockerfileContextAddFile(filepath.Join(c.Context, contextAddFile)); err != nil {
				return newDetailedConfigError(err.Error(), nil, c.raw.doc)
			}
		}
	}

	return nil
}

func (c *ImageFromDockerfile) GetName() string {
	return c.Name
}
