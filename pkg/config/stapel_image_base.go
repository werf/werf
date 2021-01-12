package config

import (
	"context"
	"fmt"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/giterminism_inspector"
)

type StapelImageBase struct {
	Name             string
	From             string
	FromLatest       bool
	FromImageName    string
	FromArtifactName string
	FromCacheVersion string
	Git              *GitManager
	Shell            *Shell
	Ansible          *Ansible
	Mount            []*Mount
	Import           []*Import

	raw *rawStapelImage
}

func (c *StapelImageBase) GetName() string {
	return c.Name
}

func (c *StapelImageBase) imports() []*Import {
	return c.Import
}

func (c *StapelImageBase) ImageBaseConfig() *StapelImageBase {
	return c
}

func (c *StapelImageBase) IsArtifact() bool {
	return false
}

func (c *StapelImageBase) exportsAutoExcluding() error {
	for _, exp1 := range c.exports() {
		for _, exp2 := range c.exports() {
			if exp1 == exp2 {
				continue
			}

			if !exp1.AutoExcludeExportAndCheck(exp2) {
				errMsg := fmt.Sprintf("Conflict between imports!\n\n%s\n%s", dumpConfigSection(exp1.GetRaw()), dumpConfigSection(exp2.GetRaw()))
				return newDetailedConfigError(errMsg, nil, c.raw.doc)
			}
		}
	}

	return nil
}

func (c *StapelImageBase) exports() []autoExcludeExport {
	var exports []autoExcludeExport
	if c.Git != nil {
		for _, git := range c.Git.Local {
			exports = append(exports, git)
		}

		for _, git := range c.Git.Remote {
			exports = append(exports, git)
		}
	}

	for _, imp := range c.Import {
		exports = append(exports, imp)
	}

	return exports
}

func (c *StapelImageBase) validate() error {
	if c.FromLatest && !giterminism_inspector.LooseGiterminism {
		if err := giterminism_inspector.ReportConfigStapelFromLatest(context.Background()); err != nil {
			errMsg := "\n\n" + err.Error()
			return newDetailedConfigError(errMsg, nil, c.raw.doc)
		}
	}

	if c.From == "" && c.raw.FromImage == "" && c.raw.FromArtifact == "" && c.FromImageName == "" && c.FromArtifactName == "" {
		return newDetailedConfigError("`from: DOCKER_IMAGE`, `fromImage: IMAGE_NAME`, `fromArtifact: IMAGE_ARTIFACT_NAME` required!", nil, c.raw.doc)
	}

	mountByTo := map[string]bool{}
	for _, mount := range c.Mount {
		_, exist := mountByTo[mount.To]
		if exist {
			return newDetailedConfigError("conflict between mounts!", nil, c.raw.doc)
		}

		mountByTo[mount.To] = true
	}

	if !oneOrNone([]bool{c.From != "", c.raw.FromImage != "", c.raw.FromArtifact != ""}) {
		return newDetailedConfigError("conflict between `from`, `fromImage` and `fromArtifact` directives!", nil, c.raw.doc)
	}

	if c.raw.FromArtifact != "" {
		logboek.Context(context.Background()).Warn().LogLn("DEPRECATION WARNING: The directive `fromArtifact` will be removed in v1.3. Use `fromImage` or `import` directive instead!")
	}

	// TODO: валидацию формата `From`

	return nil
}
