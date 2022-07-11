package synchronization

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/lockgate/pkg/distributed_locker"
	"github.com/werf/lockgate/pkg/distributed_locker/optimistic_locking_store"
	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/git_repo/gitdata"
	"github.com/werf/werf/pkg/kubeutils"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/storage/synchronization_server"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

var cmdData struct {
	Kubernetes                bool
	KubernetesNamespacePrefix string

	Local                          bool
	LocalLockManagerBaseDir        string
	LocalStagesStorageCacheBaseDir string

	TTL  string
	Host string
	Port string
}

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "synchronization",
		Short:                 "Run synchronization server",
		Long:                  common.GetLongCommandDescription(`Run synchronization server`),
		DisableFlagsInUseLine: true,
		Annotations:           map[string]string{},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error { return runSynchronization(ctx) })
		},
	})

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)

	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeConfigBase64(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)

	cmd.Flags().BoolVarP(&cmdData.Local, "local", "", util.GetBoolEnvironmentDefaultTrue("WERF_LOCAL"), "Use file lock-manager and file stages-storage-cache (true by default or $WERF_LOCAL)")
	cmd.Flags().StringVarP(&cmdData.LocalLockManagerBaseDir, "local-lock-manager-base-dir", "", os.Getenv("WERF_LOCAL_LOCK_MANAGER_BASE_DIR"), "Use specified directory as base for file lock-manager (~/.werf/synchronization_server/lock_manager by default or $WERF_LOCAL_LOCK_MANAGER_BASE_DIR)")
	cmd.Flags().StringVarP(&cmdData.LocalStagesStorageCacheBaseDir, "local-stages-storage-cache-base-dir", "", os.Getenv("WERF_LOCAL_STAGES_STORAGE_CACHE_BASE_DIR"), "Use specified directory as base for file stages-storage-cache (~/.werf/synchronization_server/stages_storage_cache by default or $WERF_LOCAL_STAGES_STORAGE_CACHE_BASE_DIR)")

	cmd.Flags().BoolVarP(&cmdData.Kubernetes, "kubernetes", "", util.GetBoolEnvironmentDefaultFalse("WERF_KUBERNETES"), "Use kubernetes lock-manager stages-storage-cache (default $WERF_KUBERNETES)")
	cmd.Flags().StringVarP(&cmdData.KubernetesNamespacePrefix, "kubernetes-namespace-prefix", "", os.Getenv("WERF_KUBERNETES_NAMESPACE_PREFIX"), "Use specified prefix for namespaces created for lock-manager and stages-storage-cache (defaults to 'werf-synchronization-' when --kubernetes option is used or $WERF_KUBERNETES_NAMESPACE_PREFIX)")

	cmd.Flags().StringVarP(&cmdData.TTL, "ttl", "", os.Getenv("WERF_TTL"), "Time to live for lock-manager locks and stages-storage-cache records (default $WERF_TTL)")
	cmd.Flags().StringVarP(&cmdData.Host, "host", "", os.Getenv("WERF_HOST"), "Bind synchronization server to the specified host (default localhost or $WERF_HOST)")
	cmd.Flags().StringVarP(&cmdData.Port, "port", "", os.Getenv("WERF_PORT"), "Bind synchronization server to the specified port (default 55581 or $WERF_PORT)")

	return cmd
}

func runSynchronization(ctx context.Context) error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %w", err)
	}

	gitDataManager, err := gitdata.GetHostGitDataManager(ctx)
	if err != nil {
		return fmt.Errorf("error getting host git data manager: %w", err)
	}

	if err := git_repo.Init(gitDataManager); err != nil {
		return err
	}

	host, port := cmdData.Host, cmdData.Port
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "55581"
	}

	var distributedLockerBackendFactoryFunc func(clientID string) (distributed_locker.DistributedLockerBackend, error)
	var stagesStorageCacheFactoryFunc func(clientID string) (synchronization_server.StagesStorageCacheInterface, error)

	if cmdData.Kubernetes {
		if err := kube.Init(kube.InitOptions{kube.KubeConfigOptions{
			Context:          *commonCmdData.KubeContext,
			ConfigPath:       *commonCmdData.KubeConfig,
			ConfigDataBase64: *commonCmdData.KubeConfigBase64,
		}}); err != nil {
			return fmt.Errorf("cannot initialize kube: %w", err)
		}

		if err := common.InitKubedog(ctx); err != nil {
			return fmt.Errorf("cannot init kubedog: %w", err)
		}

		distributedLockerBackendFactoryFunc = func(clientID string) (distributed_locker.DistributedLockerBackend, error) {
			namespace := "werf-synchronization"
			configMapName := fmt.Sprintf("werf-%s", clientID)

			if _, err := kubeutils.GetOrCreateConfigMapWithNamespaceIfNotExists(kube.Client, namespace, configMapName); err != nil {
				return nil, fmt.Errorf("unable to create cm/%s in ns/%s: %w", configMapName, namespace, err)
			}

			store := optimistic_locking_store.NewKubernetesResourceAnnotationsStore(
				kube.DynamicClient, schema.GroupVersionResource{
					Group:    "",
					Version:  "v1",
					Resource: "configmaps",
				}, fmt.Sprintf("werf-%s", clientID), "werf-synchronization",
			)
			return distributed_locker.NewOptimisticLockingStorageBasedBackend(store), nil
		}

		stagesStorageCacheFactoryFunc = func(clientID string) (synchronization_server.StagesStorageCacheInterface, error) {
			return storage.NewKubernetesStagesStorageCache("werf-synchronization", kube.Client, func(projectName string) string {
				return fmt.Sprintf("werf-%s", clientID)
			}), nil
		}
	} else {
		stagesStorageCacheBaseDir := cmdData.LocalStagesStorageCacheBaseDir
		if stagesStorageCacheBaseDir == "" {
			stagesStorageCacheBaseDir = filepath.Join(werf.GetHomeDir(), "synchronization_server", "stages_storage_cache")
		}

		distributedLockerBackendFactoryFunc = func(clientID string) (distributed_locker.DistributedLockerBackend, error) {
			store := optimistic_locking_store.NewInMemoryStore()
			return distributed_locker.NewOptimisticLockingStorageBasedBackend(store), nil
		}

		stagesStorageCacheFactoryFunc = func(clientID string) (synchronization_server.StagesStorageCacheInterface, error) {
			return storage.NewFileStagesStorageCache(filepath.Join(stagesStorageCacheBaseDir, clientID)), nil
		}
	}

	return synchronization_server.RunSynchronizationServer(ctx, host, port, distributedLockerBackendFactoryFunc, stagesStorageCacheFactoryFunc)
}
