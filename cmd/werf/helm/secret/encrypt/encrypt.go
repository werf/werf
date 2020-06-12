package secret

import (
	"bytes"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/werf/werf/cmd/werf/common"
	secret_common "github.com/werf/werf/cmd/werf/helm/secret/common"
	"github.com/werf/werf/pkg/deploy/secret"
	"github.com/werf/werf/pkg/werf"
)

var cmdData struct {
	OutputFilePath string
}

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "encrypt",
		DisableFlagsInUseLine: true,
		Short:                 "Encrypt data",
		Long: common.GetLongCommandDescription(`Encrypt data from standard input.
Encryption key should be in $WERF_SECRET_KEY or .werf_secret_key file`),
		Example: `  # Encrypt data in interactive mode
  $ werf helm secret encrypt
  Enter secret:
  100044d3f6a2ffd6dd2b73fa8f50db5d61fb6ac04da29955c77d13bb44e937448ee4

  # Encrypt from a pipe and save result in file
  $ date | werf helm secret encrypt -o .helm/secret/date`,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfSecretKey),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			return runSecretEncrypt()
		},
	}

	common.SetupDir(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)

	cmd.Flags().StringVarP(&cmdData.OutputFilePath, "output-file-path", "o", "", "Write to file instead of stdout")

	return cmd
}

func runSecretEncrypt() error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	projectDir, err := common.GetProjectDir(&commonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	m, err := secret.GetManager(projectDir)
	if err != nil {
		return err
	}

	return secretEncrypt(m)
}

func secretEncrypt(m secret.Manager) error {
	var data []byte
	var encodedData []byte
	var err error

	if terminal.IsTerminal(int(os.Stdin.Fd())) {
		data, err = secret_common.InputFromInteractiveStdin()
		if err != nil {
			return err
		}
	} else {
		data, err = secret_common.InputFromStdin()
		if err != nil {
			return err
		}
	}

	if len(data) == 0 {
		return nil
	}

	encodedData, err = m.Encrypt(data)
	if err != nil {
		return err
	}

	if !bytes.HasSuffix(encodedData, []byte("\n")) {
		encodedData = append(encodedData, []byte("\n")...)
	}

	if cmdData.OutputFilePath != "" {
		if err := secret_common.SaveGeneratedData(cmdData.OutputFilePath, encodedData); err != nil {
			return err
		}
	} else {
		fmt.Printf("%s", string(encodedData))
	}

	return nil
}
