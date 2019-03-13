package secret

import (
	"bytes"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/flant/werf/cmd/werf/common"
	secret_common "github.com/flant/werf/cmd/werf/helm/secret/common"
	"github.com/flant/werf/pkg/deploy/secret"
	"github.com/flant/werf/pkg/werf"
)

var CmdData struct {
	OutputFilePath string
}

var CommonCmdData common.CmdData

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
			return runSecretDecrypt()
		},
	}

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)

	cmd.Flags().StringVarP(&CmdData.OutputFilePath, "output-file-path", "o", "", "Write to file instead of stdout")

	return cmd
}

func runSecretDecrypt() error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	projectDir, err := common.GetProjectDir(&CommonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	m, err := secret.GetManager(projectDir)
	if err != nil {
		return err
	}

	return secretDecrypt(m)
}

func secretDecrypt(m secret.Manager) error {
	var encodedData []byte
	var data []byte
	var err error

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
	data, err = m.Decrypt(encodedData)
	if err != nil {
		return err
	}

	if CmdData.OutputFilePath != "" {
		if err := secret_common.SaveGeneratedData(CmdData.OutputFilePath, data); err != nil {
			return err
		}
	} else {
		if terminal.IsTerminal(int(os.Stdout.Fd())) {
			if !bytes.HasSuffix(data, []byte("\n")) {
				data = append(data, []byte("\n")...)
			}
		}

		fmt.Printf(string(data))
	}

	return nil
}
