package config

import (
	"context"

	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

type Docker struct {
	Volume      []string
	Expose      []string
	Env         map[string]string
	Label       map[string]string
	Cmd         string
	Workdir     string
	User        string
	Entrypoint  string
	HealthCheck string

	ExactValues bool

	raw *rawDocker
}

func (c *Docker) validate() error {
	global_warnings.GlobalDeprecationWarningLn(context.Background(), "The `docker` directive is deprecated and will be removed in v3. Please use the `imageSpec` directive instead.")

	return nil
}
