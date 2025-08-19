package config

import (
	"context"
	"fmt"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/giterminism_manager"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

type StapelImageBase struct {
	Name             string
	From             string
	FromLatest       bool
	FromArtifactName string
	FromCacheVersion string
	Git              *GitManager
	Shell            *Shell
	Ansible          *Ansible
	Mount            []*Mount
	Import           []*Import
	Dependencies     []*Dependency
	Secrets          []Secret
	ImageSpec        *ImageSpec

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
	if c.FromArtifactName != "" {
		return c.FromArtifactName
	}
	return c.From
}

func (c *StapelImageBase) SetFromExternal() {
	c.FromExternal = true
}

func (c *StapelImageBase) dependsOn() DependsOn {
	var dependsOn DependsOn

	if c.FromArtifactName != "" {
		dependsOn.From = c.FromArtifactName
	}

	for _, imp := range c.Import {
		if imp.ImageName != "" && !imp.ExternalImage {
			dependsOn.Imports = append(dependsOn.Imports, imp.ImageName)
		}

		if imp.ArtifactName != "" {
			dependsOn.Imports = append(dependsOn.Imports, imp.ArtifactName)
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

	if c.From == "" && c.raw.FromImage == "" && c.raw.FromArtifact == "" && c.FromArtifactName == "" {
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

	if !oneOrNone([]bool{c.From != "", c.raw.FromArtifact != ""}) {
		return newDetailedConfigError("conflict between `from`, `fromImage` and `fromArtifact` directives!", nil, c.raw.doc)
	}

	if c.raw.FromArtifact != "" {
		printArtifactDepricationWarning()
	}

	// TODO: валидацию формата `From`

	return nil
}

var isArtifactDepricationWarningPrinted bool

func printArtifactDepricationWarning() {
	if isArtifactDepricationWarningPrinted {
		return
	}

	global_warnings.GlobalDeprecationWarningLn(context.Background(), `The 'artifact', 'fromArtifact' and 'import.artifact' directives are deprecated and will be completely removed in version v3.

Instead of 'artifact', use the 'image' directive:

- If you need to preserve artifact behavior when working with 'git' (disabling source updates by skipping the 'gitCache' and 'gitLatestPatch' stages), use the 'disableGitAfterPatch' directive:

    '''
    image: builder
    from: alpine:3.10
    disableGitAfterPatch: true
    git:
    - add: /
      to: /app
    '''

- If you simply need to limit the scope of image use, use the 'final' directive:

    '''
    image: builder
    from: alpine:3.10
    final: false
    '''`)

	isArtifactDepricationWarningPrinted = true
}
