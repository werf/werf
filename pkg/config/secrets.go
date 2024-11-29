package config

import (
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"

	"github.com/werf/kubedog/pkg/utils"
	"github.com/werf/werf/v2/pkg/giterminism_manager"
	"github.com/werf/werf/v2/pkg/util"
)

type Secret interface {
	GetSecretStringArg() (string, error)
	GetSecretId() string
	InspectByGiterminism(giterminismManager giterminism_manager.Interface) error
	GetMountPath(stageHostTmpDir string) (string, error)
}

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

func newSecretFromEnv(s *rawSecret) (*SecretFromEnv, error) {
	if _, exists := os.LookupEnv(s.Env); !exists {
		return nil, fmt.Errorf("specified env variable doesn't exist")
	}
	if s.Id == "" {
		s.Id = s.Env
	}
	return &SecretFromEnv{
		Id:    s.Id,
		Value: s.Env,
	}, nil
}

func newSecretFromSrc(s *rawSecret) (*SecretFromSrc, error) {
	absPath, err := util.ExpandPath(s.Src)
	if err != nil {
		return nil, fmt.Errorf("error load secret from src: %w", err)
	}

	if exists, _ := utils.FileExists(absPath); !exists {
		return nil, fmt.Errorf("error load secret from src: path %s doesn't exist", absPath)
	}

	if s.Id == "" {
		s.Id = filepath.Base(absPath)
	}
	return &SecretFromSrc{
		Id:    s.Id,
		Value: absPath,
	}, nil
}

func newSecretFromPlainValue(s *rawSecret) (*SecretFromPlainValue, error) {
	if s.Id == "" {
		return nil, fmt.Errorf("type value should be used with id parameter")
	}
	return &SecretFromPlainValue{
		Id:    s.Id,
		Value: s.PlainValue,
	}, nil
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

func (s *SecretFromEnv) GetSecretId() string {
	return s.Id
}

func (s *SecretFromSrc) GetSecretId() string {
	return s.Id
}

func (s *SecretFromPlainValue) GetSecretId() string {
	return s.Id
}

func (s *SecretFromEnv) InspectByGiterminism(giterminismManager giterminism_manager.Interface) error {
	return giterminismManager.Inspector().InspectConfigSecretEnvAccepted(s.Value)
}

func (s *SecretFromSrc) InspectByGiterminism(giterminismManager giterminism_manager.Interface) error {
	return giterminismManager.Inspector().InspectConfigSecretSrcAccepted(s.Value)
}

func (s *SecretFromPlainValue) InspectByGiterminism(giterminismManager giterminism_manager.Interface) error {
	return nil
}

func GetValidatedSecrets(rawSecrets []*rawSecret, giterminismManager giterminism_manager.Interface, doc *doc) ([]Secret, error) {
	secretIds := make(map[string]struct{})
	secrets := make([]Secret, 0, len(rawSecrets))

	for _, s := range rawSecrets {
		secret, err := s.toDirective()
		if err != nil {
			return nil, err
		}

		secretId := secret.GetSecretId()
		if _, ok := secretIds[secretId]; !ok {
			secretIds[secretId] = struct{}{}
		} else {
			return nil, newDetailedConfigError("duplicated secret %s", secretId, s.doc)
		}

		err = secret.InspectByGiterminism(giterminismManager)
		if err != nil {
			return nil, err
		}

		secrets = append(secrets, secret)
	}

	return secrets, nil
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
