package main

import (
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"os"

	"github.com/spf13/cobra"

	"github.com/flant/dapp/pkg/deploy/secret"
)

var secretExtractCmdData struct {
	FilePath       string
	OutputFilePath string
	Values         bool
}

func newSecretExtractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extract",
		Short: "Extract data",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSecretExtract()
		},
	}

	cmd.PersistentFlags().StringVarP(&secretExtractCmdData.FilePath, "file-path", "", "", "Decode file data by specified path")
	cmd.PersistentFlags().StringVarP(&secretExtractCmdData.OutputFilePath, "output-file-path", "", "", "Save decoded data by specified file path")
	cmd.PersistentFlags().BoolVarP(&secretExtractCmdData.Values, "values", "", false, "Decode specified FILE_PATH (--file-path) as secret values file")

	return cmd
}

func runSecretExtract() error {
	projectDir, err := getProjectDir()
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	options := &secretGenerateOptions{
		FilePath:       secretExtractCmdData.FilePath,
		OutputFilePath: secretExtractCmdData.OutputFilePath,
		Values:         secretExtractCmdData.Values,
	}

	m, err := secret.GetManager(projectDir)
	if err != nil {
		return err
	}

	return secretExtract(m, options)
}

func secretExtract(m secret.Manager, options *secretGenerateOptions) error {
	var encodedData []byte
	var data []byte
	var err error

	if options.FilePath != "" {
		encodedData, err = readFileData(options.FilePath)
		if err != nil {
			return err
		}
	} else {
		encodedData, err = readStdin()
		if err != nil {
			return err
		}

		if len(encodedData) == 0 {
			return nil
		}
	}

	encodedData = bytes.TrimSpace(encodedData)

	if options.FilePath != "" && options.Values {
		data, err = m.ExtractYamlData(encodedData)
		if err != nil {
			return err
		}
	} else {
		data, err = m.Extract(encodedData)
		if err != nil {
			return err
		}
	}

	if options.OutputFilePath != "" {
		if err := saveGeneratedData(options.OutputFilePath, data); err != nil {
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
