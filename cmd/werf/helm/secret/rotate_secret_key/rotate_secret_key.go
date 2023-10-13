package secret

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"
	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/cmd/werf/docs/replacers/helm"
	"github.com/werf/werf/pkg/deploy/secrets_manager"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/git_repo/gitdata"
	"github.com/werf/werf/pkg/secret"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "rotate-secret-key [EXTRA_SECRET_VALUES_FILE_PATH...]",
		DisableFlagsInUseLine: true,
		Short:                 "Regenerate secret files with new secret key",
		Long:                  common.GetLongCommandDescription(helm.GetHelmSecretRotateSecretKeyDocs().Long),
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfSecretKey, common.WerfOldSecretKey),
			common.DocsLongMD: helm.GetHelmSecretRotateSecretKeyDocs().LongMD,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			return runRotateSecretKey(ctx, cmd, args...)
		},
	})

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupGiterminismConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)

	return cmd
}

func runRotateSecretKey(ctx context.Context, cmd *cobra.Command, secretValuesPaths ...string) error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %w", err)
	}

	gitDataManager, err := gitdata.GetHostGitDataManager(ctx)
	if err != nil {
		return fmt.Errorf("error getting host git data manager: %w", err)
	}

	if err := git_repo.Init(gitDataManager); err != nil {
		return err
	}

	if err := true_git.Init(ctx, true_git.Options{LiveGitOutput: *commonCmdData.LogDebug}); err != nil {
		return err
	}

	giterminismManager, err := common.GetGiterminismManager(ctx, &commonCmdData)
	if err != nil {
		return err
	}

	werfConfigPath, werfConfig, err := common.GetRequiredWerfConfig(context.Background(), &commonCmdData, giterminismManager, common.GetWerfConfigOptions(&commonCmdData, true))
	if err != nil {
		return fmt.Errorf("unable to load werf config: %w", err)
	}

	helmChartDir, err := common.GetHelmChartDir(werfConfigPath, werfConfig, giterminismManager)
	if err != nil {
		return fmt.Errorf("getting helm chart dir failed: %w", err)
	}

	secretsManager := secrets_manager.NewSecretsManager(secrets_manager.SecretsManagerOptions{})

	newEncoder, err := secretsManager.GetYamlEncoder(ctx, giterminismManager.ProjectDir())
	if err != nil {
		common.PrintHelp(cmd)
		return err
	}

	oldEncoder, err := secretsManager.GetYamlEncoderForOldKey(ctx)
	if err != nil {
		common.PrintHelp(cmd)
		return err
	}

	return secretsRegenerate(newEncoder, oldEncoder, helmChartDir, secretValuesPaths...)
}

func secretsRegenerate(newEncoder, oldEncoder *secret.YamlEncoder, helmChartDir string, secretValuesPaths ...string) error {
	var secretFilesPaths []string
	var secretFilesData map[string][]byte
	var secretValuesFilesData map[string][]byte
	regeneratedFilesData := map[string][]byte{}

	isHelmChartDirExist, err := util.FileExists(helmChartDir)
	if err != nil {
		return err
	}

	if isHelmChartDirExist {
		defaultSecretValuesPath := filepath.Join(helmChartDir, "secret-values.yaml")
		isDefaultSecretValuesExist, err := util.FileExists(defaultSecretValuesPath)
		if err != nil {
			return err
		}

		if isDefaultSecretValuesExist {
			secretValuesPaths = append(secretValuesPaths, defaultSecretValuesPath)
		}

		secretDirectory := filepath.Join(helmChartDir, "secret")
		isSecretDirectoryExist, err := util.FileExists(secretDirectory)
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

	if err := regenerateSecrets(secretFilesData, regeneratedFilesData, oldEncoder.Decrypt, newEncoder.Encrypt); err != nil {
		return err
	}

	if err := regenerateSecrets(secretValuesFilesData, regeneratedFilesData, oldEncoder.DecryptYamlData, newEncoder.EncryptYamlData); err != nil {
		return err
	}

	for filePath, fileData := range regeneratedFilesData {
		err := logboek.LogProcess(fmt.Sprintf("Saving file %q", filePath)).DoError(func() error {
			fileData = append(bytes.TrimSpace(fileData), []byte("\n")...)
			return ioutil.WriteFile(filePath, fileData, 0o644)
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func regenerateSecrets(filesData, regeneratedFilesData map[string][]byte, decodeFunc, encodeFunc func([]byte) ([]byte, error)) error {
	for filePath, fileData := range filesData {
		err := logboek.LogProcess(fmt.Sprintf("Regenerating file %q", filePath)).
			DoError(func() error {
				data, err := decodeFunc(fileData)
				if err != nil {
					return fmt.Errorf("check old encryption key and file data: %w", err)
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
