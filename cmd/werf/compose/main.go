package compose

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/werf/common-go/pkg/graceful"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/build"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/giterminism_manager"
	"github.com/werf/werf/v2/pkg/tmp_manager"
	"github.com/werf/werf/v2/pkg/true_git"
	werfExec "github.com/werf/werf/v2/pkg/werf/exec"
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

	ImageNameListFromArgs []string
}

func (d *composeCmdData) extractImageNameListFromComposeConfig(ctx context.Context, werfConfig *config.WerfConfig) ([]string, error) {
	// Replace all special characters in image name with empty string to find the same image name in werf config.
	replaceAllFunc := func(s string) string {
		for _, l := range []string{"_", "-", "/", "."} {
			s = strings.ReplaceAll(s, l, "")
		}
		return s
	}

	extractedImageNameList, err := extractImageNamesFromComposeConfig(ctx, d.getComposeFileCustomPathList())
	if err != nil {
		return nil, fmt.Errorf("unable to extract image names from docker-compose file: %w", err)
	}

	configImageNameList := werfConfig.GetImageNameList(false)

	var imageNameList []string
	for _, configImageName := range configImageNameList {
		for _, imageName := range extractedImageNameList {
			if configImageName == imageName {
				imageNameList = append(imageNameList, imageName)
				continue
			}

			if replaceAllFunc(configImageName) == replaceAllFunc(imageName) {
				imageNameList = append(imageNameList, configImageName)
				continue
			}
		}
	}

	return imageNameList, nil
}

func (d *composeCmdData) getComposeFileCustomPathList() []string {
	var result []string
	for ind, value := range d.ComposeOptions {
		if strings.HasPrefix(value, "-f") || strings.HasPrefix(value, "--file") {
			parts := strings.Split(value, "=")
			if len(parts) == 2 {
				result = append(result, parts[1])
			} else if len(d.ComposeOptions) > ind+1 {
				result = append(result, d.ComposeOptions[ind+1])
			}
		}
	}

	return result
}

func extractImageNamesFromComposeConfig(ctx context.Context, customConfigPathList []string) ([]string, error) {
	composeArgs := []string{"compose"}
	for _, p := range customConfigPathList {
		composeArgs = append(composeArgs, "--file", p)
	}
	composeArgs = append(composeArgs, "config", "--no-interpolate")

	cmd := werfExec.CommandContextCancellation(ctx, "docker", composeArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		graceful.Terminate(ctx, err, werfExec.ExitCode(err))
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return nil, fmt.Errorf("error running command %q: %w\n\nStdout:\n%s\nStderr:\n%s", cmd, err, stdout.String(), stderr.String())
		}
		return nil, fmt.Errorf("error running command %q: %w", cmd, err)
	}

	output := stdout.Bytes()

	// Matches $WERF_<IMAGE_NAME>_DOCKER_IMAGE_NAME and ${WERF_<IMAGE_NAME>_DOCKER_IMAGE_NAME}.
	re := regexp.MustCompile(`\${?WERF_(.*)_DOCKER_IMAGE_NAME}?`)

	var imageNames []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Ignore commented lines.
		if strings.HasPrefix(line, "#") {
			continue
		}

		matches := re.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				imageName := strings.ToLower(match[1])
				imageNames = append(imageNames, imageName)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading compose config: %v", err)
	}

	return imageNames, nil
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

	f := cmd.Flag("stub-tags")
	f.DefValue = "true"
	if !f.Changed {
		err := f.Value.Set("true")
		if err != nil {
			panic(fmt.Sprintf("unable to set stub-tags flag value: %s", err))
		}
	}

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
	commonCmdData.SetupFinalImagesOnly(cmd, false)
	common.SetupStubTags(&commonCmdData, cmd)

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
	common.SetupAllowedBackendStorageVolumeUsage(&commonCmdData, cmd)
	common.SetupAllowedBackendStorageVolumeUsageMargin(&commonCmdData, cmd)
	common.SetupAllowedLocalCacheVolumeUsage(&commonCmdData, cmd)
	common.SetupAllowedLocalCacheVolumeUsageMargin(&commonCmdData, cmd)
	common.SetupBackendStoragePath(&commonCmdData, cmd)
	common.SetupProjectName(&commonCmdData, cmd, false)

	commonCmdData.SetupPlatform(cmd)
	commonCmdData.SetupDebugTemplates(cmd)

	cmd.Flags().StringVarP(&cmdData.RawComposeOptions, "docker-compose-options", "", os.Getenv("WERF_DOCKER_COMPOSE_OPTIONS"), "Define docker-compose options (default $WERF_DOCKER_COMPOSE_OPTIONS)")
	cmd.Flags().StringVarP(&cmdData.RawComposeCommandOptions, "docker-compose-command-options", "", os.Getenv("WERF_DOCKER_COMPOSE_COMMAND_OPTIONS"), "Define docker-compose command options (default $WERF_DOCKER_COMPOSE_COMMAND_OPTIONS)")

	// TODO: delete this flag in v3
	cmd.Flags().StringVarP(&cmdData.ComposeBinPath, "docker-compose-bin-path", "", os.Getenv("WERF_DOCKER_COMPOSE_BIN_PATH"), "DEPRECATED: \"docker compose\" command always used now, this option is ignored. Define docker-compose bin path (default $WERF_DOCKER_COMPOSE_BIN_PATH)")

	commonCmdData.SetupSkipImageSpecStage(cmd)
	commonCmdData.SetupAllowIncludesUpdate(cmd)

	return cmd
}

func processArgs(cmdData *composeCmdData, cmd *cobra.Command, args []string) {
	doubleDashInd := cmd.ArgsLenAtDash()
	doubleDashExist := cmd.ArgsLenAtDash() != -1

	if doubleDashExist {
		cmdData.ImageNameListFromArgs = args[:doubleDashInd]
		cmdData.ComposeCommandArgs = args[doubleDashInd:]
	} else if len(args) != 0 {
		cmdData.ImageNameListFromArgs = args
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
	commonManager, ctx, err := common.InitCommonComponents(ctx, common.InitCommonComponentsOptions{
		Cmd: &commonCmdData,
		InitTrueGitWithOptions: &common.InitTrueGitOptions{
			Options: true_git.Options{LiveGitOutput: *commonCmdData.LogDebug},
		},
		InitDockerRegistry:          true,
		InitProcessContainerBackend: true,
		InitWerf:                    true,
		InitGitDataManager:          true,
		InitManifestCache:           true,
		InitLRUImagesCache:          true,
		InitSSHAgent:                true,
	})
	if err != nil {
		return fmt.Errorf("component init error: %w", err)
	}

	containerBackend := commonManager.ContainerBackend()

	defer func() {
		if err := common.RunAutoHostCleanup(ctx, &commonCmdData, containerBackend); err != nil {
			logboek.Context(ctx).Error().LogF("Auto host cleanup failed: %s\n", err)
		}
	}()

	defer func() {
		commonManager.TerminateSSHAgent()
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

		return common.FollowGitHead(ctx, &commonCmdData, func(ctx context.Context, headCommitGiterminismManager *giterminism_manager.Manager) error {
			return run(ctx, containerBackend, headCommitGiterminismManager, commonCmdData, cmdData, dockerComposeCmdName)
		})
	} else {
		return run(ctx, containerBackend, giterminismManager, commonCmdData, cmdData, dockerComposeCmdName)
	}
}

func run(ctx context.Context, containerBackend container_backend.ContainerBackend, giterminismManager giterminism_manager.Interface, commonCmdData common.CmdData, cmdData composeCmdData, dockerComposeCmdName string) error {
	_, werfConfig, err := common.GetRequiredWerfConfig(ctx, &commonCmdData, giterminismManager, common.GetWerfConfigOptions(&commonCmdData, true))
	if err != nil {
		return fmt.Errorf("unable to load werf config: %w", err)
	}

	var imageNameList []string
	if len(cmdData.ImageNameListFromArgs) != 0 {
		imageNameList = cmdData.ImageNameListFromArgs
	} else {
		imageNameListFromComposeConfig, err := cmdData.extractImageNameListFromComposeConfig(ctx, werfConfig)
		if err != nil {
			return err
		}
		imageNameList = imageNameListFromComposeConfig
	}

	imagesToProcess, err := config.NewImagesToProcess(werfConfig, imageNameList, *commonCmdData.FinalImagesOnly, *commonCmdData.WithoutImages)
	if err != nil {
		return err
	}

	shouldBeBuilt := !*commonCmdData.StubTags

	var envArray []string
	if !imagesToProcess.WithoutImages && shouldBeBuilt {
		common.SetupOndemandKubeInitializer(*commonCmdData.KubeContext, *commonCmdData.KubeConfig, *commonCmdData.KubeConfigBase64, *commonCmdData.KubeConfigPathMergeList)
		if err := common.GetOndemandKubeInitializer().Init(ctx); err != nil {
			return err
		}

		projectName := werfConfig.Meta.Project

		projectTmpDir, err := tmp_manager.CreateProjectDir(ctx)
		if err != nil {
			return fmt.Errorf("getting project tmp dir failed: %w", err)
		}
		defer tmp_manager.ReleaseProjectDir(projectTmpDir)

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

		logboek.Context(ctx).Default().LogOptionalLn()

		conveyorOptions, err := common.GetConveyorOptions(ctx, &commonCmdData, imagesToProcess)
		if err != nil {
			return err
		}

		conveyorWithRetry := build.NewConveyorWithRetryWrapper(werfConfig, giterminismManager, giterminismManager.ProjectDir(), projectTmpDir, containerBackend, storageManager, storageManager.StorageLockManager, conveyorOptions)
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
	} else {
		for _, imageName := range imagesToProcess.ImageNameList {
			envArray = append(envArray, build.GenerateImageEnv(imageName, "STUB:TAG"))
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

		cmd := werfExec.CommandContextCancellation(ctx, "docker", dockerComposeArgs...)

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = append(os.Environ(), envArray...)

		if err := cmd.Run(); err != nil {
			werfExec.TerminateIfCanceled(ctx)
			return fmt.Errorf("error running command %q: %s", cmd, err)
		}
	}

	return nil
}
