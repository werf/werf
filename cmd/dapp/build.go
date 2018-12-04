package main

import (
	"fmt"

	"github.com/flant/dapp/pkg/build"
	"github.com/flant/dapp/pkg/dapp"
	"github.com/flant/dapp/pkg/docker"
	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/dapp/pkg/ssh_agent"
	"github.com/flant/dapp/pkg/true_git"
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
			return runBuild()
		},
	}

	cmd.PersistentFlags().StringVarP(&buildCmdData.PullUsername, "pull-username", "", "", "Docker registry username to authorize pull of base images")
	cmd.PersistentFlags().StringVarP(&buildCmdData.PullPassword, "pull-password", "", "", "Docker registry password to authorize pull of base images")
	cmd.PersistentFlags().StringVarP(&buildCmdData.PullUsername, "registry-username", "", "", "Docker registry username to authorize pull of base images")
	cmd.PersistentFlags().StringVarP(&buildCmdData.PullPassword, "registry-password", "", "", "Docker registry password to authorize pull of base images")

	return cmd
}

func runBuild() error {
	if err := dapp.Init(rootCmdData.TmpDir, rootCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
		return err
	}

	if err := ssh_agent.Init(rootCmdData.SSHKeys); err != nil {
		return fmt.Errorf("cannot initialize ssh-agent: %s", err)
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

	hostDockerConfigDir, err := hostDockerConfigDir(projectTmpDir, buildCmdData.PullUsername, buildCmdData.PullPassword, "")
	if err != nil {
		return fmt.Errorf("getting host docker config dir failed: %s", err)
	}

	if err := docker.Init(hostDockerConfigDir); err != nil {
		return err
	}

	dappfile, err := parseDappfile(projectDir)
	if err != nil {
		return fmt.Errorf("parsing dappfile failed: %s", err)
	}

	authorizer, err := getDockerAuthorizer(buildCmdData.PullUsername, buildCmdData.PullPassword, "")
	if err != nil {
		return err
	}

	c := build.NewConveyor(dappfile, projectDir, projectName, projectBuildDir, projectTmpDir, ssh_agent.SSHAuthSock, authorizer)
	if err = c.Build(); err != nil {
		return err
	}

	return nil
}
