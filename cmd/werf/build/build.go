package build

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/cmd/werf/docker_authorizer"
	"github.com/flant/werf/pkg/build"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/logger"
	"github.com/flant/werf/pkg/project_tmp_dir"
	"github.com/flant/werf/pkg/ssh_agent"
	"github.com/flant/werf/pkg/true_git"
	"github.com/flant/werf/pkg/werf"
)

var CmdData struct {
	PullUsername string
	PullPassword string

	IntrospectBeforeError bool
	IntrospectAfterError  bool
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

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)
	common.SetupSSHKey(&CommonCmdData, cmd)

	cmd.PersistentFlags().StringVarP(&CmdData.PullUsername, "pull-username", "", "", "Docker registry username to authorize pull of base images")
	cmd.PersistentFlags().StringVarP(&CmdData.PullPassword, "pull-password", "", "", "Docker registry password to authorize pull of base images")
	cmd.PersistentFlags().StringVarP(&CmdData.PullUsername, "registry-username", "", "", "Docker registry username to authorize pull of base images")
	cmd.PersistentFlags().StringVarP(&CmdData.PullPassword, "registry-password", "", "", "Docker registry password to authorize pull of base images")

	cmd.PersistentFlags().BoolVarP(&CmdData.IntrospectAfterError, "introspect-error", "", false, "Introspect failed stage in the state, right after running failed assembly instruction")
	cmd.PersistentFlags().BoolVarP(&CmdData.IntrospectBeforeError, "introspect-before-error", "", false, "Introspect failed stage in the clean state, before running all assembly instructions of the stage")

	return cmd
}

func runBuild(dimgsToProcess []string) error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
		return err
	}

	if err := true_git.Init(); err != nil {
		return err
	}

	if err := docker.Init(docker_authorizer.GetHomeDockerConfigDir()); err != nil {
		return err
	}

	projectDir, err := common.GetProjectDir(&CommonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	werfConfig, err := common.GetWerfConfig(projectDir)
	if err != nil {
		return fmt.Errorf("cannot parse werf config: %s", err)
	}

	projectName := werfConfig.Meta.Project

	projectBuildDir, err := common.GetProjectBuildDir(projectName)
	if err != nil {
		return fmt.Errorf("getting project build dir failed: %s", err)
	}

	projectTmpDir, err := project_tmp_dir.Get()
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}
	defer project_tmp_dir.Release(projectTmpDir)

	dockerAuthorizer, err := docker_authorizer.GetBuildDockerAuthorizer(projectTmpDir, CmdData.PullUsername, CmdData.PullPassword)
	if err != nil {
		return err
	}

	if err := ssh_agent.Init(*CommonCmdData.SSHKeys); err != nil {
		return fmt.Errorf("cannot initialize ssh agent: %s", err)
	}
	defer func() {
		err := ssh_agent.Terminate()
		if err != nil {
			logger.LogWarningF("WARNING: ssh agent termination failed: %s\n", err)
		}
	}()

	buildOpts := build.BuildOptions{
		ImageBuildOptions: image.BuildOptions{
			IntrospectAfterError:  CmdData.IntrospectAfterError,
			IntrospectBeforeError: CmdData.IntrospectBeforeError,
		},
	}

	c := build.NewConveyor(werfConfig, dimgsToProcess, projectDir, projectBuildDir, projectTmpDir, ssh_agent.SSHAuthSock, dockerAuthorizer)
	if err = c.Build(buildOpts); err != nil {
		return err
	}

	return nil
}
