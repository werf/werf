package config

import (
	"context"
	"fmt"

	"github.com/werf/logboek"
)

type StapelImageBase struct {
	Name                                                string
	From                                                string
	FromLatest                                          bool
	HerebyIAdmitThatFromLatestMightBreakReproducibility bool
	FromImageName                                       string
	FromImageArtifactName                               string
	FromCacheVersion                                    string
	Git                                                 *GitManager
	Shell                                               *Shell
	Ansible                                             *Ansible
	Mount                                               []*Mount
	Import                                              []*Import

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
	if c.FromLatest && !c.HerebyIAdmitThatFromLatestMightBreakReproducibility {
		msg := `Pay attention, werf uses actual base image digest in stage signature if 'fromLatest' is specified. Thus, the usage of this directive might break the reproducibility of previous builds. If the base image is changed in the registry, all previously built stages become not usable.

* Previous pipeline jobs (e.g. deploy) cannot be retried without the image rebuild after changing base image in the registry.
* If base image is modified unexpectedly it might lead to the inexplicably failed pipeline. For instance, the modification occurs after successful build and the following jobs will be failed due to changing of stages signatures alongside base image digest.

If you still want to use this directive, add 'herebyIAdmitThatFromLatestMightBreakReproducibility: true' alongside 'fromLatest'.

We do not recommend using the actual base image such way. Use a particular unchangeable tag or periodically change 'fromCacheVersion' value to provide controllable and predictable lifecycle of software`

		msg = "\n\n" + msg

		return newDetailedConfigError(msg, nil, c.raw.doc)
	}

	if c.From == "" && c.raw.FromImage == "" && c.raw.FromImageArtifact == "" && c.FromImageName == "" && c.FromImageArtifactName == "" {
		return newDetailedConfigError("`from: DOCKER_IMAGE`, `fromImage: IMAGE_NAME`, `fromImageArtifact: IMAGE_ARTIFACT_NAME` required!", nil, c.raw.doc)
	}

	mountByTo := map[string]bool{}
	for _, mount := range c.Mount {
		_, exist := mountByTo[mount.To]
		if exist {
			return newDetailedConfigError("conflict between mounts!", nil, c.raw.doc)
		}

		mountByTo[mount.To] = true
	}

	if !oneOrNone([]bool{c.From != "", c.raw.FromImage != "", c.raw.FromImageArtifact != ""}) {
		return newDetailedConfigError("conflict between `from`, `fromImage` and `fromImageArtifact` directives!", nil, c.raw.doc)
	}

	if c.raw.FromImageArtifact != "" {
		logboek.Context(context.Background()).Warn().LogLn("WARNING: Do not use artifacts as a base for other images and artifacts. The feature is deprecated, and the directive 'fromImageArtifact' will be completely removed in version v1.3.\n\nCareless use of artifacts may lead to difficult to trace issues that may arise long after the configuration has been written. The artifact image is cached after the first build and ignores any changes in the project git repository unless the user has explicitly specified stage dependencies. As found, this behavior is completely unexpected for users despite the fact that it is absolutely correct in the werf logic.")
	}

	// TODO: валидацию формата `From`

	return nil
}
