package secrets_manager

import (
	"context"
	"fmt"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/secret"
)

type SecretsManager struct {
	DisableSecretsDecryption bool
}

type SecretsManagerOptions struct {
	DisableSecretsDecryption bool
}

func NewSecretsManager(opts SecretsManagerOptions) *SecretsManager {
	return &SecretsManager{
		DisableSecretsDecryption: opts.DisableSecretsDecryption,
	}
}

func (manager *SecretsManager) GetYamlEncoder(ctx context.Context, workingDir string) (*secret.YamlEncoder, error) {
	if manager.DisableSecretsDecryption {
		logboek.Context(ctx).Default().LogLnDetails("Secrets decryption disabled")
		return secret.NewYamlEncoder(nil), nil
	}

	if key, err := GetRequiredSecretKey(workingDir); err != nil {
		return nil, fmt.Errorf("unable to load secret key: %s", err)
	} else if enc, err := secret.NewAesEncoder(key); err != nil {
		return nil, fmt.Errorf("check encryption key: %s", err)
	} else {
		return secret.NewYamlEncoder(enc), nil
	}
}

func (manager *SecretsManager) GetYamlEncoderForOldKey(ctx context.Context) (*secret.YamlEncoder, error) {
	if key, err := GetRequiredOldSecretKey(); err != nil {
		return nil, fmt.Errorf("unable to load old secret key: %s", err)
	} else if enc, err := secret.NewAesEncoder(key); err != nil {
		return nil, fmt.Errorf("check old encryption key: %s", err)
	} else {
		return secret.NewYamlEncoder(enc), nil
	}
}
