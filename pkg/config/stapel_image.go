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
	if c.Docker != nil && c.ImageSpec != nil {
		return newDetailedConfigError("`docker` directive is deprecated and can't be used along with `imageSpec`!", nil, c.StapelImageBase.raw.doc)
	}

	if c.Name == "" {
		global_warnings.GlobalDeprecationWarningLn(context.Background(), "Support for the nameless image (`image: ~`) is deprecated and will be removed in v3!")
	}

	return nil
}
