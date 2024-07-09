package config

import (
	"context"

	"github.com/werf/werf/v2/pkg/werf/global_warnings"
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
		global_warnings.GlobalDeprecationWarningLn(context.Background(), "Support for the nameless image (`image: ~`) is deprecated and will be removed in v3!")
	}

	return nil
}
