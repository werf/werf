package config

import (
	"fmt"
)

type Mount struct {
	To   string
	From string
	Type string

	raw *rawMount
}

func (c *Mount) validate(disableDeterminism bool) error {
	if !disableDeterminism {
		return fmt.Errorf("'mount' directive is forbidden, it is recommended to avoid this directive, otherwise disable werf determinism mode with option --disable-determinism (or WERF_DISABLE_DETERMINISM=1 environment variable)")
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
