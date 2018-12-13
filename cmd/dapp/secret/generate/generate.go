package secret

import (
	"bytes"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flant/dapp/cmd/dapp/common"
	secret_common "github.com/flant/dapp/cmd/dapp/secret/common"
	"github.com/flant/dapp/pkg/dapp"
	"github.com/flant/dapp/pkg/deploy/secret"
)

var CmdData struct {
	FilePath       string
	OutputFilePath string
	Values         bool
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate secret data",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := runSecretGenerate()
			if err != nil {
				return fmt.Errorf("secret generate failed: %s", err)
			}
			return nil
		},
	}

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)

	cmd.PersistentFlags().StringVarP(&CmdData.FilePath, "file-path", "", "", "Encode file data by specified path")
	cmd.PersistentFlags().StringVarP(&CmdData.OutputFilePath, "output-file-path", "", "", "Save encoded data by specified file path")
	cmd.PersistentFlags().BoolVarP(&CmdData.Values, "values", "", false, "Encode specified FILE_PATH (--file-path) as secret values file")

	return cmd
}

func runSecretGenerate() error {
	if err := dapp.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	projectDir, err := common.GetProjectDir(&CommonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	options := &secret_common.GenerateOptions{
		FilePath:       CmdData.FilePath,
		OutputFilePath: CmdData.OutputFilePath,
		Values:         CmdData.Values,
	}

	m, err := secret.GetManager(projectDir)
	if err != nil {
		return err
	}

	return secretGenerate(m, options)
}

func secretGenerate(m secret.Manager, options *secret_common.GenerateOptions) error {
	var data []byte
	var encodedData []byte
	var err error

	if options.FilePath != "" {
		data, err = secret_common.ReadFileData(options.FilePath)
		if err != nil {
			return err
		}
	} else {
		data, err = secret_common.ReadStdin()
		if err != nil {
			return err
		}

		if len(data) == 0 {
			return nil
		}
	}

	if options.FilePath != "" && options.Values {
		encodedData, err = m.GenerateYamlData(data)
		if err != nil {
			return err
		}
	} else {
		encodedData, err = m.Generate(data)
		if err != nil {
			return err
		}
	}

	if !bytes.HasSuffix(encodedData, []byte("\n")) {
		encodedData = append(encodedData, []byte("\n")...)
	}

	if options.OutputFilePath != "" {
		if err := secret_common.SaveGeneratedData(options.OutputFilePath, encodedData); err != nil {
			return err
		}
	} else {
		fmt.Printf(string(encodedData))
	}

	return nil
}
