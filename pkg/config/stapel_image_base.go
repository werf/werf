package config

import (
	"fmt"
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

func (c *StapelImageBase) validate(disableGitermenism bool) error {
	if c.FromLatest && !disableGitermenism {
		msg := `Pay attention, werf uses actual base image digest in stage digest if 'fromLatest' is specified. Thus, the usage of this directive might break the reproducibility of previous builds. If the base image is changed in the registry, all previously built stages become not usable.

* Previous pipeline jobs (e.g. deploy) cannot be retried without the image rebuild after changing base image in the registry.
* If base image is modified unexpectedly it might lead to the inexplicably failed pipeline. For instance, the modification occurs after successful build and the following jobs will be failed due to changing of stages digests alongside base image digest.

If you still want to use this directive, then disable werf gitermenism mode with option --disable-gitermenism (or WERF_DISABLE_GITERMENISM=1 env var). However it is NOT RECOMMENDED to use the actual base image in a such way. Use a particular unchangeable tag or periodically change 'fromCacheVersion' value to provide controllable and predictable lifecycle of software.`

		msg = "\n\n" + msg

		return newDetailedConfigError(msg, nil, c.raw.doc)
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

	// TODO: валидацию формата `From`
	// TODO: валидация формата `Name`

	return nil
}
