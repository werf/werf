package main

import (
	"github.com/spf13/cobra"
)

var pushCmdData struct {
	Repo       string
	WithStages bool

	PushUsername string
	PushPassword string

	Tag        []string
	TagBranch  bool
	TagBuildId bool
	TagCi      bool
	TagCommit  bool
}

func newPushCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "push",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&pushCmdData.Repo, "repo", "", "", "Docker repository name to push images to. CI_REGISTRY_IMAGE will be used by default if available.")
	cmd.PersistentFlags().BoolVarP(&pushCmdData.WithStages, "with-stages", "", false, "Push images with stages cache")

	cmd.PersistentFlags().StringVarP(&pushCmdData.PushUsername, "push-username", "", "", "Docker registry username to authorize push to the docker repo")
	cmd.PersistentFlags().StringVarP(&pushCmdData.PushPassword, "push-password", "", "", "Docker registry password to authorize push to the docker repo")
	cmd.PersistentFlags().StringVarP(&pushCmdData.PushUsername, "registry-username", "", "", "Docker registry username to authorize push to the docker repo")
	cmd.PersistentFlags().StringVarP(&pushCmdData.PushPassword, "registry-password", "", "", "Docker registry password to authorize push to the docker repo")

	cmd.PersistentFlags().StringArrayVarP(&pushCmdData.Tag, "tag", "", []string{}, "Add tag (can be used one or more times)")
	cmd.PersistentFlags().BoolVarP(&pushCmdData.TagBranch, "tag-branch", "", false, "Tag by git branch")
	cmd.PersistentFlags().BoolVarP(&pushCmdData.TagBuildId, "tag-build-id", "", false, "Tag by CI build id")
	cmd.PersistentFlags().BoolVarP(&pushCmdData.TagCi, "tag-ci", "", false, "Tag by CI branch and tag")
	cmd.PersistentFlags().BoolVarP(&pushCmdData.TagCommit, "tag-commit", "", false, "Tag by git commit")

	return cmd
}
