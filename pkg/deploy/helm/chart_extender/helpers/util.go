package helpers

import (
	"context"
	"os"

	"sigs.k8s.io/yaml"

	"github.com/werf/logboek"
)

func DebugPrintValues(ctx context.Context, name string, vals map[string]interface{}) {
	data, err := yaml.Marshal(vals)
	if err != nil {
		logboek.Context(ctx).Debug().LogF("Unable to marshal %q values: %s\n", err)
	} else {
		logboek.Context(ctx).Debug().LogF("%q values:\n%s---\n", name, data)
	}
}

func DebugSecretValues() bool {
	return os.Getenv("WERF_DEBUG_SECRET_VALUES") == "1"
}
