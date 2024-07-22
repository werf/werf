package compose

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/build"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/git_repo/gitdata"
	"github.com/werf/werf/v2/pkg/giterminism_manager"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/ssh_agent"
	"github.com/werf/werf/v2/pkg/storage/lrumeta"
	"github.com/werf/werf/v2/pkg/storage/manager"
	"github.com/werf/werf/v2/pkg/tmp_manager"
	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

type newCmdOptions struct {
	Use           string
	Example       string
	FollowSupport bool
	ArgsSupport   bool
	ArgsRequired  bool
}

type composeCmdData struct {
	RawComposeOptions        string
	RawComposeCommandOptions string

	ComposeBinPath        string
	ComposeOptions        []string
	ComposeCommandOptions []string
	ComposeCommandArgs    []string

	imagesToProcess build.ImagesToProcess
}

func NewConfigCmd(ctx context.Context) *cobra.Command {
	return newCmd(ctx, "config", &newCmdOptions{
		Use: "config [IMAGE_NAME...] [options] [--docker-compose-options=\"OPTIONS\"] [--docker-compose-command-options=\"OPTIONS\"]",
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
  export WERF_APP_DOCKER_IMAGE_NAME=project:570c59946a7f77873d361efd25a637c4ccde86abf3d3186add19bded-1604928781528
  docker-compose -f docker-compose-test.yml config --resolve-image-digests`,
		FollowSupport: false,
		ArgsSupport:   false,
	})
}

func NewRunCmd(ctx context.Context) *cobra.Command {
	return newCmd(ctx, "run", &newCmdOptions{
		Use: "run [IMAGE_NAME...] [options] [--docker-compose-options=\"OPTIONS\"] [--docker-compose-command-options=\"OPTIONS\"] -- SERVICE [COMMAND] [ARGS...]",
		Example: `  # Print docker-compose command without executing
  $ werf compose run --docker-compose-options="-f docker-compose-test.yml" --docker-compose-command-options="-e TOKEN=123" --dry-run --quiet -- test
  export WERF_TEST_DOCKER_IMAGE_NAME=test:03dc2e0bceb09833f54fcab39e89e6e4137316ebbe544aeec8184420-1620123105753
  docker-compose -f docker-compose-test.yml run -e TOKEN=123 -- test`,
		FollowSupport: false,
		ArgsSupport:   true,
		ArgsRequired:  true,
	})
}

func NewUpCmd(ctx context.Context) *cobra.Command {
	return newCmd(ctx, "up", &newCmdOptions{
		Use: "up [IMAGE_NAME...] [options] [--docker-compose-options=\"OPTIONS\"] [--docker-compose-command-options=\"OPTIONS\"] [--] [SERVICE...]",
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

func NewDownCmd(ctx context.Context) *cobra.Command {
	cmd := newCmd(ctx, "down", &newCmdOptions{
		Use: "down [IMAGE_NAME...] [options] [--docker-compose-options=\"OPTIONS\"] [--docker-compose-command-options=\"OPTIONS\"]",
		Example: `  # Run docker-compose down with forwarded image names
  $ werf compose down

  # Print docker-compose command without executing
  $ werf compose down --docker-compose-options="-f docker-compose-test.yml" --docker-compose-command-options="--rmi=all --remove-orphans" --dry-run --quiet
  export WERF_APP_IMAGE_NAME=localhost:5000/project:570c59946a7f77873d361efd25a637c4ccde86abf3d3186add19bded-1604928781528
  docker-compose -f docker-compose-test.yml down --rmi=all --remove-orphans`,
		FollowSupport: false,
		ArgsSupport:   false,
	})

	cmd.Flag("without-images").DefValue = "true"

	return cmd
}

func newCmd(ctx context.Context, composeCmdName string, options *newCmdOptions) *cobra.Command {
	var cmdData composeCmdData
	var commonCmdData common.CmdData

	short := GetComposeShort(composeCmdName).Short
	long := GetComposeDocs(GetComposeShort(composeCmdName).Short).Long

	long = common.GetLongCommandDescription(long)
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   options.Use,
		Short:                 short,
		Long:                  long,
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.DisableOptionsInUseLineAnno: "1",
			common.DocsLongMD:                  GetComposeDocs(GetComposeShort(composeCmdName).ShortMD).LongMD,
		},
		Example: options.Example,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			defer global_warnings.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			if options.ArgsSupport {
				processArgs(&cmdData, cmd, args)
			}

			if options.ArgsRequired && len(cmdData.ComposeCommandArgs) == 0 {
				common.PrintHelp(cmd)
				return fmt.Errorf("unsupported position args format")
			}

			if len(cmdData.RawComposeOptions) != 0 {
				cmdData.ComposeOptions = strings.Fields(cmdData.RawComposeOptions)
			}

			if len(cmdData.RawComposeCommandOptions) != 0 {
				cmdData.ComposeCommandOptions = strings.Fields(cmdData.RawComposeCommandOptions)
			}

			return runMain(ctx, composeCmdName, cmdData, commonCmdData, options.FollowSupport)
		},
	})

	commonCmdData.SetupWithoutImages(cmd)

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupGiterminismConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})
	common.SetupSSHKey(&commonCmdData, cmd)

	common.SetupSecondaryStagesStorageOptions(&commonCmdData, cmd)
	common.SetupCacheStagesStorageOptions(&commonCmdData, cmd)
	common.SetupRepoOptions(&commonCmdData, cmd, common.RepoDataOptions{OptionalRepo: true})
	common.SetupFinalRepo(&commonCmdData, cmd)

	common.SetupRequireBuiltImages(&commonCmdData, cmd)

	if options.FollowSupport {
		common.SetupFollow(&commonCmdData, cmd)
	}

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read and pull images from the specified repo")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupInsecureHelmDependencies(&commonCmdData, cmd, true)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	common.SetupContainerRegistryMirror(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupSynchronization(&commonCmdData, cmd)
	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeConfigBase64(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)

	common.SetupDryRun(&commonCmdData, cmd)

	common.SetupVirtualMerge(&commonCmdData, cmd)

	common.SetupDisableAutoHostCleanup(&commonCmdData, cmd)
	common.SetupAllowedDockerStorageVolumeUsage(&commonCmdData, cmd)
	common.SetupAllowedDockerStorageVolumeUsageMargin(&commonCmdData, cmd)
	common.SetupAllowedLocalCacheVolumeUsage(&commonCmdData, cmd)
	common.SetupAllowedLocalCacheVolumeUsageMargin(&commonCmdData, cmd)
	common.SetupDockerServerStoragePath(&commonCmdData, cmd)

	commonCmdData.SetupPlatform(cmd)

	cmd.Flags().StringVarP(&cmdData.RawComposeOptions, "docker-compose-options", "", os.Getenv("WERF_DOCKER_COMPOSE_OPTIONS"), "Define docker-compose options (default $WERF_DOCKER_COMPOSE_OPTIONS)")
	cmd.Flags().StringVarP(&cmdData.RawComposeCommandOptions, "docker-compose-command-options", "", os.Getenv("WERF_DOCKER_COMPOSE_COMMAND_OPTIONS"), "Define docker-compose command options (default $WERF_DOCKER_COMPOSE_COMMAND_OPTIONS)")

	// TODO: delete this flag in v3
	cmd.Flags().StringVarP(&cmdData.ComposeBinPath, "docker-compose-bin-path", "", os.Getenv("WERF_DOCKER_COMPOSE_BIN_PATH"), "DEPRECATED: \"docker compose\" command always used now, this option is ignored. Define docker-compose bin path (default $WERF_DOCKER_COMPOSE_BIN_PATH)")

	return cmd
}

func processArgs(cmdData *composeCmdData, cmd *cobra.Command, args []string) {
	doubleDashInd := cmd.ArgsLenAtDash()
	doubleDashExist := cmd.ArgsLenAtDash() != -1

	if doubleDashExist {
		cmdData.imagesToProcess = build.NewImagesToProcess(args[:doubleDashInd], false)
		cmdData.ComposeCommandArgs = args[doubleDashInd:]
	} else if len(args) != 0 {
		cmdData.imagesToProcess = build.NewImagesToProcess(args, false)
	}
}

func checkDetachDockerComposeOption(cmdData composeCmdData) error {
	for _, value := range cmdData.ComposeCommandOptions {
		if value == "-d" || value == "--detach" {
			return nil
		}
	}

	return fmt.Errorf("the containers must be launched in the background (in follow mode): pass -d/--detach with --docker-compose-command-options option")
}

func runMain(ctx context.Context, dockerComposeCmdName string, cmdData composeCmdData, commonCmdData common.CmdData, followSupport bool) error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %w", err)
	}

	registryMirrors, err := common.GetContainerRegistryMirror(ctx, &commonCmdData)
	if err != nil {
		return fmt.Errorf("get container registry mirrors: %w", err)
	}

	containerBackend, processCtx, err := common.InitProcessContainerBackend(ctx, &commonCmdData, registryMirrors)
	if err != nil {
		return err
	}
	ctx = processCtx

	gitDataManager, err := gitdata.GetHostGitDataManager(ctx)
	if err != nil {
		return fmt.Errorf("error getting host git data manager: %w", err)
	}

	if err := git_repo.Init(gitDataManager); err != nil {
		return err
	}

	if err := image.Init(); err != nil {
		return err
	}

	if err := lrumeta.Init(); err != nil {
		return err
	}

	if err := true_git.Init(ctx, true_git.Options{LiveGitOutput: *commonCmdData.LogDebug}); err != nil {
		return err
	}

	if err := common.DockerRegistryInit(ctx, &commonCmdData, registryMirrors); err != nil {
		return err
	}

	defer func() {
		if err := common.RunAutoHostCleanup(ctx, &commonCmdData, containerBackend); err != nil {
			logboek.Context(ctx).Error().LogF("Auto host cleanup failed: %s\n", err)
		}
	}()

	if err := ssh_agent.Init(ctx, common.GetSSHKey(&commonCmdData)); err != nil {
		return fmt.Errorf("cannot initialize ssh agent: %w", err)
	}
	defer func() {
		err := ssh_agent.Terminate()
		if err != nil {
			logboek.Warn().LogF("WARNING: ssh agent termination failed: %s\n", err)
		}
	}()

	giterminismManager, err := common.GetGiterminismManager(ctx, &commonCmdData)
	if err != nil {
		return err
	}

	common.ProcessLogProjectDir(&commonCmdData, giterminismManager.ProjectDir())

	if followSupport && *commonCmdData.Follow {
		if err := checkDetachDockerComposeOption(cmdData); err != nil {
			return err
		}

		return common.FollowGitHead(ctx, &commonCmdData, func(ctx context.Context, headCommitGiterminismManager giterminism_manager.Interface) error {
			return run(ctx, containerBackend, headCommitGiterminismManager, commonCmdData, cmdData, dockerComposeCmdName)
		})
	} else {
		if err := run(ctx, containerBackend, giterminismManager, commonCmdData, cmdData, dockerComposeCmdName); err != nil {
			// TODO: use docker cli StatusError after switching on docker compose command
			if exitErr, ok := err.(*exec.ExitError); ok {
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					common.TerminateWithError(err.Error(), status.ExitStatus())
				}
			}

			return err
		}

		return nil
	}
}

func run(ctx context.Context, containerBackend container_backend.ContainerBackend, giterminismManager giterminism_manager.Interface, commonCmdData common.CmdData, cmdData composeCmdData, dockerComposeCmdName string) error {
	imagesToProcess := cmdData.imagesToProcess

	_, werfConfig, err := common.GetRequiredWerfConfig(ctx, &commonCmdData, giterminismManager, common.GetWerfConfigOptions(&commonCmdData, true))
	if err != nil {
		return fmt.Errorf("unable to load werf config: %w", err)
	}

	if err := imagesToProcess.CheckImagesExistence(werfConfig); err != nil {
		return err
	}

	var envArray []string
	if imagesToProcess.HaveImagesToProcess(werfConfig) {
		projectName := werfConfig.Meta.Project

		projectTmpDir, err := tmp_manager.CreateProjectDir(ctx)
		if err != nil {
			return fmt.Errorf("getting project tmp dir failed: %w", err)
		}
		defer tmp_manager.ReleaseProjectDir(projectTmpDir)

		stagesStorage, err := common.GetStagesStorage(ctx, containerBackend, &commonCmdData)
		if err != nil {
			return err
		}
		finalStagesStorage, err := common.GetOptionalFinalStagesStorage(ctx, containerBackend, &commonCmdData)
		if err != nil {
			return err
		}
		synchronization, err := common.GetSynchronization(ctx, &commonCmdData, projectName, stagesStorage)
		if err != nil {
			return err
		}
		storageLockManager, err := common.GetStorageLockManager(ctx, synchronization)
		if err != nil {
			return err
		}
		secondaryStagesStorageList, err := common.GetSecondaryStagesStorageList(ctx, stagesStorage, containerBackend, &commonCmdData)
		if err != nil {
			return err
		}
		cacheStagesStorageList, err := common.GetCacheStagesStorageList(ctx, containerBackend, &commonCmdData)
		if err != nil {
			return err
		}

		storageManager := manager.NewStorageManager(projectName, stagesStorage, finalStagesStorage, secondaryStagesStorageList, cacheStagesStorageList, storageLockManager)

		logboek.Context(ctx).Info().LogOptionalLn()

		conveyorOptions, err := common.GetConveyorOptions(ctx, &commonCmdData, imagesToProcess)
		if err != nil {
			return err
		}

		conveyorWithRetry := build.NewConveyorWithRetryWrapper(werfConfig, giterminismManager, giterminismManager.ProjectDir(), projectTmpDir, ssh_agent.SSHAuthSock, containerBackend, storageManager, storageLockManager, conveyorOptions)
		defer conveyorWithRetry.Terminate()

		if err := conveyorWithRetry.WithRetryBlock(ctx, func(c *build.Conveyor) error {
			if common.GetRequireBuiltImages(ctx, &commonCmdData) {
				if err := c.ShouldBeBuilt(ctx, build.ShouldBeBuiltOptions{}); err != nil {
					return err
				}
			} else {
				if err := c.Build(ctx, build.BuildOptions{SkipImageMetadataPublication: *commonCmdData.Dev}); err != nil {
					return err
				}
			}

			for _, img := range c.GetExportedImages() {
				if err := c.FetchLastImageStage(ctx, img.TargetPlatform, img.Name); err != nil {
					return err
				}
			}

			envArray = c.GetImagesEnvArray()

			return nil
		}); err != nil {
			return err
		}
	}

	var dockerComposeArgs []string
	dockerComposeArgs = append(dockerComposeArgs, cmdData.ComposeOptions...)
	dockerComposeArgs = append(dockerComposeArgs, dockerComposeCmdName)
	dockerComposeArgs = append(dockerComposeArgs, cmdData.ComposeCommandOptions...)

	if len(cmdData.ComposeCommandArgs) != 0 {
		dockerComposeArgs = append(dockerComposeArgs, "--")
		dockerComposeArgs = append(dockerComposeArgs, cmdData.ComposeCommandArgs...)
	}

	// TODO: use docker SDK instead of host docker binary
	if *commonCmdData.DryRun {
		for _, env := range envArray {
			fmt.Println("export", env)
		}
		fmt.Printf("docker compose %s\n", strings.Join(dockerComposeArgs, " "))
		return nil
	} else {
		dockerComposeArgs = append([]string{"compose"}, dockerComposeArgs...)

		cmd := exec.Command("docker", dockerComposeArgs...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = append(os.Environ(), envArray...)
		return cmd.Run()
	}
}
