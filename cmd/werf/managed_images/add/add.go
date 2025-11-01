package add

import (
	"context"
	"fmt"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/level"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/tmp_manager"
	"github.com/werf/werf/v2/pkg/true_git"
)

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "add",
		DisableFlagsInUseLine: true,
		Short:                 "Add image record to the list of managed images which will be preserved during cleanup procedure.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			if err := common.ValidateArgumentCount(1, args, cmd); err != nil {
				return err
			}

			return run(ctx, args[0])
		},
	})

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigRenderDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupGiterminismConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})
	common.SetupSSHKey(&commonCmdData, cmd)

	common.SetupSecondaryStagesStorageOptions(&commonCmdData, cmd)
	common.SetupCacheStagesStorageOptions(&commonCmdData, cmd)
	common.SetupRepoOptions(&commonCmdData, cmd, common.RepoDataOptions{})
	common.SetupFinalRepo(&commonCmdData, cmd)

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read and write images to the specified repo")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.StubSetupInsecureHelmDependencies(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	common.SetupContainerRegistryMirror(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupSynchronization(&commonCmdData, cmd)

	commonCmdData.SetupPlatform(cmd)
	commonCmdData.SetupDebugTemplates(cmd)
	commonCmdData.SetupAllowIncludesUpdate(cmd)

	lo.Must0(common.SetupMinimalKubeConnectionFlags(&commonCmdData, cmd))

	return cmd
}

func run(ctx context.Context, imageName string) error {
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
		SetupOndemandKubeInitializer: true,
	})
	if err != nil {
		return fmt.Errorf("component init error: %w", err)
	}

	containerBackend := commonManager.ContainerBackend()

	if logboek.Context(ctx).IsAcceptedLevel(level.Default) {
		logboek.Context(ctx).SetAcceptedLevel(level.Error)
	}

	projectTmpDir, err := tmp_manager.CreateProjectDir(ctx)
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %w", err)
	}
	defer tmp_manager.ReleaseProjectDir(projectTmpDir)

	giterminismManager, err := common.GetGiterminismManager(ctx, &commonCmdData)
	if err != nil {
		return err
	}

	_, werfConfig, err := common.GetOptionalWerfConfig(ctx, &commonCmdData, giterminismManager, common.GetWerfConfigOptions(&commonCmdData, false))
	if err != nil {
		return fmt.Errorf("unable to load werf config: %w", err)
	}

	var projectName string
	if werfConfig != nil {
		projectName = werfConfig.Meta.Project
	} else {
		return fmt.Errorf("run command in the project directory with werf.yaml")
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

	if err := storageManager.StagesStorage.AddManagedImage(ctx, projectName, common.GetManagedImageName(imageName)); err != nil {
		return fmt.Errorf("unable to add managed image %q for project %q: %w", imageName, projectName, err)
	}

	return nil
}
