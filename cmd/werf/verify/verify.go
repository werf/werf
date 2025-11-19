package verify

import (
	"context"
	"errors"
	"fmt"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/build"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/giterminism_manager"
	"github.com/werf/werf/v2/pkg/signature"
	"github.com/werf/werf/v2/pkg/tmp_manager"
	"github.com/werf/werf/v2/pkg/util/parallel"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, common.SetCommandContext(ctx, &cobra.Command{
		Use:                   GetVerifyDocs().Use,
		Short:                 GetVerifyDocs().Short,
		Example:               GetVerifyDocs().Example,
		Long:                  common.GetLongCommandDescription(GetVerifyDocs().Long),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfDebugAnsibleArgs),
			common.DocsLongMD: GetVerifyDocs().LongMD,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			defer global_warnings.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error {
				return runMain(ctx, args)
			})
		},
	}))

	// common.SetupDir(&commonCmdData, cmd)
	// common.SetupGitWorkTree(&commonCmdData, cmd)
	// common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	// common.SetupConfigPath(&commonCmdData, cmd)
	// common.SetupGiterminismConfigPath(&commonCmdData, cmd)
	// common.SetupEnvironment(&commonCmdData, cmd)

	// common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})
	// common.SetupSSHKey(&commonCmdData, cmd)

	// common.SetupSecondaryStagesStorageOptions(&commonCmdData, cmd)
	// common.SetupCacheStagesStorageOptions(&commonCmdData, cmd)
	// common.SetupRepoOptions(&commonCmdData, cmd, common.RepoDataOptions{OptionalRepo: true})
	// common.SetupFinalRepo(&commonCmdData, cmd)

	// common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and push images into the specified repo, to pull base images")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	common.SetupContainerRegistryMirror(&commonCmdData, cmd)

	common.SetupVerificationOptions(&commonCmdData, cmd)
	common.SetupELFVerificationOptions(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	// common.SetupSynchronization(&commonCmdData, cmd)

	common.SetupParallelOptions(&commonCmdData, cmd, common.DefaultBuildParallelTasksLimit)

	// commonCmdData.SetupDebugTemplates(cmd)

	return cmd
}

func runMain(ctx context.Context, imageNameListFromArgs []string) error {
	global_warnings.PostponeMultiwerfNotUpToDateWarning(ctx)
	_, ctx, err := common.InitCommonComponents(ctx, common.InitCommonComponentsOptions{
		Cmd:      &commonCmdData,
		InitWerf: true,
		// InitGitDataManager: true,
		// InitManifestCache:  true,
		// InitLRUImagesCache: true,
		// InitTrueGitWithOptions: &common.InitTrueGitOptions{
		//	 Options: true_git.Options{LiveGitOutput: *commonCmdData.LogDebug},
		// },
		InitDockerRegistry: true,
		// InitProcessContainerBackend: true,
		// InitSSHAgent:                true,
	})
	if err != nil {
		return fmt.Errorf("component init error: %w", err)
	}

	defer func() {
		if err := tmp_manager.DelegateCleanup(ctx); err != nil {
			logboek.Context(ctx).Warn().LogF("Temporary files cleanup preparation failed: %s\n", err)
		}
	}()

	// defer commonManager.TerminateSSHAgent()

	// giterminismManager, err := common.GetGiterminismManager(ctx, &commonCmdData)
	// if err != nil {
	// 	return err
	// }
	//
	// common.ProcessLogProjectDir(&commonCmdData, giterminismManager.ProjectDir())

	// return run(ctx, commonManager.ContainerBackend(), giterminismManager, imageNameListFromArgs)
	return run(ctx, nil, nil, imageNameListFromArgs)
}

func run(ctx context.Context, containerBackend container_backend.ContainerBackend, giterminismManager giterminism_manager.Interface, imageNameListFromArgs []string) error {
	// _, _, err := common.GetRequiredWerfConfig(ctx, &commonCmdData, giterminismManager, common.GetWerfConfigOptions(&commonCmdData, true))
	// if err != nil {
	//	 return fmt.Errorf("unable to load werf config: %w", err)
	// }

	verifyOptions, err := common.GetVerifyOptions(&commonCmdData)
	if err != nil {
		return fmt.Errorf("verify options error: %w", err)
	}

	if verifyOptions.References = common.GetImageReferences(&commonCmdData); len(verifyOptions.References) == 0 {
		// TODO: only this case supported right now
		return fmt.Errorf("no image references provided via --image-ref option")
	}

	// TODO: this case is not supported yet
	// if verifyOptions.References, err = buildAndGetImageReferencies(ctx, werfConfig, containerBackend, giterminismManager, imageNameListFromArgs); err != nil {
	// 	return fmt.Errorf("unable to build and get image references: %w", err)
	// }

	parallelOptions := parallel.DoTasksOptions{
		MaxNumberOfWorkers: int(common.GetParallelTasksLimit(&commonCmdData)),
	}

	logboek.Context(ctx).Default().LogOptionalLn()
	if err = signature.Verify(ctx, verifyOptions, parallelOptions); err != nil {
		return err
	}

	return nil
}

func buildAndGetImageReferencies(ctx context.Context, werfConfig *config.WerfConfig, containerBackend container_backend.ContainerBackend, giterminismManager giterminism_manager.Interface, imageNameListFromArgs []string) ([]string, error) { //nolint:unused
	projectTmpDir, err := tmp_manager.CreateProjectDir(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting project tmp dir failed: %w", err)
	}

	storageManager, err := common.NewStorageManager(ctx, &common.NewStorageManagerConfig{
		ProjectName:                    werfConfig.Meta.Project,
		ContainerBackend:               containerBackend,
		CmdData:                        &commonCmdData,
		CleanupDisabled:                true,
		GitHistoryBasedCleanupDisabled: true,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to init storage manager: %w", err)
	}

	imagesToProcess, err := config.NewImagesToProcess(werfConfig, imageNameListFromArgs, false, false)
	if err != nil {
		return nil, fmt.Errorf("unable to get images to process: %w", err)
	}

	buildOptions, err := common.GetBuildOptions(ctx, &commonCmdData, werfConfig, imagesToProcess)
	if err != nil {
		return nil, fmt.Errorf("unable to get build options: %w", err)
	}

	// TODO: validate option combinations
	// TODO: defer auto host cleanup

	// TODO: если мы активируем опцию --sign-manifest, то нам не хватает входящих данных, чтобы корректно посчитать хэш стадии pkg/build/stage/sign.go:46
	// buildOptions.ManifestSigningOptions.Enabled = common.GetVerifyManifest(&commonCmdData)
	// buildOptions.ELFSigningOptions.BsignEnabled = common.GetVerifyELFFiles(&commonCmdData)
	// buildOptions.ELFSigningOptions.InHouseEnabled = common.GetVerifyBSignELFFiles(&commonCmdData)

	conveyorOptions, err := common.GetConveyorOptionsWithParallel(ctx, &commonCmdData, imagesToProcess, buildOptions)
	if err != nil {
		return nil, fmt.Errorf("unable to get conveyor options: %w", err)
	}

	// Always print logs.
	conveyorOptions.DeferBuildLog = false

	conveyorWithRetry := build.NewConveyorWithRetryWrapper(werfConfig, giterminismManager, giterminismManager.ProjectDir(), projectTmpDir, containerBackend, storageManager, storageManager.StorageLockManager, conveyorOptions)
	defer conveyorWithRetry.Terminate()

	// Always require built images.
	commonCmdData.RequireBuiltImages = lo.ToPtr(true)

	var reports []*build.ImagesReport
	if err = logboek.Context(ctx).Default().LogProcess("Loading stages from repo").DoError(func() error {
		return conveyorWithRetry.WithRetryBlock(ctx, func(c *build.Conveyor) error {
			if !common.GetRequireBuiltImages(&commonCmdData) {
				return errors.New("option --require-built-images must be set to true")
			}

			shouldBeBuiltOptions, err := common.GetShouldBeBuiltOptions(&commonCmdData, imagesToProcess)
			if err != nil {
				return fmt.Errorf("unable to get should be built options: %w", err)
			}

			if reports, err = c.ShouldBeBuilt(ctx, shouldBeBuiltOptions); err != nil {
				return fmt.Errorf("unable to check images should be built: %w", err)
			}

			return nil
		})
	}); err != nil {
		return nil, err
	}

	return mapImageReportsToImageReferences(reports), nil
}

func mapImageReportsToImageReferences(reports []*build.ImagesReport) []string { //nolint:unused
	return lo.Flatten(lo.Map(reports, func(report *build.ImagesReport, _ int) []string {
		return lo.MapToSlice(report.Images, func(_ string, img build.ReportImageRecord) string {
			return img.DockerImageName
		})
	}))
}
