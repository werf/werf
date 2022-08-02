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

	return NewExternalDependencyFoundError(fmt.Sprintf(`env name %q not allowed by giterminism

The use of the function env complicates the sharing and reproducibility of the configuration in CI jobs and among developers, because the value of the environment variable affects the final digest of built images.`, envName))
}
