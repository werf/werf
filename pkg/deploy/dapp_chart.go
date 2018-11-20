package deploy

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/flant/dapp/pkg/dapp"
	"github.com/flant/dapp/pkg/secret"
	"github.com/otiai10/copy"
	uuid "github.com/satori/go.uuid"
)

const (
	DefaultSecretValuesFile = "secret-values.yaml"
	SecretDirName           = "secret"
	DecodedSecretDirName    = "decoded-secret"
)

type DappChart struct {
	ChartDir string
	Values   []string
	Set      []string
}

func (chart *DappChart) Deploy() error {
	return nil
}

func (chart *DappChart) Render() error {
	return nil
}

func (chart *DappChart) Lint() error {
	return nil
}

func GenerateDappChart(projectDir string, opts DappChartOptions) (*DappChart, error) {
	tmpChartPath := filepath.Join(dapp.GetTmpDir(), fmt.Sprintf("dapp-chart-%s", uuid.NewV4().String()))
	return PrepareDappChart(projectDir, tmpChartPath, opts)
}

type DappChartOptions struct {
	Secret       secret.Secret
	SecretValues []string
}

func PrepareDappChart(projectDir string, targetDir string, opts DappChartOptions) (*DappChart, error) {
	dappChart := &DappChart{ChartDir: targetDir}

	projectHelmDir := filepath.Join(projectDir, ".helm")
	err := copy.Copy(projectHelmDir, targetDir)
	if err != nil {
		return nil, fmt.Errorf("unable to copy project helm dir %s into %s: %s", projectHelmDir, targetDir, err)
	}

	helpersTplPath := filepath.Join(targetDir, "templates/_dapp_helpers.tpl")
	f, err := os.Create(helpersTplPath)
	if err != nil {
		return nil, fmt.Errorf("unable to create %s: %s", helpersTplPath, err)
	}
	_, err = f.Write(DappChartHelpersTpl)
	if err != nil {
		return nil, fmt.Errorf("unable to write %s: %s", helpersTplPath, err)
	}

	allSecretValues := []string{}
	defaultSecretValues := filepath.Join(projectDir, DefaultSecretValuesFile)
	if _, err := os.Stat(defaultSecretValues); !os.IsNotExist(err) {
		allSecretValues = append(allSecretValues, defaultSecretValues)
	}
	for _, path := range opts.SecretValues {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil, fmt.Errorf("secret values yaml file %s not found", path)
		}
		allSecretValues = append(allSecretValues, path)
	}

	for fileNo, path := range allSecretValues {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("cannot read secret values file %s: %s", path, err)
		}

		if opts.Secret == nil {
			return nil, fmt.Errorf("cannot decode secret values: no Secret option specified")
		}

		decodedData, err := decodeSecretValues(data, opts.Secret)
		if err != nil {
			return nil, fmt.Errorf("cannot decode secret values file %s data: %s", path, err)
		}

		newPath := filepath.Join(targetDir, fmt.Sprintf("decoded-secret-values-%d.yaml", fileNo))
		err = ioutil.WriteFile(newPath, decodedData, 0400)
		if err != nil {
			return nil, fmt.Errorf("cannot write decoded secret values file %s: %s", newPath, err)
		}

		dappChart.Values = append(dappChart.Values, newPath)
	}

	secretDir := filepath.Join(projectDir, SecretDirName)
	if _, err := os.Stat(secretDir); !os.IsNotExist(err) {
		err := filepath.Walk(secretDir, func(path string, info os.FileInfo, accessErr error) error {
			if accessErr != nil {
				return fmt.Errorf("error accessing file %s: %s", path, accessErr)
			}

			if info.Mode().IsDir() {
				return nil
			}

			relativePath := strings.TrimPrefix(path, secretDir)
			newPath := filepath.Join(targetDir, DecodedSecretDirName, relativePath)

			if opts.Secret == nil {
				err := os.MkdirAll(filepath.Dir(newPath), os.ModePerm)
				if err != nil {
					return err
				}
				err = ioutil.WriteFile(newPath, []byte{}, 0400)
				if err != nil {
					return fmt.Errorf("unable to create decoded secret file %s: %s", newPath, err)
				}
				return nil
			}

			data, err := ioutil.ReadFile(path)
			if err != nil {
				return fmt.Errorf("error reading file %s: %s", path, err)
			}

			decodedData, err := decodeSecret([]byte(strings.TrimRightFunc(string(data), unicode.IsSpace)), opts.Secret)
			if err != nil {
				return fmt.Errorf("error decoding %s: %s", path, err)
			}

			err = os.MkdirAll(filepath.Dir(newPath), os.ModePerm)
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(newPath, decodedData, 0400)
			if err != nil {
				return fmt.Errorf("error writing file %s: %s", newPath, err)
			}

			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	return dappChart, nil
}

func decodeSecretValues(data []byte, secret secret.Secret) ([]byte, error) {
	return data, nil
}

func decodeSecret(data []byte, secret secret.Secret) ([]byte, error) {
	return data, nil
}
