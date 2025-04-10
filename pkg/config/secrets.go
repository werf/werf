package config

import (
	"fmt"
	"path/filepath"

	"github.com/werf/werf/v2/pkg/giterminism_manager"
)

type Secret struct {
	Id             string
	ValueFromEnv   string
	ValueFromSrc   string
	ValueFromPlain string
}

func newSecretFromEnv(s *rawSecret) (Secret, error) {
	if s.Id == "" {
		s.Id = s.Env
	}
	return Secret{
		Id:           s.Id,
		ValueFromEnv: s.Env,
	}, nil
}

func newSecretFromSrc(s *rawSecret) (Secret, error) {
	if s.Id == "" {
		s.Id = filepath.Base(s.Src)
	}
	return Secret{
		Id:           s.Id,
		ValueFromSrc: s.Src,
	}, nil
}

func newSecretFromPlainValue(s *rawSecret) (Secret, error) {
	if s.Id == "" {
		return Secret{}, fmt.Errorf("type value should be used with id parameter")
	}
	return Secret{
		Id:             s.Id,
		ValueFromPlain: s.PlainValue,
	}, nil
}

func inspectSecretByGiterminism(giterminismManager giterminism_manager.Interface, secret Secret) error {
	if secret.ValueFromEnv != "" {
		return giterminismManager.Inspector().InspectConfigSecretEnvAccepted(secret.ValueFromEnv)
	} else if secret.ValueFromSrc != "" {
		return giterminismManager.Inspector().InspectConfigSecretSrcAccepted(secret.ValueFromSrc)
	} else if secret.ValueFromPlain != "" {
		return giterminismManager.Inspector().InspectConfigSecretValueAccepted(secret.Id)
	}
	return nil
}

func GetValidatedSecrets(rawSecrets []*rawSecret, giterminismManager giterminism_manager.Interface, doc *doc) ([]Secret, error) {
	secretIds := make(map[string]struct{})
	secrets := make([]Secret, 0, len(rawSecrets))

	for _, s := range rawSecrets {
		secret, err := s.toDirective()
		if err != nil {
			return nil, newDetailedConfigError(fmt.Sprintf("unable to load build secrets: %s", err.Error()), s, s.parent.getDoc())
		}

		secretId := secret.Id
		if _, ok := secretIds[secretId]; !ok {
			secretIds[secretId] = struct{}{}
		} else {
			return nil, newDetailedConfigError(fmt.Sprintf("duplicated secret %q", secretId), nil, s.parent.getDoc())
		}

		err = inspectSecretByGiterminism(giterminismManager, secret)
		if err != nil {
			return nil, err
		}

		secrets = append(secrets, secret)
	}

	return secrets, nil
}
