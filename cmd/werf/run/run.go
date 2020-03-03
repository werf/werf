package run

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/flant/logboek"
	"github.com/flant/shluz"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/build"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/logging"
	"github.com/flant/werf/pkg/ssh_agent"
	"github.com/flant/werf/pkg/tmp_manager"
	"github.com/flant/werf/pkg/true_git"
	"github.com/flant/werf/pkg/werf"
)

type cmdDataType struct {
	Shell            bool
	Bash             bool
	RawDockerOptions string

	DockerOptions []string
	DockerCommand []string
	ImageName     string
}

var cmdData cmdDataType
var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "run [options] [IMAGE_NAME] [-- COMMAND ARG...]",
		Short:                 "Run container for specified project image",
		DisableFlagsInUseLine: true,
		Example: `  # Run specified image
  $ werf run --stages-storage :local application

  # Run image with predefined docker run options and command for debug
  $ werf run --stages-storage :local --shell

  # Run image with specified docker run options and command
  $ werf run --stages-storage :local --docker-options="-d -p 5000:5000 --restart=always --name registry" -- /app/run.sh

  # Print a resulting docker run command
  $ werf run --stages-storage :local --shell --dry-run
  docker run -ti --rm image-stage-test:1ffe83860127e68e893b6aece5b0b7619f903f8492a285c6410371c87018c6a0 /bin/sh`,
		Annotations: map[string]string{
			common.DisableOptionsInUseLineAnno: "1",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			if err := processArgs(cmd, args); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			if cmdData.RawDockerOptions != "" {
				cmdData.DockerOptions = strings.Split(cmdData.RawDockerOptions, " ")
			}

			if cmdData.Shell && cmdData.Bash {
				return fmt.Errorf("cannot use --shell and --bash options at the same time!")
			}

			if cmdData.Shell || cmdData.Bash {
				if len(cmdData.DockerOptions) == 0 && len(cmdData.DockerCommand) == 0 {
					cmdData.DockerOptions = []string{"-ti", "--rm"}
					if cmdData.Shell {
						cmdData.DockerCommand = []string{"/bin/sh"}
					}

					if cmdData.Bash {
						cmdData.DockerCommand = []string{"/bin/bash"}
					}
				} else {
					common.PrintHelp(cmd)
					return fmt.Errorf("shell option cannot be used with other docker run arguments")
				}
			}

			return runRun()
		},
	}

	common.SetupDir(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)
	common.SetupSSHKey(&commonCmdData, cmd)

	common.SetupStagesStorage(&commonCmdData, cmd)
	common.SetupStagesStorageLock(&commonCmdData, cmd)
	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read and pull images from the specified stages storage")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupDryRun(&commonCmdData, cmd)

	cmd.Flags().BoolVarP(&cmdData.Shell, "shell", "", false, "Use predefined docker options and command for debug")
	cmd.Flags().BoolVarP(&cmdData.Bash, "bash", "", false, "Use predefined docker options and command for debug")
	cmd.Flags().StringVarP(&cmdData.RawDockerOptions, "docker-options", "", "", "Define docker run options")

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
			cmdData.DockerCommand = args[doubleDashInd:]
		case 1:
			cmdData.ImageName = args[0]
			cmdData.DockerCommand = args[doubleDashInd:]
		default:
			return fmt.Errorf("unsupported position args format")
		}
	} else {
		switch len(args) {
		case 0:
		case 1:
			cmdData.ImageName = args[0]
		default:
			return fmt.Errorf("unsupported position args format")
		}
	}

	return nil
}

func runRun() error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := shluz.Init(filepath.Join(werf.GetServiceDir(), "locks")); err != nil {
		return err
	}

	if err := true_git.Init(true_git.Options{Out: logboek.GetOutStream(), Err: logboek.GetErrStream()}); err != nil {
		return err
	}

	if err := docker_registry.Init(docker_registry.Options{InsecureRegistry: *commonCmdData.InsecureRegistry, SkipTlsVerifyRegistry: *commonCmdData.SkipTlsVerifyRegistry}); err != nil {
		return err
	}

	if err := docker.Init(*commonCmdData.DockerConfig, *commonCmdData.LogVerbose, *commonCmdData.LogDebug); err != nil {
		return err
	}

	projectDir, err := common.GetProjectDir(&commonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	common.ProcessLogProjectDir(&commonCmdData, projectDir)

	werfConfig, err := common.GetRequiredWerfConfig(projectDir, false)
	if err != nil {
		return fmt.Errorf("unable to load werf config: %s", err)
	}

	projectTmpDir, err := tmp_manager.CreateProjectDir()
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}
	defer tmp_manager.ReleaseProjectDir(projectTmpDir)

	_, err = common.GetStagesStorage(&commonCmdData)
	if err != nil {
		return err
	}

	_, err = common.GetStagesStorageLock(&commonCmdData)
	if err != nil {
		return err
	}

	if err := ssh_agent.Init(*commonCmdData.SSHKeys); err != nil {
		return fmt.Errorf("cannot initialize ssh agent: %s", err)
	}
	defer func() {
		err := ssh_agent.Terminate()
		if err != nil {
			logboek.LogWarnF("WARNING: ssh agent termination failed: %s\n", err)
		}
	}()

	imageName := cmdData.ImageName
	if imageName == "" && len(werfConfig.GetAllImages()) == 1 {
		imageName = werfConfig.GetAllImages()[0].GetName()
	}

	if !werfConfig.HasImage(imageName) {
		return fmt.Errorf("image '%s' is not defined in werf.yaml", logging.ImageLogName(imageName, false))
	}

	logboek.Info.LogOptionalLn()
	c := build.NewConveyor(werfConfig, []string{imageName}, projectDir, projectTmpDir, ssh_agent.SSHAuthSock)
	defer c.Terminate()

	if err = c.ShouldBeBuilt(); err != nil {
		return err
	}

	dockerImageName := c.GetImageLastStageImageName(imageName)
	var dockerRunArgs []string
	dockerRunArgs = append(dockerRunArgs, cmdData.DockerOptions...)
	dockerRunArgs = append(dockerRunArgs, dockerImageName)
	dockerRunArgs = append(dockerRunArgs, cmdData.DockerCommand...)

	if *commonCmdData.DryRun {
		fmt.Printf("docker run %s\n", strings.Join(dockerRunArgs, " "))
	} else {
		return logboek.WithRawStreamsOutputModeOn(func() error {
			return common.WithoutTerminationSignalsTrap(func() error {
				return docker.CliRun_LiveOutput(dockerRunArgs...)
			})
		})
	}

	return nil
}
