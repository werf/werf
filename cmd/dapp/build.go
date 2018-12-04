package main

import (
	"github.com/spf13/cobra"
)

var buildCmdData struct {
	PullUsername string
	PullPassword string
}

func newBuildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "build",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&buildCmdData.PullUsername, "pull-username", "", "", "Docker registry username to authorize pull of base images")
	cmd.PersistentFlags().StringVarP(&buildCmdData.PullPassword, "pull-password", "", "", "Docker registry password to authorize pull of base images")
	cmd.PersistentFlags().StringVarP(&buildCmdData.PullUsername, "registry-username", "", "", "Docker registry username to authorize pull of base images")
	cmd.PersistentFlags().StringVarP(&buildCmdData.PullPassword, "registry-password", "", "", "Docker registry password to authorize pull of base images")

	return cmd
}
