package secret

import (
	"bytes"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/common"
	secret_common "github.com/flant/werf/cmd/werf/helm/secret/common"
	"github.com/flant/werf/pkg/deploy/secret"
	"github.com/flant/werf/pkg/werf"
)

var CmdData struct {
	FilePath       string
	OutputFilePath string
	Values         bool
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "encrypt",
		DisableFlagsInUseLine: true,
		Short:                 "Encrypt data",
		Long: common.GetLongCommandDescription(`Encrypt provided data.

Provide data onto stdin by default.

Data can be provided in file by specifying --file-path option. Option --values should be specified in the case when values yaml file provided`),
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfSecretKey),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSecretEncrypt()
		},
	}

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)

	cmd.Flags().StringVarP(&CmdData.FilePath, "file-path", "", "", "Encode file data by specified path")
	cmd.Flags().StringVarP(&CmdData.OutputFilePath, "output-file-path", "", "", "Save encoded data by specified file path")
	cmd.Flags().BoolVarP(&CmdData.Values, "values", "", false, "Encode specified FILE_PATH (--file-path) as secret values file")

	return cmd
}

func runSecretEncrypt() error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
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

	return secretEncrypt(m, options)
}

func secretEncrypt(m secret.Manager, options *secret_common.GenerateOptions) error {
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
		encodedData, err = m.EncryptYamlData(data)
		if err != nil {
			return err
		}
	} else {
		encodedData, err = m.Encrypt(data)
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
