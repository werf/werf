package secret

import (
	"bytes"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flant/dapp/cmd/dapp/common"
	"github.com/flant/dapp/pkg/deploy/secret"
)

var GenerateCmdData struct {
	FilePath       string
	OutputFilePath string
	Values         bool
}

func NewGenerateCmd() *cobra.Command {
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

	cmd.PersistentFlags().StringVarP(&GenerateCmdData.FilePath, "file-path", "", "", "Encode file data by specified path")
	cmd.PersistentFlags().StringVarP(&GenerateCmdData.OutputFilePath, "output-file-path", "", "", "Save encoded data by specified file path")
	cmd.PersistentFlags().BoolVarP(&GenerateCmdData.Values, "values", "", false, "Encode specified FILE_PATH (--file-path) as secret values file")

	return cmd
}

func runSecretGenerate() error {
	projectDir, err := common.GetProjectDir(&CommonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	options := &secretGenerateOptions{
		FilePath:       GenerateCmdData.FilePath,
		OutputFilePath: GenerateCmdData.OutputFilePath,
		Values:         GenerateCmdData.Values,
	}

	m, err := secret.GetManager(projectDir)
	if err != nil {
		return err
	}

	return secretGenerate(m, options)
}

func secretGenerate(m secret.Manager, options *secretGenerateOptions) error {
	var data []byte
	var encodedData []byte
	var err error

	if options.FilePath != "" {
		data, err = readFileData(options.FilePath)
		if err != nil {
			return err
		}
	} else {
		data, err = readStdin()
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
		if err := saveGeneratedData(options.OutputFilePath, encodedData); err != nil {
			return err
		}
	} else {
		fmt.Printf(string(encodedData))
	}

	return nil
}
