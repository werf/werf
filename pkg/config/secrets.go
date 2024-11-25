package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/werf/werf/v2/pkg/giterminism_manager"
)

type Secret interface {
	GetSecretStringArg() (string, error)
	GetSecretId() string
	InspectByGiterminism(giterminismManager giterminism_manager.Interface) error
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
	if s.Id == "" {
		s.Id = s.Env
	}
	return &SecretFromEnv{
		Id:    s.Id,
		Value: s.Env,
	}, nil
}

func newSecretFromSrc(s *rawSecret) (*SecretFromSrc, error) {
	if s.Id == "" {
		s.Id = filepath.Base(s.Src)
	}
	return &SecretFromSrc{
		Id:    s.Id,
		Value: s.Src,
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
	if _, exists := os.LookupEnv(s.Value); !exists {
		return "", fmt.Errorf("specified secret env %q doesn't exist", s.Value)
	}
	return fmt.Sprintf("id=%s,env=%s", s.Id, s.Value), nil
}

func (s *SecretFromSrc) GetSecretStringArg() (string, error) {
	if _, err := os.Stat(s.Value); errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("specified secret path %s doesn't exist", s.Value)
	}
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
	t := time.Now().Unix()
	envKey := fmt.Sprintf("tmpbuild%d_%s", t, s.Id) // generate unique value
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
