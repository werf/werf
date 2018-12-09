package main

import (
	"fmt"

	"github.com/flant/dapp/pkg/dapp"
	"github.com/flant/dapp/pkg/deploy"
	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/dapp/pkg/true_git"
	"github.com/spf13/cobra"
)

var renderCmdData struct {
	Values       []string
	SecretValues []string
	Set          []string
}

func newRenderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "render",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := runRender()
			if err != nil {
				return fmt.Errorf("render failed: %s", err)
			}
			return nil
		},
	}

	cmd.PersistentFlags().StringArrayVarP(&renderCmdData.Values, "values", "", []string{}, "Additional helm values")
	cmd.PersistentFlags().StringArrayVarP(&renderCmdData.SecretValues, "secret-values", "", []string{}, "Additional helm secret values")
	cmd.PersistentFlags().StringArrayVarP(&renderCmdData.Set, "set", "", []string{}, "Additional helm sets")

	return cmd
}

func runRender() error {
	if err := dapp.Init(rootCmdData.TmpDir, rootCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
		return err
	}

	if err := true_git.Init(); err != nil {
		return err
	}

	projectDir, err := getProjectDir()
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	projectName, err := getProjectName(projectDir)
	if err != nil {
		return fmt.Errorf("getting project name failed: %s", err)
	}

	dappfile, err := parseDappfile(projectDir)
	if err != nil {
		return fmt.Errorf("dappfile parsing failed: %s", err)
	}

	return deploy.RunRender(projectName, projectDir, dappfile, deploy.RenderOptions{
		Values:       renderCmdData.Values,
		SecretValues: renderCmdData.SecretValues,
		Set:          renderCmdData.Set,
	})
}
