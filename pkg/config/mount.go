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

func (c *Mount) validate() error {
	if c.To == "" || !isAbsolutePath(c.To) {
		return newDetailedConfigError("`to: PATH` absolute path required for mount!", c.raw, c.raw.rawDimg.doc)
	} else if c.Type == "custom_dir" {
		if (c.From != "" && isRelativePath(c.From)) || c.From == "" {
			return newDetailedConfigError("`fromPath: PATH` should be absolute path for mount!", c.raw, c.raw.rawDimg.doc)
		}
	} else if c.Type != "tmp_dir" && c.Type != "build_dir" {
		return newDetailedConfigError(fmt.Sprintf("Invalid `from: %s` for mount: expected `tmp_dir` or `build_dir`!", c.Type), c.raw, c.raw.rawDimg.doc)
	}
	return nil
}
