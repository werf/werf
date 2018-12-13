package build

import (
	"fmt"
	"os"

	"github.com/flant/dapp/cmd/dapp/common"
	"github.com/flant/dapp/cmd/dapp/docker_authorizer"
	"github.com/flant/dapp/pkg/build"
	"github.com/flant/dapp/pkg/dapp"
	"github.com/flant/dapp/pkg/docker"
	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/dapp/pkg/ssh_agent"
	"github.com/flant/dapp/pkg/true_git"
	"github.com/spf13/cobra"
)

var CmdData struct {
	PullUsername string
	PullPassword string
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "build [DIMG_NAME...]",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := runBuild(args)
			if err != nil {
				return fmt.Errorf("build failed: %s", err)
			}
			return nil
		},
	}

	common.SetupName(&CommonCmdData, cmd)
	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)
	common.SetupSSHKey(&CommonCmdData, cmd)

	cmd.PersistentFlags().StringVarP(&CmdData.PullUsername, "pull-username", "", "", "Docker registry username to authorize pull of base images")
	cmd.PersistentFlags().StringVarP(&CmdData.PullPassword, "pull-password", "", "", "Docker registry password to authorize pull of base images")
	cmd.PersistentFlags().StringVarP(&CmdData.PullUsername, "registry-username", "", "", "Docker registry username to authorize pull of base images")
	cmd.PersistentFlags().StringVarP(&CmdData.PullPassword, "registry-password", "", "", "Docker registry password to authorize pull of base images")

	return cmd
}

func runBuild(dimgsToProcess []string) error {
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

	projectBuildDir, err := common.GetProjectBuildDir(projectName)
	if err != nil {
		return fmt.Errorf("getting project build dir failed: %s", err)
	}

	projectTmpDir, err := common.GetProjectTmpDir()
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}
	if !docker.Debug() {
		defer common.RemoveProjectTmpDir(projectTmpDir)
	}

	dappfile, err := common.GetDappfile(projectDir)
	if err != nil {
		return fmt.Errorf("dappfile parsing failed: %s", err)
	}

	dockerAuthorizer, err := docker_authorizer.GetBuildDockerAuthorizer(projectTmpDir, CmdData.PullUsername, CmdData.PullPassword)
	if err != nil {
		return err
	}

	if err := docker.Init(dockerAuthorizer.HostDockerConfigDir); err != nil {
		return err
	}

	if err := ssh_agent.Init(*CommonCmdData.SSHKeys); err != nil {
		return fmt.Errorf("cannot initialize ssh agent: %s", err)
	}
	defer func() {
		err := ssh_agent.Terminate()
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: ssh agent termination failed: %s", err)
		}
	}()

	c := build.NewConveyor(dappfile, dimgsToProcess, projectDir, projectName, projectBuildDir, projectTmpDir, ssh_agent.SSHAuthSock, dockerAuthorizer)
	if err = c.Build(); err != nil {
		return err
	}

	return nil
}
