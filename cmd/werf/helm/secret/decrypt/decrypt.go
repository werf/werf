package secret

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/cmd/werf/docs/replacers/helm"
	secret_common "github.com/werf/werf/cmd/werf/helm/secret/common"
	"github.com/werf/werf/pkg/deploy/secrets_manager"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/git_repo/gitdata"
	"github.com/werf/werf/pkg/secret"
	"github.com/werf/werf/pkg/werf"
)

var cmdData struct {
	OutputFilePath string
}

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "decrypt",
		DisableFlagsInUseLine: true,
		Short:                 "Decrypt data",
		Long:                  common.GetLongCommandDescription(helm.GetHelmSecretDecryptDocs().Long),
		Example: `  # Decrypt data in interactive mode
  $ werf helm secret decrypt
  Enter secret:
  test

  # Decrypt from a pipe
  $ cat .helm/secret/date | werf helm secret decrypt
  Tue Jun 26 09:58:10 PDT 1990`,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfSecretKey),
			common.DocsLongMD: helm.GetHelmSecretDecryptDocs().LongMD,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			return runSecretDecrypt(ctx)
		},
	})

	common.SetupDir(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)

	cmd.Flags().StringVarP(&cmdData.OutputFilePath, "output-file-path", "o", "", "Write to file instead of stdout")

	return cmd
}

func runSecretDecrypt(ctx context.Context) error {
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

	workingDir := common.GetWorkingDir(&commonCmdData)

	return secretDecrypt(ctx, secrets_manager.NewSecretsManager(secrets_manager.SecretsManagerOptions{}), workingDir)
}

func secretDecrypt(ctx context.Context, m *secrets_manager.SecretsManager, workingDir string) error {
	var encodedData []byte
	var data []byte
	var err error

	var encoder *secret.YamlEncoder
	if enc, err := m.GetYamlEncoder(ctx, workingDir); err != nil {
		return err
	} else {
		encoder = enc
	}

	if terminal.IsTerminal(int(os.Stdin.Fd())) {
		encodedData, err = secret_common.InputFromInteractiveStdin("Enter secret: ")
		if err != nil {
			return err
		}
	} else {
		encodedData, err = secret_common.InputFromStdin()
		if err != nil {
			return err
		}
	}

	if len(encodedData) == 0 {
		return nil
	}

	encodedData = bytes.TrimSpace(encodedData)
	data, err = encoder.Decrypt(encodedData)
	if err != nil {
		return err
	}

	if cmdData.OutputFilePath != "" {
		if err := secret_common.SaveGeneratedData(cmdData.OutputFilePath, data); err != nil {
			return err
		}
	} else {
		if terminal.IsTerminal(int(os.Stdout.Fd())) {
			if !bytes.HasSuffix(data, []byte("\n")) {
				data = append(data, []byte("\n")...)
			}
		}

		fmt.Printf("%s", string(data))
	}

	return nil
}
