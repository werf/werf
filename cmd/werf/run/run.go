package run

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/docker/cli/cli"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/spf13/cobra"

	"github.com/werf/common-go/pkg/graceful"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/build"
	"github.com/werf/werf/v2/pkg/buildah"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker"
	"github.com/werf/werf/v2/pkg/giterminism_manager"
	"github.com/werf/werf/v2/pkg/tmp_manager"
	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

type cmdDataType struct {
	Shell            bool
	Bash             bool
	RawDockerOptions string

	DockerOptions []string
	DockerCommand []string
	ImageName     string
}

var (
	cmdData       cmdDataType
	commonCmdData common.CmdData
)

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "run [options] [IMAGE_NAME] [-- COMMAND ARG...]",
		Short:                 "Run container for project image",
		Long:                  common.GetLongCommandDescription(GetRunDocs().Long),
		DisableFlagsInUseLine: true,
		Example: `  # Run specified image and remove after execution
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
			common.DocsLongMD:                  GetRunDocs().LongMD,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if mode, _, err := common.GetBuildahMode(); err != nil {
				return err
			} else if *mode != buildah.ModeDisabled {
				return fmt.Errorf(`command "werf run" is not implemented for Buildah mode`)
			}

			ctx := cmd.Context()

			defer global_warnings.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			if err := processArgs(cmd, args); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			if cmdData.RawDockerOptions != "" {
				cmdData.DockerOptions = strings.Fields(cmdData.RawDockerOptions)
			}

			if cmdData.Shell && cmdData.Bash {
				return fmt.Errorf("cannot use --shell and --bash options at the same time")
			}

			if cmdData.Shell || cmdData.Bash {
				if len(cmdData.DockerOptions) == 0 && len(cmdData.DockerCommand) == 0 {
					cmdData.DockerOptions = []string{"-ti", "--rm"}
					if cmdData.Shell {
						cmdData.DockerOptions = append(cmdData.DockerOptions, "--entrypoint=/bin/sh")
					}

					if cmdData.Bash {
						cmdData.DockerOptions = append(cmdData.DockerOptions, "--entrypoint=/bin/bash")
					}
				} else {
					common.PrintHelp(cmd)
					return fmt.Errorf("shell option cannot be used with other docker run arguments")
				}
			} else if len(cmdData.DockerOptions) == 0 {
				cmdData.DockerOptions = append(cmdData.DockerOptions, "--rm")
			}

			return runMain(ctx)
		},
	})

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

	common.SetupAnnotateLayersWithDmVerityRootHash(&commonCmdData, cmd)
	common.SetupSigningOptions(&commonCmdData, cmd)
	common.SetupELFSigningOptions(&commonCmdData, cmd)
	common.SetupRequireBuiltImages(&commonCmdData, cmd)

	common.SetupFollow(&commonCmdData, cmd)

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

	commonCmdData.SetupPlatform(cmd)

	cmd.Flags().BoolVarP(&cmdData.Shell, "shell", "", false, "Use predefined docker options and command for debug")
	cmd.Flags().BoolVarP(&cmdData.Bash, "bash", "", false, "Use predefined docker options and command for debug")
	cmd.Flags().StringVarP(&cmdData.RawDockerOptions, "docker-options", "", os.Getenv("WERF_DOCKER_OPTIONS"), "Define docker run options (default $WERF_DOCKER_OPTIONS)")

	commonCmdData.SetupSkipImageSpecStage(cmd)
	commonCmdData.SetupDebugTemplates(cmd)
	commonCmdData.SetupAllowIncludesUpdate(cmd)

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

func checkDetachDockerOption() error {
	for _, value := range cmdData.DockerOptions {
		if value == "-d" || value == "--detach" {
			return nil
		}
	}

	return fmt.Errorf("the container must be launched in the background (in follow mode): pass -d/--detach with --docker-options option")
}

func getContainerName() string {
	for ind, value := range cmdData.DockerOptions {
		if value == "--name" {
			if ind+1 < len(cmdData.DockerOptions) {
				return cmdData.DockerOptions[ind+1]
			}
		} else if strings.HasPrefix(value, "--name=") {
			return strings.TrimPrefix(value, "--name=")
		}
	}

	return ""
}

func runMain(ctx context.Context) error {
	global_warnings.PostponeMultiwerfNotUpToDateWarning(ctx)
	commonManager, ctx, err := common.InitCommonComponents(ctx, common.InitCommonComponentsOptions{
		Cmd: &commonCmdData,
		InitTrueGitWithOptions: &common.InitTrueGitOptions{
			Options: true_git.Options{LiveGitOutput: *commonCmdData.LogDebug},
		},
		InitDockerRegistry:           true,
		InitProcessContainerBackend:  true,
		InitWerf:                     true,
		InitGitDataManager:           true,
		InitManifestCache:            true,
		InitLRUImagesCache:           true,
		InitSSHAgent:                 true,
		SetupOndemandKubeInitializer: true,
	})
	if err != nil {
		return fmt.Errorf("component init error: %w", err)
	}

	containerBackend := commonManager.ContainerBackend()

	defer func() {
		commonManager.TerminateSSHAgent()
	}()

	giterminismManager, err := common.GetGiterminismManager(ctx, &commonCmdData)
	if err != nil {
		return err
	}

	common.ProcessLogProjectDir(&commonCmdData, giterminismManager.ProjectDir())

	if *commonCmdData.Follow {
		if cmdData.Shell || cmdData.Bash {
			return fmt.Errorf("follow mode does not work with --shell and --bash options")
		}

		if err := checkDetachDockerOption(); err != nil {
			return err
		}

		containerName := getContainerName()
		if containerName == "" {
			return fmt.Errorf("follow mode does not work without specific container name: pass --name=CONTAINER_NAME with --docker-options option")
		}

		return common.FollowGitHead(ctx, &commonCmdData, func(ctx context.Context, headCommitGiterminismManager *giterminism_manager.Manager) error {
			if err := safeDockerCliRmFunc(ctx, containerName); err != nil {
				return err
			}

			if err := run(ctx, containerBackend, headCommitGiterminismManager); err != nil {
				return err
			}

			go func() {
				time.Sleep(500 * time.Millisecond)
				fmt.Printf("Attaching to container %s ...\n", containerName)

				resp, err := docker.ContainerAttach(ctx, containerName, types.ContainerAttachOptions{
					Stream: true,
					Stdout: true,
					Stderr: true,
					Logs:   true,
				})
				if err != nil {
					_, _ = fmt.Fprintln(os.Stderr, "WARNING:", err)
				}

				if _, err := stdcopy.StdCopy(os.Stdout, os.Stderr, resp.Reader); err != nil {
					_, _ = fmt.Fprintln(os.Stderr, "WARNING:", err)
				}
			}()

			return nil
		})
	} else {
		if err := run(ctx, containerBackend, giterminismManager); err != nil {
			if statusErr, ok := err.(cli.StatusError); ok {
				graceful.Terminate(ctx, err, statusErr.StatusCode)
				return err
			}

			return err
		}

		return nil
	}
}

func run(ctx context.Context, containerBackend container_backend.ContainerBackend, giterminismManager giterminism_manager.Interface) error {
	_, werfConfig, err := common.GetRequiredWerfConfig(ctx, &commonCmdData, giterminismManager, common.GetWerfConfigOptions(&commonCmdData, false))
	if err != nil {
		return fmt.Errorf("unable to load werf config: %w", err)
	}

	projectName := werfConfig.Meta.Project

	projectTmpDir, err := tmp_manager.CreateProjectDir(ctx)
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %w", err)
	}
	defer tmp_manager.ReleaseProjectDir(projectTmpDir)

	imageName := cmdData.ImageName
	if imageName == "" && len(werfConfig.Images(true)) == 1 {
		// The only final image by default.
		imageName = werfConfig.Images(true)[0].GetName()
	}

	imagesToProcess, err := config.NewImagesToProcess(werfConfig, []string{imageName}, false, false)
	if err != nil {
		return err
	}

	storageManager, err := common.NewStorageManager(ctx, &common.NewStorageManagerConfig{
		ProjectName:                    projectName,
		ContainerBackend:               containerBackend,
		CmdData:                        &commonCmdData,
		CleanupDisabled:                werfConfig.Meta.Cleanup.DisableCleanup,
		GitHistoryBasedCleanupDisabled: werfConfig.Meta.Cleanup.DisableGitHistoryBasedPolicy,
	})
	if err != nil {
		return fmt.Errorf("unable to init storage manager: %w", err)
	}

	logboek.Context(ctx).Info().LogOptionalLn()

	conveyorOptions, err := common.GetConveyorOptions(ctx, &commonCmdData, imagesToProcess)
	if err != nil {
		return err
	}

	conveyorWithRetry := build.NewConveyorWithRetryWrapper(werfConfig, giterminismManager, giterminismManager.ProjectDir(), projectTmpDir, containerBackend, storageManager, storageManager.StorageLockManager, conveyorOptions)
	defer conveyorWithRetry.Terminate()

	var dockerImageName string
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

		dockerImageName, err = c.GetFullImageName(ctx, imageName)
		if err != nil {
			return fmt.Errorf("unable to get full name for image %q: %w", imageName, err)
		}
		return nil
	}); err != nil {
		return err
	}

	var dockerRunArgs []string
	dockerRunArgs = append(dockerRunArgs, cmdData.DockerOptions...)
	dockerRunArgs = append(dockerRunArgs, dockerImageName)
	dockerRunArgs = append(dockerRunArgs, cmdData.DockerCommand...)

	if *commonCmdData.DryRun {
		fmt.Printf("docker run %s\n", strings.Join(dockerRunArgs, " "))
		return nil
	} else {
		return logboek.Streams().DoErrorWithoutProxyStreamDataFormatting(func() error {
			return docker.CliRun_ProvidedOutput(ctx, os.Stdout, os.Stderr, dockerRunArgs...)
		})
	}
}

func safeDockerCliRmFunc(ctx context.Context, containerName string) error {
	if exist, err := docker.ContainerExist(ctx, containerName); err != nil {
		return fmt.Errorf("unable to check container %s existence: %w", containerName, err)
	} else if exist {
		logboek.Context(ctx).LogF("Removing container %s ...\n", containerName)
		if err := docker.CliRm(ctx, "-f", containerName); err != nil {
			return fmt.Errorf("unable to remove container %s: %w", containerName, err)
		}
	}

	return nil
}
