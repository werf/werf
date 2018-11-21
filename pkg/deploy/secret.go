package deploy

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

func GetSecret(projectDir string) (secret.Secret, error) {
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

	s, err := secret.NewSecret(secretKey)
	if err != nil {
		return nil, fmt.Errorf("check encryption key: %s", err)
	}

	return s, nil
}
