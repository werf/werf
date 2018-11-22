package secret

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/flant/dapp/pkg/dapp"
	"github.com/flant/dapp/pkg/secret"

	"k8s.io/kubernetes/pkg/util/file"
)

type Secret interface {
	secret.Secret

	GenerateYamlData(data []byte) ([]byte, error)
	ExtractYamlData(encodedData []byte) ([]byte, error)
}

func GenerateSecretKey() ([]byte, error) {
	return secret.GenerateAexSecretKey()
}

func GetSecret(projectDir string) (Secret, error) {
	var s Secret
	var ss secret.Secret
	var key []byte
	var err error

	key, err = getSecretKey(projectDir)
	if err != nil {
		return nil, err
	}

	if len(key) != 0 {
		s, err = NewSecretByKey(key)
		if err != nil {
			return nil, err
		}
	} else {
		s, err = NewSecret(ss)
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

func getSecretKey(projectDir string) ([]byte, error) {
	var secretKey []byte
	var dappSecretKeyPaths []string
	var notFoundIn []string

	secretKey = []byte(os.Getenv("DAPP_SECRET_KEY"))
	if len(secretKey) == 0 {
		notFoundIn = append(notFoundIn, "$DAPP_SECRET_KEY")

		var dappSecretKeyPath string

		projectDappSecretKeyPath, err := filepath.Abs(filepath.Join(projectDir, ".dapp_secret_key"))
		if err != nil {
			return nil, err
		}

		homeDappSecretKeyPath := filepath.Join(dapp.GetHomeDir(), ".dapp_secret_key")

		dappSecretKeyPaths = []string{
			projectDappSecretKeyPath,
			homeDappSecretKeyPath,
		}

		for _, path := range dappSecretKeyPaths {
			exist, err := file.FileExists(path)
			if err != nil {
				return nil, err
			}

			if exist {
				dappSecretKeyPath = path
				break
			} else {
				notFoundIn = append(notFoundIn, fmt.Sprintf("%s", path))
			}
		}

		if dappSecretKeyPath != "" {
			data, err := ioutil.ReadFile(dappSecretKeyPath)
			if err != nil {
				return nil, err
			}

			secretKey = []byte(strings.TrimSpace(string(data)))
		}
	}

	if len(secretKey) == 0 {
		return nil, fmt.Errorf("encryption key not found in: '%s'", strings.Join(notFoundIn, "', '"))
	}

	return secretKey, nil
}

func NewSecretByKey(key []byte) (Secret, error) {
	ss, err := secret.NewSecret(key)
	if err != nil {
		return nil, fmt.Errorf("check encryption key: %s", err)
	}

	return NewSecret(ss)
}

func NewSecret(ss secret.Secret) (Secret, error) {
	s := &BaseSecret{}

	if ss != nil {
		s.generateFunc = ss.Generate
		s.extractFunc = ss.Extract
	} else {
		s.generateFunc = doNothing
		s.extractFunc = doNothing
	}

	return s, nil
}
