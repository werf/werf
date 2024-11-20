package config

import (
	"errors"
	"fmt"
	"os"
	"time"
)

type Secret interface {
	GetSecretStringArg() (string, error)
	GetSecretId() string
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

func (s *SecretFromEnv) GetSecretStringArg() (string, error) {
	if _, exists := os.LookupEnv(s.Value); !exists {
		return "", fmt.Errorf("specified env variable doesn't exist")
	}
	return fmt.Sprintf("id=%s,env=%s", s.Id, s.Value), nil
}

func (s *SecretFromSrc) GetSecretStringArg() (string, error) {
	if _, err := os.Stat(s.Value); errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("path %s doesn't exist", s.Value)
	}
	return fmt.Sprintf("id=%s,src=%s", s.Id, s.Value), nil
}

func (s *SecretFromPlainValue) GetSecretStringArg() (string, error) {
	secret, err := s.setPalinValueAsEnv()
	if err != nil {
		return "", err
	}
	return secret.GetSecretStringArg()
}

func (s *SecretFromPlainValue) setPalinValueAsEnv() (*SecretFromEnv, error) {
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
