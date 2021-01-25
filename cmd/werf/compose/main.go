package compose

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/werf/werf/pkg/giterminism_manager"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/build"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/ssh_agent"
	"github.com/werf/werf/pkg/storage/manager"
	"github.com/werf/werf/pkg/tmp_manager"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/werf"
	"github.com/werf/werf/pkg/werf/global_warnings"
)

type newCmdOptions struct {
	Use           string
	Example       string
	FollowSupport bool
	ArgsSupport   bool
}

type cmdDataType struct {
	RawComposeOptions        string
	RawComposeCommandOptions string

	ComposeBinPath        string
	ComposeOptions        []string
	ComposeCommandOptions []string
	ComposeServices       []string
}

func NewConfigCmd() *cobra.Command {
	return newCmd("config", &newCmdOptions{
		Use: "config [options] [--docker-compose-options=\"OPTIONS\"] [--docker-compose-command-options=\"OPTIONS\"]",
		Example: `  # Render compose file
  $ werf compose config --repo localhost:5000/test --quiet
  version: '3.8'
  services:
    web:
      image: localhost:5000/project:570c59946a7f77873d361efd25a637c4ccde86abf3d3186add19bded-1604928781528
      ports:
      - published: 5000
        target: 5000

  # Print docker-compose command without executing
  $ werf compose config --docker-compose-options="-f docker-compose-test.yml" --docker-compose-command-options="--resolve-image-digests" --dry-run --quiet
  export WERF_APP_DOCKER_IMAGE_NAME=localhost:5000/project:570c59946a7f77873d361efd25a637c4ccde86abf3d3186add19bded-1604928781528
  docker-compose -f docker-compose-test.yml config --resolve-image-digests`,
		FollowSupport: false,
		ArgsSupport:   false,
	})
}

func NewUpCmd() *cobra.Command {
	return newCmd("up", &newCmdOptions{
		Use: "up [options] [--docker-compose-options=\"OPTIONS\"] [--docker-compose-command-options=\"OPTIONS\"] [--] [SERVICE...]",
		Example: `  # Run docker-compose up with forwarded image names
  $ werf compose up

  # Follow git HEAD and run docker-compose up for each new commit
  $ werf compose up --follow --docker-compose-command-options="-d"

  # Print docker-compose command without executing
  $ werf compose up --docker-compose-options="-f docker-compose-test.yml" --docker-compose-command-options="--abort-on-container-exit -t 20" --dry-run --quiet
  export WERF_APP_DOCKER_IMAGE_NAME=localhost:5000/project:570c59946a7f77873d361efd25a637c4ccde86abf3d3186add19bded-1604928781528
  docker-compose -f docker-compose-test.yml up --abort-on-container-exit -t 20`,
		FollowSupport: true,
		ArgsSupport:   true,
	})
}

func NewDownCmd() *cobra.Command {
	return newCmd("down", &newCmdOptions{
		Use: "down [options] [--docker-compose-options=\"OPTIONS\"] [--docker-compose-command-options=\"OPTIONS\"]",
		Example: `  # Run docker-compose down with forwarded image names
  $ werf compose down

  # Print docker-compose command without executing
  $ werf compose down --docker-compose-options="-f docker-compose-test.yml" --docker-compose-command-options="--rmi=all --remove-orphans" --dry-run --quiet
  export WERF_APP_IMAGE_NAME=localhost:5000/project:570c59946a7f77873d361efd25a637c4ccde86abf3d3186add19bded-1604928781528
  docker-compose -f docker-compose-test.yml down --rmi=all --remove-orphans`,
		FollowSupport: false,
		ArgsSupport:   false,
	})
}

func newCmd(composeCmdName string, options *newCmdOptions) *cobra.Command {
	var cmdData cmdDataType
	var commonCmdData common.CmdData

	short := fmt.Sprintf("Run docker-compose %s command with forwarded image names", composeCmdName)
	long := short
	long += `

Image environment name format: $WERF_<FORMATTED_WERF_IMAGE_NAME>_DOCKER_IMAGE_NAME ($WERF_DOCKER_IMAGE_NAME for nameless image).
<FORMATTED_WERF_IMAGE_NAME> is werf image name from werf.yaml modified according to the following rules:
- all characters are uppercase (app -> APP);
- charset /- is replaced with _ (DEV/APP-FRONTEND -> DEV_APP_FRONTEND).`
	long = common.GetLongCommandDescription(long)
	cmd := &cobra.Command{
		Use:                   options.Use,
		Short:                 short,
		Long:                  long,
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.DisableOptionsInUseLineAnno: "1",
		},
		Example: options.Example,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := common.BackgroundContext()
			defer global_warnings.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			if options.ArgsSupport {
				if err := processArgs(cmdData, cmd, args); err != nil {
					common.PrintHelp(cmd)
					return err
				}
			}

			if len(cmdData.RawComposeOptions) != 0 {
				cmdData.ComposeOptions = strings.Fields(cmdData.RawComposeOptions)
			}

			if len(cmdData.RawComposeCommandOptions) != 0 {
				cmdData.ComposeCommandOptions = strings.Fields(cmdData.RawComposeCommandOptions)
			}

			if err := checkComposeBin(cmdData); err != nil {
				return err
			}

			return runMain(composeCmdName, cmdData, commonCmdData, options.FollowSupport)
		},
	}

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupGiterminismInspectorOptions(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)
	common.SetupSSHKey(&commonCmdData, cmd)

	common.SetupSecondaryStagesStorageOptions(&commonCmdData, cmd)
	common.SetupStagesStorageOptions(&commonCmdData, cmd)

	common.SetupSkipBuild(&commonCmdData, cmd)

	if options.FollowSupport {
		common.SetupFollow(&commonCmdData, cmd)
	}

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read and pull images from the specified repo")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupSynchronization(&commonCmdData, cmd)
	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeConfigBase64(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)

	common.SetupDryRun(&commonCmdData, cmd)

	common.SetupVirtualMerge(&commonCmdData, cmd)
	common.SetupVirtualMergeFromCommit(&commonCmdData, cmd)
	common.SetupVirtualMergeIntoCommit(&commonCmdData, cmd)

	cmd.Flags().StringVarP(&cmdData.RawComposeOptions, "docker-compose-options", "", os.Getenv("WERF_DOCKER_COMPOSE_OPTIONS"), "Define docker-compose options (default $WERF_DOCKER_COMPOSE_OPTIONS)")
	cmd.Flags().StringVarP(&cmdData.RawComposeCommandOptions, "docker-compose-command-options", "", os.Getenv("WERF_DOCKER_COMPOSE_COMMAND_OPTIONS"), "Define docker-compose command options (default $WERF_DOCKER_COMPOSE_COMMAND_OPTIONS)")
	cmd.Flags().StringVarP(&cmdData.ComposeBinPath, "docker-compose-bin-path", "", os.Getenv("WERF_DOCKER_COMPOSE_BIN_PATH"), "Define docker-compose bin path (default $WERF_DOCKER_COMPOSE_BIN_PATH)")

	return cmd
}

func checkComposeBin(cmdData cmdDataType) error {
	dockerComposeBinPath := "docker-compose"
	if cmdData.ComposeBinPath != "" {
		dockerComposeBinPath = cmdData.ComposeBinPath
	}

	if _, err := exec.LookPath(dockerComposeBinPath); err != nil {
		return fmt.Errorf("%s: %s", dockerComposeBinPath, err)
	}
	return nil
}

func processArgs(cmdData cmdDataType, cmd *cobra.Command, args []string) error {
	doubleDashInd := cmd.ArgsLenAtDash()
	doubleDashExist := cmd.ArgsLenAtDash() != -1

	if doubleDashExist {
		if doubleDashInd == len(args) {
			return fmt.Errorf("unsupported position args format")
		}

		switch doubleDashInd {
		case 0:
			cmdData.ComposeServices = args[doubleDashInd:]
		default:
			return fmt.Errorf("unsupported position args format")
		}
	} else if len(args) != 0 {
		return fmt.Errorf("unsupported position args format")
	}

	return nil
}

func checkDetachDockerComposeOption(cmdData cmdDataType) error {
	for _, value := range cmdData.ComposeCommandOptions {
		if value == "-d" || value == "--detach" {
			return nil
		}
	}

	return fmt.Errorf("the containers must be launched in the background (in follow mode): pass -d/--detach with --docker-compose-command-options option")
}

func runMain(dockerComposeCmdName string, cmdData cmdDataType, commonCmdData common.CmdData, followSupport bool) error {
	ctx := common.BackgroundContext()

	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := common.InitGiterminismInspector(&commonCmdData); err != nil {
		return err
	}

	if err := git_repo.Init(); err != nil {
		return err
	}

	if err := image.Init(); err != nil {
		return err
	}

	if err := true_git.Init(true_git.Options{LiveGitOutput: *commonCmdData.LogVerbose || *commonCmdData.LogDebug}); err != nil {
		return err
	}

	if err := common.DockerRegistryInit(&commonCmdData); err != nil {
		return err
	}

	if err := docker.Init(ctx, *commonCmdData.DockerConfig, *commonCmdData.LogVerbose, *commonCmdData.LogDebug); err != nil {
		return err
	}

	ctxWithDockerCli, err := docker.NewContext(ctx)
	if err != nil {
		return err
	}
	ctx = ctxWithDockerCli

	giterminismManager, err := common.GetGiterminismManager(&commonCmdData)
	if err != nil {
		return err
	}

	common.ProcessLogProjectDir(&commonCmdData, giterminismManager.ProjectDir())

	if err := ssh_agent.Init(ctx, common.GetSSHKey(&commonCmdData)); err != nil {
		return fmt.Errorf("cannot initialize ssh agent: %s", err)
	}
	defer func() {
		err := ssh_agent.Terminate()
		if err != nil {
			logboek.Warn().LogF("WARNING: ssh agent termination failed: %s\n", err)
		}
	}()

	if followSupport && *commonCmdData.Follow {
		if err := checkDetachDockerComposeOption(cmdData); err != nil {
			return err
		}

		return common.FollowGitHead(ctx, &commonCmdData, func(ctx context.Context) error {
			return run(dockerComposeCmdName, cmdData, commonCmdData, ctx, giterminismManager)
		})
	} else {
		return run(dockerComposeCmdName, cmdData, commonCmdData, ctx, giterminismManager)
	}
}

func run(dockerComposeCmdName string, cmdData cmdDataType, commonCmdData common.CmdData, ctx context.Context, giterminismManager giterminism_manager.Interface) error {
	werfConfig, err := common.GetRequiredWerfConfig(ctx, &commonCmdData, giterminismManager, common.GetWerfConfigOptions(&commonCmdData, true))
	if err != nil {
		return fmt.Errorf("unable to load werf config: %s", err)
	}

	projectName := werfConfig.Meta.Project

	projectTmpDir, err := tmp_manager.CreateProjectDir(ctx)
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}
	defer tmp_manager.ReleaseProjectDir(projectTmpDir)

	stagesStorageAddress := common.GetOptionalStagesStorageAddress(&commonCmdData)
	containerRuntime := &container_runtime.LocalDockerServerRuntime{} // TODO
	stagesStorage, err := common.GetStagesStorage(stagesStorageAddress, containerRuntime, &commonCmdData)
	if err != nil {
		return err
	}
	synchronization, err := common.GetSynchronization(ctx, &commonCmdData, projectName, stagesStorage)
	if err != nil {
		return err
	}
	stagesStorageCache, err := common.GetStagesStorageCache(synchronization)
	if err != nil {
		return err
	}
	storageLockManager, err := common.GetStorageLockManager(ctx, synchronization)
	if err != nil {
		return err
	}
	secondaryStagesStorageList, err := common.GetSecondaryStagesStorageList(stagesStorage, containerRuntime, &commonCmdData)
	if err != nil {
		return err
	}

	storageManager := manager.NewStorageManager(projectName, stagesStorage, secondaryStagesStorageList, storageLockManager, stagesStorageCache)

	logboek.Context(ctx).Info().LogOptionalLn()

	conveyorWithRetry := build.NewConveyorWithRetryWrapper(werfConfig, giterminismManager, []string{}, giterminismManager.ProjectDir(), projectTmpDir, ssh_agent.SSHAuthSock, containerRuntime, storageManager, storageLockManager, common.GetConveyorOptions(&commonCmdData))
	defer conveyorWithRetry.Terminate()

	var envArray []string
	if err := conveyorWithRetry.WithRetryBlock(ctx, func(c *build.Conveyor) error {
		if *commonCmdData.SkipBuild {
			if err := c.ShouldBeBuilt(ctx); err != nil {
				return err
			}
		} else {
			if err := c.Build(ctx, build.BuildOptions{}); err != nil {
				return err
			}
		}

		envArray = c.GetImagesEnvArray()

		return nil
	}); err != nil {
		return err
	}

	var dockerComposeArgs []string
	dockerComposeArgs = append(dockerComposeArgs, cmdData.ComposeOptions...)
	dockerComposeArgs = append(dockerComposeArgs, dockerComposeCmdName)
	dockerComposeArgs = append(dockerComposeArgs, cmdData.ComposeCommandOptions...)

	if len(cmdData.ComposeServices) != 0 {
		dockerComposeArgs = append(dockerComposeArgs, "--")
		dockerComposeArgs = append(dockerComposeArgs, cmdData.ComposeServices...)
	}

	if *commonCmdData.DryRun {
		for _, env := range envArray {
			fmt.Println("export", env)
		}
		fmt.Printf("docker-compose %s\n", strings.Join(dockerComposeArgs, " "))
		return nil
	} else {
		cmd := exec.Command("docker-compose", dockerComposeArgs...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = append(os.Environ(), envArray...)
		return cmd.Run()
	}
}
