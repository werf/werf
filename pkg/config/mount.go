package config

import (
	"fmt"

	"github.com/werf/werf/pkg/giterminism_manager"
)

type Mount struct {
	To   string
	From string
	Type string

	raw *rawMount
}

func (c *Mount) validate(giterminismManager giterminism_manager.Interface) error {
	var err error
	if c.raw.FromPath != "" {
		err = giterminismManager.Inspector().InspectConfigStapelMountFromPath(c.raw.FromPath)
	} else if c.Type == "build_dir" {
		err = giterminismManager.Inspector().InspectConfigStapelMountBuildDir()
	}

	if err != nil {
		return newDetailedConfigError(err.Error(), c.raw, c.raw.rawStapelImage.doc)
	}

	if c.raw.From != "" && c.raw.FromPath != "" {
		return newDetailedConfigError(fmt.Sprintf("cannot use `from: %s` and `fromPath: %s` at the same time for mount!", c.raw.From, c.raw.FromPath), c, c.raw.rawStapelImage.doc)
	}

	switch {
	case c.To == "" || !isAbsolutePath(c.To):
		return newDetailedConfigError("`to: PATH` absolute path required for mount!", c.raw, c.raw.rawStapelImage.doc)
	case c.Type == "custom_dir":
		if c.From == "" {
			return newDetailedConfigError("`fromPath: PATH` absolute or relative path required for mount!", c.raw, c.raw.rawStapelImage.doc)
		}
	case c.Type != "tmp_dir" && c.Type != "build_dir":
		return newDetailedConfigError(fmt.Sprintf("invalid `from: %s` for mount: expected `tmp_dir` or `build_dir`!", c.Type), c.raw, c.raw.rawStapelImage.doc)
	}

	return nil
}
