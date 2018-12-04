package main

import (
	"github.com/spf13/cobra"
)

var bpCmdData struct {
	PullUsername     string
	PullPassword     string
	PushUsername     string
	PushPassword     string
	RegistryUsername string
	RegistryPassword string

	Tag        []string
	TagBranch  bool
	TagBuildId bool
	TagCi      bool
	TagCommit  bool
}

func newBPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "bp",
		RunE: func(cmd *cobra.Command, args []string) error {
			if bpCmdData.PullUsername == "" {
				bpCmdData.PullUsername = bpCmdData.RegistryUsername
			}
			if bpCmdData.PullPassword == "" {
				bpCmdData.PullPassword = bpCmdData.RegistryPassword
			}
			if bpCmdData.PushUsername == "" {
				bpCmdData.PushUsername = bpCmdData.RegistryUsername
			}
			if bpCmdData.PushPassword == "" {
				bpCmdData.PushPassword = bpCmdData.RegistryPassword
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&bpCmdData.PullUsername, "pull-username", "", "", "Docker registry username to authorize pull of base images")
	cmd.PersistentFlags().StringVarP(&bpCmdData.PullPassword, "pull-password", "", "", "Docker registry password to authorize pull of base images")
	cmd.PersistentFlags().StringVarP(&bpCmdData.PushUsername, "push-username", "", "", "Docker registry username to authorize push to the docker repo")
	cmd.PersistentFlags().StringVarP(&bpCmdData.PushPassword, "push-password", "", "", "Docker registry password to authorize push to the docker repo")
	cmd.PersistentFlags().StringVarP(&bpCmdData.RegistryUsername, "registry-username", "", "", "Docker registry username to authorize pull of base images and push to the docker repo")
	cmd.PersistentFlags().StringVarP(&bpCmdData.RegistryUsername, "registry-password", "", "", "Docker registry password to authorize pull of base images and push to the docker repo")

	cmd.PersistentFlags().StringArrayVarP(&bpCmdData.Tag, "tag", "", []string{}, "Add tag (can be used one or more times)")
	cmd.PersistentFlags().BoolVarP(&bpCmdData.TagBranch, "tag-branch", "", false, "Tag by git branch")
	cmd.PersistentFlags().BoolVarP(&bpCmdData.TagBuildId, "tag-build-id", "", false, "Tag by CI build id")
	cmd.PersistentFlags().BoolVarP(&bpCmdData.TagCi, "tag-ci", "", false, "Tag by CI branch and tag")
	cmd.PersistentFlags().BoolVarP(&bpCmdData.TagCommit, "tag-commit", "", false, "Tag by git commit")

	return cmd
}
