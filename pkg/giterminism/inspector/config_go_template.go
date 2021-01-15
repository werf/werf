package inspector

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/giterminism/errors"
)

func (i Inspector) InspectConfigGoTemplateRenderingEnv(ctx context.Context, envName string) error {
	if i.manager.LooseGiterminism() {
		return nil
	}

	if isAccepted, err := i.manager.Config().IsConfigGoTemplateRenderingEnvNameAccepted(envName); err != nil {
		return err
	} else if isAccepted {
		return nil
	}

	return errors.NewError(fmt.Sprintf(`the configuration with external dependency found in the werf config: env name '%s' not allowed`, envName))
}
