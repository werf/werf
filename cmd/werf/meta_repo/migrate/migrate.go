package migrate

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/tmp_manager"
	"github.com/werf/werf/v2/pkg/true_git"
)

var (
	commonCmdData common.CmdData
	removeSource  bool
)

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "migrate",
		DisableFlagsInUseLine: true,
		Short:                 "Move existing project metadata from --from into --to and enable the safeguard.",
		Long:                  "Copy the project's managed-images, image-metadata-by-commit, custom-tag metadata and last-cleanup record from --from into --to, then record the meta-repo marker in --from so subsequent runs are forced to use the same --meta-repo. Copy is verified before any source record is removed.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			return run(ctx)
		},
	})

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigRenderPath(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupGiterminismConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})
	common.SetupSSHKey(&commonCmdData, cmd)

	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	commonCmdData.Repo = common.NewRepoData("from", common.RepoDataOptions{})
	commonCmdData.Repo.SetupCmd(cmd)
	commonCmdData.MetaRepo = common.NewRepoData("to", common.RepoDataOptions{OnlyAddress: true})
	commonCmdData.MetaRepo.SetupCmd(cmd)

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read and write images to --from and --to")
	common.SetupContainerRegistryMirror(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	commonCmdData.SetupPlatform(cmd)
	commonCmdData.SetupDebugTemplates(cmd)
	commonCmdData.SetupAllowIncludesUpdate(cmd)

	cmd.Flags().BoolVar(&removeSource, "remove-source", util.GetBoolEnvironmentDefaultTrue("WERF_REMOVE_SOURCE"), "Delete the original metadata records from --from after they are verified present in --to (default $WERF_REMOVE_SOURCE or true if not specified)")

	return cmd
}

func run(ctx context.Context) error {
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
	})
	if err != nil {
		return fmt.Errorf("component init error: %w", err)
	}

	defer func() {
		if err := tmp_manager.DelegateCleanup(ctx); err != nil {
			logboek.Context(ctx).Warn().LogF("Temporary files cleanup preparation failed: %s\n", err)
		}
	}()

	containerBackend := commonManager.ContainerBackend()

	if _, err = tmp_manager.CreateProjectDir(ctx); err != nil {
		return fmt.Errorf("getting project tmp dir failed: %w", err)
	}

	giterminismManager, err := common.GetGiterminismManager(ctx, &commonCmdData)
	if err != nil {
		return err
	}

	_, werfConfig, err := common.GetOptionalWerfConfig(ctx, &commonCmdData, giterminismManager, common.GetWerfConfigOptions(&commonCmdData, false))
	if err != nil {
		return fmt.Errorf("unable to load werf config: %w", err)
	}
	if werfConfig == nil {
		return fmt.Errorf("run command in the project directory with werf.yaml")
	}
	projectName := werfConfig.Meta.Project

	storageManager, err := common.NewStorageManagerWithOptions(ctx, &common.NewStorageManagerConfig{
		ProjectName:                    projectName,
		ContainerBackend:               containerBackend,
		CmdData:                        &commonCmdData,
		CleanupDisabled:                werfConfig.Meta.Cleanup.DisableCleanup,
		GitHistoryBasedCleanupDisabled: werfConfig.Meta.Cleanup.DisableGitHistoryBasedPolicy,
	}, common.WithSkipMetaRepoSafeguard())
	if err != nil {
		return fmt.Errorf("unable to init storage manager: %w", err)
	}

	src, ok := storageManager.GetStagesStorage().(*storage.RepoStagesStorage)
	if !ok {
		return fmt.Errorf("--from must be a container registry address")
	}
	dst, ok := storageManager.GetMetaStorage().(*storage.RepoStagesStorage)
	if !ok || dst == src {
		return fmt.Errorf("--to must be set to a container registry address different from --from")
	}

	if err := storage.MigrateMetaRepo(ctx, projectName, src, dst, storage.MigrateMetaRepoOptions{RemoveSource: removeSource}); err != nil {
		return fmt.Errorf("unable to migrate metadata for project %q: %w", projectName, err)
	}

	logboek.Context(ctx).Default().LogF("Metadata for project %q migrated to meta-repo %s; safeguard enabled in %s\n", projectName, dst.Address(), src.Address())
	return nil
}
