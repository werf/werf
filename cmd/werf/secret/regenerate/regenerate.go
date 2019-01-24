package secret

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/util/file"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/deploy/secret"
	"github.com/flant/werf/pkg/logger"
	"github.com/flant/werf/pkg/werf"
)

var CmdData struct {
	OldKey string
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "regenerate [EXTRA_SECRET_VALUES_FILE_PATH...]",
		DisableFlagsInUseLine: true,
		Short:                 "Regenerate secret files with new secret key",
		Long: common.GetLongCommandDescription(`Regenerate secret files with new secret key.

Old key should be specified with the --old-key option.
New key should reside either in the WERF_SECRET_KEY environment variable or .werf_secret_key file.

Command will extract data with the old key, generate new secret data and rewrite files:
* standard raw secret files in the .helm/secret folder;
* standard secret values yaml file .helm/secret-values.yaml;
* additional secret values yaml files specified with EXTRA_SECRET_VALUES_FILE_PATH params.`),
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfSecretKey),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := runSecretRegenerate(args...)
			if err != nil {
				return fmt.Errorf("secret regenerate failed: %s", err)
			}
			return nil
		},
	}

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)

	cmd.Flags().StringVarP(&CmdData.OldKey, "old-key", "", "", "Old secret key")
	cmd.MarkPersistentFlagRequired("old-key")

	return cmd
}

func runSecretRegenerate(secretValuesPaths ...string) error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	projectDir, err := common.GetProjectDir(&CommonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	newSecret, err := secret.GetManager(projectDir)
	if err != nil {
		return err
	}

	oldSecret, err := secret.NewManager([]byte(CmdData.OldKey), secret.NewManagerOptions{IgnoreWarning: true})
	if err != nil {
		return err
	}

	return secretsRegenerate(newSecret, oldSecret, projectDir, secretValuesPaths...)
}

func secretsRegenerate(newManager, oldManager secret.Manager, projectPath string, secretValuesPaths ...string) error {
	var secretFilesPaths []string
	regeneratedFilesData := map[string][]byte{}
	secretFilesData := map[string][]byte{}
	secretValuesFilesData := map[string][]byte{}

	helmChartPath := filepath.Join(projectPath, ".helm")
	isHelmChartDirExist, err := file.FileExists(helmChartPath)
	if err != nil {
		return err
	}

	if isHelmChartDirExist {
		defaultSecretValuesPath := filepath.Join(helmChartPath, "secret-values.yaml")
		isDefaultSecretValuesExist, err := file.FileExists(defaultSecretValuesPath)
		if err != nil {
			return err
		}

		if isDefaultSecretValuesExist {
			secretValuesPaths = append(secretValuesPaths, defaultSecretValuesPath)
		}

		secretDirectory := filepath.Join(helmChartPath, "secret")
		isSecretDirectoryExist, err := file.FileExists(defaultSecretValuesPath)
		if err != nil {
			return err
		}

		if isSecretDirectoryExist {
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
		err := logger.LogProcess(fmt.Sprintf("File '%s'", filePath), "[SAVING]", func() error {
			fileData = append(bytes.TrimSpace(fileData), []byte("\n")...)
			return ioutil.WriteFile(filePath, fileData, 0644)
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func regenerateSecrets(filesData, regeneratedFilesData map[string][]byte, decodeFunc, encodeFunc func([]byte) ([]byte, error)) error {
	for filePath, fileData := range filesData {
		err := logger.LogProcess(fmt.Sprintf("File '%s'", filePath), "[REGENERATING]", func() error {
			data, err := decodeFunc(fileData)
			if err != nil {
				return fmt.Errorf("check old encryption key and file data: %s", err)
			}

			resultData, err := encodeFunc(data)
			if err != nil {
				return err
			}

			regeneratedFilesData[filePath] = resultData

			return nil
		})

		if err != nil {
			return err
		}
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
