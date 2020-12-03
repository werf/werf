package config

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/giterminism_inspector"
)

type Mount struct {
	To   string
	From string
	Type string

	raw *rawMount
}

func (c *Mount) validate(disableGiterminism bool) error {
	if !disableGiterminism {
		if err := giterminism_inspector.ReportMountDirectiveUsage(context.Background()); err != nil {
			return err
		}
	}

	if c.raw.From != "" && c.raw.FromPath != "" {
		return newDetailedConfigError(fmt.Sprintf("cannot use `from: %s` and `fromPath: %s` at the same time for mount!", c.raw.From, c.raw.FromPath), c, c.raw.rawStapelImage.doc)
	}

	if c.To == "" || !isAbsolutePath(c.To) {
		return newDetailedConfigError("`to: PATH` absolute path required for mount!", c.raw, c.raw.rawStapelImage.doc)
	} else if c.Type == "custom_dir" {
		if c.From == "" {
			return newDetailedConfigError("`fromPath: PATH` absolute or relative path required for mount!", c.raw, c.raw.rawStapelImage.doc)
		}
	} else if c.Type != "tmp_dir" && c.Type != "build_dir" {
		return newDetailedConfigError(fmt.Sprintf("invalid `from: %s` for mount: expected `tmp_dir` or `build_dir`!", c.Type), c.raw, c.raw.rawStapelImage.doc)
	}
	return nil
}
