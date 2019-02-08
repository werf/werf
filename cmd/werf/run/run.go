package run

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/build"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/logger"
	"github.com/flant/werf/pkg/ssh_agent"
	"github.com/flant/werf/pkg/tmp_manager"
	"github.com/flant/werf/pkg/true_git"
	"github.com/flant/werf/pkg/werf"
)

type CmdDataType struct {
	Shell            bool
	RawDockerOptions string

	DockerOptions []string
	DockerCommand []string
	ImageName     string
}

var CmdData CmdDataType
var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	return NewCmdWithData(&CommonCmdData)
}

func NewCmdWithData(commonCmdData *common.CmdData) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "run [options] [IMAGE_NAME] [-- COMMAND ARG...]",
		Short:                 "Run container for specified project image",
		DisableFlagsInUseLine: true,
		Example: `  # Run specified image
  $ werf run application

  # Run image with predefined docker run options and command for debug
  $ werf run --shell

  # Run image with specified docker run options and command
  $ werf run --docker-options="-d -p 5000:5000 --restart=always --name registry" -- /app/run.sh

  # Print a resulting docker run command
  $ werf run --shell --dry-run
  docker run -ti --rm image-stage-test:1ffe83860127e68e893b6aece5b0b7619f903f8492a285c6410371c87018c6a0 /bin/sh`,
		Annotations: map[string]string{
			common.DisableOptionsInUseLineAnno: "1",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := processArgs(cmd, args); err != nil {
				cmd.Help()
				fmt.Println()
				return err
			}

			if CmdData.RawDockerOptions != "" {
				CmdData.DockerOptions = strings.Split(CmdData.RawDockerOptions, " ")
			}

			if CmdData.Shell {
				if len(CmdData.DockerOptions) == 0 && len(CmdData.DockerCommand) == 0 {
					CmdData.DockerOptions = []string{"-ti", "--rm"}
					CmdData.DockerCommand = []string{"/bin/sh"}
				} else {
					cmd.Help()
					fmt.Println()
					return fmt.Errorf("shell option cannot be used with other docker run arguments")
				}
			}

			return runRun(commonCmdData)
		},
	}

	common.SetupDir(commonCmdData, cmd)
	common.SetupTmpDir(commonCmdData, cmd)
	common.SetupHomeDir(commonCmdData, cmd)
	common.SetupSSHKey(commonCmdData, cmd)

	common.SetupStagesRepo(commonCmdData, cmd)
	common.SetupDockerConfig(&CommonCmdData, cmd)

	common.SetupDryRun(&CommonCmdData, cmd)

	cmd.Flags().BoolVarP(&CmdData.Shell, "shell", "", false, "Use predefined docker options and command for debug")
	cmd.Flags().StringVarP(&CmdData.RawDockerOptions, "docker-options", "", "", "Define docker run options")

	return cmd
}

func processArgs(cmd *cobra.Command, args []string) error {
	doubleDashInd := cmd.ArgsLenAtDash()
	doubleDashExist := cmd.ArgsLenAtDash() != -1

	if doubleDashExist {
		if doubleDashInd == len(args) {
			return fmt.Errorf("unsupported position args format")
		}

		switch doubleDashInd {
		case 0:
			CmdData.DockerCommand = args[doubleDashInd:]
		case 1:
			CmdData.ImageName = args[0]
			CmdData.DockerCommand = args[doubleDashInd:]
		default:
			return fmt.Errorf("unsupported position args format")
		}
	} else {
		switch len(args) {
		case 0:
		case 1:
			CmdData.ImageName = args[0]
		default:
			return fmt.Errorf("unsupported position args format")
		}
	}

	return nil
}

func runRun(commonCmdData *common.CmdData) error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
		return err
	}

	if err := true_git.Init(true_git.Options{Out: logger.GetOutStream(), Err: logger.GetErrStream()}); err != nil {
		return err
	}

	if err := docker.Init(*CommonCmdData.DockerConfig); err != nil {
		return err
	}

	projectDir, err := common.GetProjectDir(commonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}
	common.LogProjectDir(projectDir)

	werfConfig, err := common.GetWerfConfig(projectDir)
	if err != nil {
		return fmt.Errorf("cannot parse werf config: %s", err)
	}

	projectTmpDir, err := tmp_manager.CreateProjectDir()
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}
	defer tmp_manager.ReleaseProjectDir(projectTmpDir)

	_, err = common.GetStagesRepo(commonCmdData)
	if err != nil {
		return err
	}

	if err := ssh_agent.Init(*commonCmdData.SSHKeys); err != nil {
		return fmt.Errorf("cannot initialize ssh agent: %s", err)
	}
	defer func() {
		err := ssh_agent.Terminate()
		if err != nil {
			logger.LogErrorF("WARNING: ssh agent termination failed: %s\n", err)
		}
	}()

	if !werfConfig.HasImage(CmdData.ImageName) {
		return fmt.Errorf("image '%s' is not defined in werf.yaml", CmdData.ImageName)
	}

	c := build.NewConveyor(werfConfig, []string{CmdData.ImageName}, projectDir, projectTmpDir, ssh_agent.SSHAuthSock)

	if err = c.ShouldBeBuilt(); err != nil {
		return err
	}

	dockerImageName := c.GetImageLatestStageImageName(CmdData.ImageName)
	var dockerRunArgs []string
	dockerRunArgs = append(dockerRunArgs, CmdData.DockerOptions...)
	dockerRunArgs = append(dockerRunArgs, dockerImageName)
	dockerRunArgs = append(dockerRunArgs, CmdData.DockerCommand...)

	if CommonCmdData.DryRun {
		fmt.Printf("docker run %s\n", strings.Join(dockerRunArgs, " "))
	} else {
		docker.CliRun(dockerRunArgs...)
	}

	return nil
}
