package secret

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/werf/werf/pkg/secret"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/werf/werf/cmd/werf/common"
	secret_common "github.com/werf/werf/cmd/werf/helm/secret/common"
	"github.com/werf/werf/pkg/deploy/secrets_manager"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/werf"
)

var cmdData struct {
	OutputFilePath string
}

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "decrypt",
		DisableFlagsInUseLine: true,
		Short:                 "Decrypt data",
		Long: common.GetLongCommandDescription(`Decrypt data from standard input.
Encryption key should be in $WERF_SECRET_KEY or .werf_secret_key file`),
		Example: `  # Decrypt data in interactive mode
  $ werf helm secret decrypt
  Enter secret:
  test

  # Decrypt from a pipe
  $ cat .helm/secret/date | werf helm secret decrypt
  Tue Jun 26 09:58:10 PDT 1990`,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfSecretKey),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			return runSecretDecrypt(common.BackgroundContext())
		},
	}

	common.SetupDir(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)

	common.SetupGiterminismInspectorOptions(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)

	cmd.Flags().StringVarP(&cmdData.OutputFilePath, "output-file-path", "o", "", "Write to file instead of stdout")

	return cmd
}

func runSecretDecrypt(ctx context.Context) error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := git_repo.Init(); err != nil {
		return err
	}

	workingDir := common.GetWorkingDir(&commonCmdData)

	return secretDecrypt(ctx, secrets_manager.NewSecretsManager(workingDir, secrets_manager.SecretsManagerOptions{}))
}

func secretDecrypt(ctx context.Context, m *secrets_manager.SecretsManager) error {
	var encodedData []byte
	var data []byte
	var err error

	var encoder *secret.YamlEncoder
	if enc, err := m.GetYamlEncoder(ctx); err != nil {
		return err
	} else {
		encoder = enc
	}

	if terminal.IsTerminal(int(os.Stdin.Fd())) {
		encodedData, err = secret_common.InputFromInteractiveStdin()
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
