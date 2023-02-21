package secrets_manager

import (
	"context"
	"fmt"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/secret"
)

type SecretsManager struct {
	DisableSecretsDecryption bool

	missedSecretKeyModeEnabled bool
}

type SecretsManagerOptions struct {
	DisableSecretsDecryption bool
}

func NewSecretsManager(opts SecretsManagerOptions) *SecretsManager {
	return &SecretsManager{
		DisableSecretsDecryption: opts.DisableSecretsDecryption,
	}
}

func (manager *SecretsManager) IsMissedSecretKeyModeEnabled() bool {
	return manager.missedSecretKeyModeEnabled
}

func (manager *SecretsManager) AllowMissedSecretKeyMode(workingDir string) error {
	_, err := GetRequiredSecretKey(workingDir)
	if err != nil {
		if _, missedKey := err.(*EncryptionKeyRequiredError); missedKey {
			manager.missedSecretKeyModeEnabled = true
			return nil
		}
		return fmt.Errorf("unable to load secret key: %w", err)
	}
	return nil
}

func (manager *SecretsManager) GetYamlEncoder(ctx context.Context, workingDir string) (*secret.YamlEncoder, error) {
	if manager.DisableSecretsDecryption {
		logboek.Context(ctx).Default().LogLnDetails("Secrets decryption disabled")
		return secret.NewYamlEncoder(nil), nil
	}
	if manager.missedSecretKeyModeEnabled {
		logboek.Context(ctx).Error().LogLn("Secrets decryption disabled due to missed key (no WERF_SECRET_KEY is set)")
		return secret.NewYamlEncoder(nil), nil
	}

	if key, err := GetRequiredSecretKey(workingDir); err != nil {
		return nil, fmt.Errorf("unable to load secret key: %w", err)
	} else if enc, err := secret.NewAesEncoder(key); err != nil {
		return nil, fmt.Errorf("check encryption key: %w", err)
	} else {
		return secret.NewYamlEncoder(enc), nil
	}
}

func (manager *SecretsManager) GetYamlEncoderForOldKey(ctx context.Context) (*secret.YamlEncoder, error) {
	if key, err := GetRequiredOldSecretKey(); err != nil {
		return nil, fmt.Errorf("unable to load old secret key: %w", err)
	} else if enc, err := secret.NewAesEncoder(key); err != nil {
		return nil, fmt.Errorf("check old encryption key: %w", err)
	} else {
		return secret.NewYamlEncoder(enc), nil
	}
}
