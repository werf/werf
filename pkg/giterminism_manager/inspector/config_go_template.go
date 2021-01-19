package inspector

import (
	"context"
	"fmt"
)

func (i Inspector) InspectConfigGoTemplateRenderingEnv(ctx context.Context, envName string) error {
	if i.sharedOptions.LooseGiterminism() {
		return nil
	}

	if isAccepted, err := i.giterminismConfig.IsConfigGoTemplateRenderingEnvNameAccepted(envName); err != nil {
		return err
	} else if isAccepted {
		return nil
	}

	return NewExternalDependencyFoundError(fmt.Sprintf(`env name '%s' not allowed`, envName))
}
