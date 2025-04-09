package secrets

import (
	"fmt"
	"math"
	"math/rand/v2"
	"os"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/config"
)

type SecretFromEnv struct {
	Id    string
	Value string
}

type SecretFromSrc struct {
	Id    string
	Value string
}

type SecretFromPlainValue struct {
	Id    string
	Value string
}

type Secret interface {
	GetSecretStringArg() (string, error)
	GetMountPath(stageHostTmpDir string) (string, error)
}

func GetSecretStringArg(secret config.Secret) (string, error) {
	s, err := parseSecret(secret)
	if err != nil {
		return "", fmt.Errorf("error parsing secrets: %w", err)
	}
	return s.GetSecretStringArg()
}

func (s *SecretFromEnv) GetSecretStringArg() (string, error) {
	return fmt.Sprintf("id=%s,env=%s", s.Id, s.Value), nil
}

func (s *SecretFromSrc) GetSecretStringArg() (string, error) {
	return fmt.Sprintf("id=%s,src=%s", s.Id, s.Value), nil
}

func (s *SecretFromPlainValue) GetSecretStringArg() (string, error) {
	secret, err := s.setPlainValueAsEnv()
	if err != nil {
		return "", err
	}
	return secret.GetSecretStringArg()
}

func (s *SecretFromPlainValue) setPlainValueAsEnv() (*SecretFromEnv, error) {
	envKey := fmt.Sprintf("tmpbuild%d_%s", rand.IntN(math.MaxInt32), s.Id) // generate unique value
	if _, e := os.LookupEnv(envKey); e {
		return nil, fmt.Errorf("can't set secret %s: id is not unique", s.Id) // should never be here
	}

	err := os.Setenv(envKey, s.Value)
	if err != nil {
		return nil, fmt.Errorf("can't set value")
	}

	return &SecretFromEnv{
		Id:    s.Id,
		Value: envKey,
	}, nil
}

func GetMountPath(secret config.Secret, stageHostTmpDir string) (string, error) {
	s, err := parseSecret(secret)
	if err != nil {
		return "", fmt.Errorf("unable to get secret mount path: %w", err)
	}
	return s.GetMountPath(stageHostTmpDir)
}

func parseSecret(secret config.Secret) (Secret, error) {
	if secret.ValueFromEnv != "" {
		return newSecretFromEnv(secret)
	} else if secret.ValueFromSrc != "" {
		return newSecretFromSrc(secret)
	} else if secret.ValueFromPlain != "" {
		return newSecretFromPlainValue(secret)
	}
	return nil, fmt.Errorf("unknown secret type")
}

func newSecretFromEnv(s config.Secret) (*SecretFromEnv, error) {
	if _, exists := os.LookupEnv(s.ValueFromEnv); !exists {
		return nil, fmt.Errorf("specified env variable `%s` is not set", s.ValueFromEnv)
	}
	return &SecretFromEnv{Id: s.Id, Value: s.ValueFromEnv}, nil
}

func newSecretFromSrc(s config.Secret) (*SecretFromSrc, error) {
	absPath, err := util.ExpandPath(s.ValueFromSrc)
	if err != nil {
		return nil, fmt.Errorf("error load secret from src: %w", err)
	}

	if exists, _ := util.FileExists(absPath); !exists {
		return nil, fmt.Errorf("error load secret from src: path %s doesn't exist", absPath)
	}
	return &SecretFromSrc{Id: s.Id, Value: absPath}, nil
}

func newSecretFromPlainValue(s config.Secret) (*SecretFromPlainValue, error) {
	return &SecretFromPlainValue{Id: s.Id, Value: s.ValueFromPlain}, nil
}

func (s *SecretFromEnv) GetMountPath(stageHostTmpDir string) (string, error) {
	data := []byte(os.Getenv(s.Value))
	return getMountPath(s.Id, stageHostTmpDir, data)
}

func (s *SecretFromSrc) GetMountPath(stageHostTmpDir string) (string, error) {
	return generateMountPath(s.Id, s.Value), nil
}

func (s *SecretFromPlainValue) GetMountPath(stageHostTmpDir string) (string, error) {
	return getMountPath(s.Id, stageHostTmpDir, []byte(s.Value))
}

func getMountPath(secretId, stageHostTmpDir string, data []byte) (string, error) {
	tmpFile, err := writeToTmpFile(stageHostTmpDir, data)
	if err != nil {
		return "", fmt.Errorf("unable to mount secret: %w", err)
	}
	return generateMountPath(secretId, tmpFile), nil
}

func writeToTmpFile(stageHostTmpDir string, data []byte) (string, error) {
	tmpFile, err := os.CreateTemp(stageHostTmpDir, "stapel*")
	if err != nil {
		return "", err
	}

	tmpFilePath := tmpFile.Name()

	if err := os.WriteFile(tmpFilePath, data, 0o400); err != nil {
		return "", err
	}

	return tmpFilePath, nil
}

func generateMountPath(id, filepath string) string {
	containerPath := fmt.Sprintf("/run/secrets/%s", id)
	return fmt.Sprintf("%s:%s:ro", filepath, containerPath)
}
