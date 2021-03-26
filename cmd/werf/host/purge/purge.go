package reset

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/cleaning"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/git_repo/gitdata"
	"github.com/werf/werf/pkg/host_cleaning"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/storage/manager"
	"github.com/werf/werf/pkg/werf"
	"github.com/werf/werf/pkg/werf/global_warnings"
)

var cmdData struct {
	Force bool
}

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "purge",
		Short: "Purge werf images, cache and other data for all projects on host machine",
		Long: common.GetLongCommandDescription(`Purge werf images, cache and other data for all projects on host machine.

The data include:
* Old service tmp dirs, which werf creates during every build, converge and other commands.
* Local cache:
  * Remote git clones cache.
  * Git worktree cache.
* Shared context:
  * Mounts which persists between several builds (mounts from build_dir).

WARNING: Do not run this command during any other werf command is working on the host machine. This command is supposed to be run manually.`),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			defer global_warnings.PrintGlobalWarnings(common.BackgroundContext())

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}
			common.LogVersion()

			return common.LogRunningTime(runReset)
		},
	}

	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)
	common.SetupDockerConfig(&commonCmdData, cmd, "")
	common.SetupProjectName(&commonCmdData, cmd)

	common.SetupSynchronization(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupDryRun(&commonCmdData, cmd)
	cmd.Flags().BoolVarP(&cmdData.Force, "force", "", false, common.CleaningCommandsForceOptionDescription)

	return cmd
}

func runReset() error {
	ctx := common.BackgroundContext()

	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	gitDataManager, err := gitdata.GetHostGitDataManager(ctx)
	if err != nil {
		return fmt.Errorf("error getting host git data manager: %s", err)
	}

	if err := git_repo.Init(gitDataManager); err != nil {
		return err
	}

	if err := image.Init(); err != nil {
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

	projectName := *commonCmdData.ProjectName
	if projectName == "" {
		logboek.LogOptionalLn()
		hostPurgeOptions := host_cleaning.HostPurgeOptions{DryRun: *commonCmdData.DryRun, RmContainersThatUseWerfImages: cmdData.Force}
		if err := host_cleaning.HostPurge(ctx, hostPurgeOptions); err != nil {
			return err
		}
	} else {
		containerRuntime := &container_runtime.LocalDockerServerRuntime{} // TODO
		stagesStorage, err := common.GetLocalStagesStorage(containerRuntime)
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

		storageManager := manager.NewStorageManager(projectName, stagesStorage, nil, storageLockManager, stagesStorageCache)
		purgeOptions := cleaning.PurgeOptions{
			RmContainersThatUseWerfImages: cmdData.Force,
			DryRun:                        *commonCmdData.DryRun,
		}

		logboek.LogOptionalLn()
		if err := cleaning.Purge(ctx, projectName, storageManager, storageLockManager, purgeOptions); err != nil {
			return err
		}
	}

	return nil
}
