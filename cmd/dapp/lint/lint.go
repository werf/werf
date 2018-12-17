package lint

import (
	"fmt"

	"github.com/flant/dapp/cmd/dapp/common"
	"github.com/flant/dapp/pkg/dapp"
	"github.com/flant/dapp/pkg/deploy"
	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/dapp/pkg/true_git"
	"github.com/spf13/cobra"
)

var CmdData struct {
	Values       []string
	SecretValues []string
	Set          []string
	SetString    []string
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "lint",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := runLint()
			if err != nil {
				return fmt.Errorf("lint failed: %s", err)
			}
			return nil
		},
	}

	common.SetupName(&CommonCmdData, cmd)
	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)

	cmd.PersistentFlags().StringArrayVarP(&CmdData.Values, "values", "", []string{}, "Additional helm values")
	cmd.PersistentFlags().StringArrayVarP(&CmdData.SecretValues, "secret-values", "", []string{}, "Additional helm secret values")
	cmd.PersistentFlags().StringArrayVarP(&CmdData.Set, "set", "", []string{}, "Additional helm sets")
	cmd.PersistentFlags().StringArrayVarP(&CmdData.SetString, "set-string", "", []string{}, "Additional helm STRING sets")

	return cmd
}

func runLint() error {
	if err := dapp.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
		return err
	}

	if err := true_git.Init(); err != nil {
		return err
	}

	projectDir, err := common.GetProjectDir(&CommonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	projectName, err := common.GetProjectName(&CommonCmdData, projectDir)
	if err != nil {
		return fmt.Errorf("getting project name failed: %s", err)
	}

	dappfile, err := common.GetDappfile(projectDir)
	if err != nil {
		return fmt.Errorf("dappfile parsing failed: %s", err)
	}

	return deploy.RunLint(projectName, projectDir, dappfile, deploy.LintOptions{
		Values:       CmdData.Values,
		SecretValues: CmdData.SecretValues,
		Set:          CmdData.Set,
		SetString:    CmdData.SetString,
	})
}
