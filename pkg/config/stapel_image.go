package config

import (
	"context"

	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

type StapelImage struct {
	*StapelImageBase
}

func (c *StapelImage) validate() error {
	if c.Name == "" {
		global_warnings.GlobalDeprecationWarningLn(context.Background(), "Support for the nameless image (`image: ~`) is deprecated and will be removed in v3!")
	}

	return nil
}
