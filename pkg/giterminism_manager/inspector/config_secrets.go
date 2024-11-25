package inspector

import (
	"fmt"
)

var secretsErrMsg = `secret %q is not allowed by giterminism

	Using env and file secrets complicates the sharing and reproducibility of the configuration in CI jobs and among developers.`

func (i Inspector) InspectConfigSecretEnvAccepted(secret string) error {
	if i.sharedOptions.LooseGiterminism() {
		return nil
	}

	if i.giterminismConfig.IsConfigSecretsEnvNameAccepted(secret) {
		return nil
	}

	return NewExternalDependencyFoundError(fmt.Sprintf(secretsErrMsg, secret))
}

func (i Inspector) InspectConfigSecretSrcAccepted(secret string) error {
	if i.sharedOptions.LooseGiterminism() {
		return nil
	}

	if i.giterminismConfig.IsConfigSecretsFileAccepted(secret) {
		return nil
	}

	return NewExternalDependencyFoundError(fmt.Sprintf(secretsErrMsg, secret))
}
