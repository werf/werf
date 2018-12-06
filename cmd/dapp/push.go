package main

import (
	"fmt"

	"github.com/flant/dapp/cmd/dapp/docker_authorizer"
	"github.com/flant/dapp/pkg/build"
	"github.com/flant/dapp/pkg/dapp"
	"github.com/flant/dapp/pkg/docker"
	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/dapp/pkg/ssh_agent"
	"github.com/flant/dapp/pkg/true_git"
	"github.com/spf13/cobra"
)

var pushCmdData struct {
	Repo       string
	WithStages bool

	PushUsername string
	PushPassword string

	Tag        []string
	TagBranch  bool
	TagBuildID bool
	TagCI      bool
	TagCommit  bool
}

func newPushCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "push",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := runPush()
			if err != nil {
				return fmt.Errorf("push failed: %s", err)
			}
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
	cmd.PersistentFlags().BoolVarP(&pushCmdData.TagBuildID, "tag-build-id", "", false, "Tag by CI build id")
	cmd.PersistentFlags().BoolVarP(&pushCmdData.TagCI, "tag-ci", "", false, "Tag by CI branch and tag")
	cmd.PersistentFlags().BoolVarP(&pushCmdData.TagCommit, "tag-commit", "", false, "Tag by git commit")

	return cmd
}

func runPush() error {
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

	projectBuildDir, err := getProjectBuildDir(projectName)
	if err != nil {
		return fmt.Errorf("getting project build dir failed: %s", err)
	}

	projectTmpDir, err := getProjectTmpDir()
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}

	dappfile, err := parseDappfile(projectDir)
	if err != nil {
		return fmt.Errorf("dappfile parsing failed: %s", err)
	}

	repo, err := getRequiredRepoName(projectName, pushCmdData.Repo)
	if err != nil {
		return err
	}

	dockerAuthorizer, err := docker_authorizer.GetPushDockerAuthorizer(projectTmpDir, pushCmdData.PushUsername, pushCmdData.PushPassword, repo)
	if err != nil {
		return err
	}

	if err := docker.Init(dockerAuthorizer.HostDockerConfigDir); err != nil {
		return err
	}

	if err := ssh_agent.Init(rootCmdData.SSHKeys); err != nil {
		return fmt.Errorf("cannot initialize ssh-agent: %s", err)
	}

	opts, err := getPushOptions(projectDir, pushCmdData.Tag, pushCmdData.TagBranch, pushCmdData.TagCommit, pushCmdData.TagBuildID, pushCmdData.TagCI, pushCmdData.WithStages)
	if err != nil {
		return err
	}

	c := build.NewConveyor(dappfile, projectDir, projectName, projectBuildDir, projectTmpDir, ssh_agent.SSHAuthSock, dockerAuthorizer)
	if err = c.Push(repo, opts); err != nil {
		return err
	}

	return nil
}
