package disable

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/tmp_manager"
	"github.com/werf/werf/v2/pkg/true_git"
)

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "disable",
		DisableFlagsInUseLine: true,
		Short:                 "Remove the meta-repo safeguard marker for the project from --repo.",
		Long:                  "Remove the per-project meta-repo marker from --repo so werf no longer forces --meta-repo usage. Metadata already stored in the meta-repo is NOT moved back.",
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

	common.SetupRepoOptions(&commonCmdData, cmd, common.RepoDataOptions{})

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read and write images to --repo")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	common.SetupContainerRegistryMirror(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	commonCmdData.SetupPlatform(cmd)
	commonCmdData.SetupDebugTemplates(cmd)
	commonCmdData.SetupAllowIncludesUpdate(cmd)

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

	repo, ok := storageManager.GetStagesStorage().(*storage.RepoStagesStorage)
	if !ok {
		return fmt.Errorf("--repo must be a container registry address")
	}

	metaRepoAddress, found, err := repo.GetMetaRepoMarker(ctx, projectName)
	if err != nil {
		return fmt.Errorf("unable to read meta-repo marker: %w", err)
	}
	if !found {
		logboek.Context(ctx).Default().LogF("No meta-repo safeguard is set for project %q in %s; nothing to do.\n", projectName, repo.Address())
		return nil
	}

	if err := repo.RmMetaRepoMarker(ctx, projectName); err != nil {
		return fmt.Errorf("unable to remove meta-repo marker: %w", err)
	}

	if metaRepoAddress == "" {
		metaRepoAddress = "unknown meta-repo"
	}
	logboek.Context(ctx).Warn().LogF("Safeguard removed for project %q. Metadata remains in meta-repo %q and was NOT moved back; werf runs without --meta-repo will no longer be blocked and may read stale metadata from %s.\n", projectName, metaRepoAddress, repo.Address())
	return nil
}
