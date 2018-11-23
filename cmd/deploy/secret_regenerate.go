package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/flant/dapp/pkg/deploy/secret"

	"k8s.io/kubernetes/pkg/util/file"
)

func SecretsRegenerate(newManager, oldManager secret.Manager, projectPath string, secretValuesPaths ...string) error {
	var secretFilesPaths []string
	regeneratedFilesData := map[string][]byte{}
	secretFilesData := map[string][]byte{}
	secretValuesFilesData := map[string][]byte{}

	helmChartPath := filepath.Join(projectPath, ".helm")
	exist, err := file.FileExists(helmChartPath)
	if err != nil {
		return err
	}

	if exist {
		defaultSecretValuesPath := filepath.Join(helmChartPath, "secret-values.yaml")
		exist, err := file.FileExists(defaultSecretValuesPath)
		if err != nil {
			return err
		}

		if exist {
			secretValuesPaths = append(secretValuesPaths, defaultSecretValuesPath)
		}

		secretDirectory := filepath.Join(helmChartPath, "secret")
		err = filepath.Walk(secretDirectory,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				fileInfo, err := os.Stat(path)
				if err != nil {
					return err
				}

				if !fileInfo.IsDir() {
					secretFilesPaths = append(secretFilesPaths, path)
				}

				return nil
			})
		if err != nil {
			return err
		}
	}

	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	secretFilesData, err = readFilesToDecode(secretFilesPaths, pwd)
	if err != nil {
		return err
	}

	secretValuesFilesData, err = readFilesToDecode(secretValuesPaths, pwd)
	if err != nil {
		return err
	}

	if err := regenerateSecrets(secretFilesData, regeneratedFilesData, oldManager.Extract, newManager.Generate); err != nil {
		return err
	}

	if err := regenerateSecrets(secretValuesFilesData, regeneratedFilesData, oldManager.ExtractYamlData, newManager.GenerateYamlData); err != nil {
		return err
	}

	for filePath, fileData := range regeneratedFilesData {
		fmt.Printf("save file '%s'\n", filePath)

		fileData = append(bytes.TrimSpace(fileData), []byte("\n")...)
		if err := ioutil.WriteFile(filePath, fileData, 0644); err != nil {
			return err
		}
	}

	return nil
}

func regenerateSecrets(filesData, regeneratedFilesData map[string][]byte, decodeFunc, encodeFunc func([]byte) ([]byte, error)) error {
	for filePath, fileData := range filesData {
		fmt.Printf("regenerate file '%s' data\n", filePath)

		data, err := decodeFunc(fileData)
		if err != nil {
			return fmt.Errorf("check old encryption key and file data: %s", err)
		}

		resultData, err := encodeFunc(data)
		if err != nil {
			return err
		}

		regeneratedFilesData[filePath] = resultData
	}

	return nil
}

func readFilesToDecode(filePaths []string, pwd string) (map[string][]byte, error) {
	filesData := map[string][]byte{}
	for _, filePath := range filePaths {
		fileData, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		if filepath.IsAbs(filePath) {
			filePath, err = filepath.Rel(pwd, filePath)
			if err != nil {
				return nil, err
			}
		}

		filesData[filePath] = bytes.TrimSpace(fileData)
	}

	return filesData, nil
}
