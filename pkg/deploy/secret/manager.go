package secret

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/flant/dapp/pkg/dapp"
	"github.com/flant/dapp/pkg/logger"
	"github.com/flant/dapp/pkg/secret"

	"k8s.io/kubernetes/pkg/util/file"
)

type Manager interface {
	secret.Secret

	GenerateYamlData(data []byte) ([]byte, error)
	ExtractYamlData(encodedData []byte) ([]byte, error)
}

func GenerateSecretKey() ([]byte, error) {
	return secret.GenerateAexSecretKey()
}

func GetManager(projectDir string) (Manager, error) {
	var m Manager
	var key []byte
	var err error

	key, err = GetSecretKey(projectDir)
	if err != nil {
		return nil, err
	}

	m, err = NewManager(key, NewManagerOptions{})
	if err != nil {
		return nil, err
	}

	return m, nil
}

func GetSecretKey(projectDir string) ([]byte, error) {
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

type NewManagerOptions struct {
	IgnoreWarning bool
}

func NewManager(key []byte, options NewManagerOptions) (Manager, error) {
	ss, err := secret.NewSecret(key)
	if err != nil {
		if strings.HasPrefix(err.Error(), "encoding/hex:") {
			if !options.IgnoreWarning {
				logger.LogWarning(`
###################################################################################################
###                       WARNING invalid encryption key, do regenerate!                        ###
### https://flant.github.io/dapp/reference/deploy/secrets.html#regeneration-of-existing-secrets ###
###################################################################################################`)
			}

			return NewManager(ruby2GoSecretKey(key), options)
		}

		return nil, fmt.Errorf("check encryption key: %s", err)
	}

	return newBaseManager(ss)
}

func ruby2GoSecretKey(key []byte) []byte {
	var newKey []byte
	hexCodes := []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'}
	for _, l := range string(key) {
		asciiCode := int(l)
		if (asciiCode >= 'a' && asciiCode <= 'z') || (asciiCode >= 'A' && asciiCode <= 'Z') {
			newKey = append(newKey, hexCodes[(asciiCode+9)%16])
		} else {
			newKey = append(newKey, hexCodes[asciiCode%16])
		}
	}

	if len(newKey)%2 != 0 {
		newKey = append(newKey, hexCodes[0])
	}

	return newKey
}

func NewSafeManager() (Manager, error) {
	return newBaseManager(nil)
}
