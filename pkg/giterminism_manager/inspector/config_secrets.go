package inspector

import (
	"fmt"
)

var secretsErrMsg = `secret %q not allowed by giterminism

	The use of the function env complicates the sharing and reproducibility of the configuration in CI jobs and among developers, because the value of the secret affects the final digest of built images.`

func (i Inspector) InspectConfigSecretsEnvNameAccepted(secret string) error {
	if i.sharedOptions.LooseGiterminism() {
		return nil
	}

	if i.giterminismConfig.IsConfigSecretsEnvNameAccepted(secret) {
		return nil
	}

	return NewExternalDependencyFoundError(fmt.Sprintf(secretsErrMsg, secret))
}

func (i Inspector) InspectConfigSecretsAllowPathsSecretsAccepted(secret string) error {
	if i.sharedOptions.LooseGiterminism() {
		return nil
	}

	if i.giterminismConfig.IsConfigSecretsAllowPathsSecretsAccepted(secret) {
		return nil
	}

	return NewExternalDependencyFoundError(fmt.Sprintf(secretsErrMsg, secret))
}
