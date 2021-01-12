package config

import (
	"context"

	"github.com/werf/logboek"
)

type StapelImage struct {
	*StapelImageBase
	Docker *Docker
}

func (c *StapelImage) validate() error {
	if !oneOrNone([]bool{c.Shell != nil, c.Ansible != nil}) {
		return newDetailedConfigError("can not use shell and ansible builders at the same time!", nil, c.StapelImageBase.raw.doc)
	}

	if c.Name == "" {
		logboek.Context(context.Background()).Warn().LogLn("DEPRECATION WARNING: Support for the nameless image, `image: ~`, will be removed in v1.3!")
	}

	return nil
}
