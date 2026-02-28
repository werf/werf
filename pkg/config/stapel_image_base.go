package config

import (
	"fmt"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/giterminism_manager"
)

type StapelImageBase struct {
	Name             string
	From             string
	FromLatest       bool
	FromCacheVersion string
	Git              *GitManager
	Shell            *Shell
	Mount            []*Mount
	Import           []*Import
	Dependencies     []*Dependency
	Secrets          []Secret
	ImageSpec        *ImageSpec
	Network          string

	FromExternal bool
	cacheVersion string
	final        bool
	platform     []string
	raw          *rawStapelImage
}

func (c *StapelImageBase) CacheVersion() string {
	return c.cacheVersion
}

func (c *StapelImageBase) GetName() string {
	return c.Name
}

func (c *StapelImageBase) IsStapel() bool {
	return true
}

func (c *StapelImageBase) ImageBaseConfig() *StapelImageBase {
	return c
}

func (c *StapelImageBase) IsGitAfterPatchDisabled() bool {
	if c.Git == nil {
		return false
	}

	return c.Git.isGitAfterPatchDisabled
}

func (c *StapelImageBase) IsFinal() bool {
	return c.final
}

func (c *StapelImageBase) Platform() []string {
	return c.platform
}

func (c *StapelImageBase) GetFrom() string {
	return c.From
}

func (c *StapelImageBase) SetFromExternal() {
	c.FromExternal = true
}

func (c *StapelImageBase) dependsOn() DependsOn {
	var dependsOn DependsOn

	for _, imp := range c.Import {
		if imp.ImageName != "" {
			dependsOn.Imports = append(dependsOn.Imports, imp.ImageName)
		}
	}

	for _, dep := range c.Dependencies {
		dependsOn.Dependencies = append(dependsOn.Dependencies, dep.ImageName)
	}
	dependsOn.Dependencies = util.UniqStrings(dependsOn.Dependencies)

	return dependsOn
}

func (c *StapelImageBase) rawDoc() *doc {
	return c.raw.doc
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

func (c *StapelImageBase) validate(giterminismManager giterminism_manager.Interface) error {
	if c.FromLatest {
		if err := giterminismManager.Inspector().InspectConfigStapelFromLatest(); err != nil {
			return newDetailedConfigError(err.Error(), nil, c.raw.doc)
		}
	}

	if c.From == "" && c.raw.FromImage == "" {
		return newDetailedConfigError("`from: IMAGE` required!", nil, c.raw.doc)
	}

	if c.Name != "" && c.From == c.Name {
		return newDetailedConfigError("image \""+c.Name+"\" cannot use itself as base image in 'from' directive", nil, c.raw.doc)
	}

	mountByTo := map[string]bool{}
	for _, mount := range c.Mount {
		_, exist := mountByTo[mount.To]
		if exist {
			return newDetailedConfigError("conflict between mounts!", nil, c.raw.doc)
		}

		mountByTo[mount.To] = true
	}

	// TODO: валидацию формата `From`

	return nil
}
