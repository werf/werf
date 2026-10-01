package synchronization

import (
	"context"
	"fmt"
	"os"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/lockgate/pkg/distributed_locker"
	"github.com/werf/lockgate/pkg/distributed_locker/optimistic_locking_store"
	"github.com/werf/logboek"
	"github.com/werf/nelm/v2/pkg/kube"
	"github.com/werf/werf/v3/cmd/werf/common"
	"github.com/werf/werf/v3/pkg/deploy"
	"github.com/werf/werf/v3/pkg/storage/synchronization/server"
	"github.com/werf/werf/v3/pkg/tmp_manager"
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
		Long:                  common.GetLongCommandDescription(`Run synchronization server. By default locks are stored in process memory and are lost on restart. Use --kubernetes to store locks in ConfigMaps in the werf-synchronization namespace.`),
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

	cmd.Flags().BoolVarP(&cmdData.Local, "local", "", util.GetBoolEnvironmentDefaultTrue("WERF_LOCAL"), "Compatibility flag; ignored. Without --kubernetes, locks are stored in process memory (default $WERF_LOCAL or true)")
	cmd.Flags().StringVarP(&cmdData.LocalLockManagerBaseDir, "local-lock-manager-base-dir", "", os.Getenv("WERF_LOCAL_LOCK_MANAGER_BASE_DIR"), "Compatibility flag; ignored. No lock files are written (default $WERF_LOCAL_LOCK_MANAGER_BASE_DIR)")
	cmd.Flags().StringVarP(&cmdData.LocalStagesStorageCacheBaseDir, "local-stages-storage-cache-base-dir", "", os.Getenv("WERF_LOCAL_STAGES_STORAGE_CACHE_BASE_DIR"), "Compatibility flag; ignored. No stages-storage cache directory is used (default $WERF_LOCAL_STAGES_STORAGE_CACHE_BASE_DIR)")

	cmd.Flags().BoolVarP(&cmdData.Kubernetes, "kubernetes", "", util.GetBoolEnvironmentDefaultFalse("WERF_KUBERNETES"), "Store locks in Kubernetes ConfigMaps in the werf-synchronization namespace (default $WERF_KUBERNETES)")
	cmd.Flags().StringVarP(&cmdData.KubernetesNamespacePrefix, "kubernetes-namespace-prefix", "", os.Getenv("WERF_KUBERNETES_NAMESPACE_PREFIX"), "Compatibility flag; ignored. The Kubernetes namespace is always werf-synchronization (default $WERF_KUBERNETES_NAMESPACE_PREFIX)")

	cmd.Flags().StringVarP(&cmdData.TTL, "ttl", "", os.Getenv("WERF_TTL"), "Compatibility flag; ignored. Does not configure lock leases (default $WERF_TTL)")
	cmd.Flags().StringVarP(&cmdData.Host, "host", "", os.Getenv("WERF_HOST"), "Bind synchronization server to the specified host (default localhost or $WERF_HOST)")
	cmd.Flags().StringVarP(&cmdData.Port, "port", "", os.Getenv("WERF_PORT"), "Bind synchronization server to the specified port (default 55581 or $WERF_PORT)")

	lo.Must0(common.SetupKubeConnectionFlags(&commonCmdData, cmd))

	return cmd
}

func runSynchronization(ctx context.Context) error {
	_, ctx, err := common.InitCommonComponents(ctx, common.InitCommonComponentsOptions{
		Cmd:                &commonCmdData,
		InitWerf:           true,
		InitGitDataManager: true,
	})
	if err != nil {
		return fmt.Errorf("component init error: %w", err)
	}

	defer func() {
		if err := tmp_manager.DelegateCleanup(ctx); err != nil {
			logboek.Context(ctx).Warn().LogF("Temporary files cleanup preparation failed: %s\n", err)
		}
	}()

	host, port := cmdData.Host, cmdData.Port
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "55581"
	}

	var distributedLockerBackendFactoryFunc func(clientID string) (distributed_locker.DistributedLockerBackend, error)

	if cmdData.Kubernetes {
		namespace := "werf-synchronization"

		kubeConfig, err := kube.NewKubeConfig(ctx, kube.KubeConfigOptions{
			KubeConnectionOptions: commonCmdData.KubeConnectionOptions,
			KubeContextNamespace:  namespace,
		})
		if err != nil {
			return fmt.Errorf("construct kube config: %w", err)
		}

		clientFactory, err := kube.NewClientFactory(ctx, kubeConfig)
		if err != nil {
			return fmt.Errorf("construct kube client factory: %w", err)
		}

		distributedLockerBackendFactoryFunc = func(clientID string) (distributed_locker.DistributedLockerBackend, error) {
			configMapName := fmt.Sprintf("werf-%s", clientID)

			if _, err := deploy.GetOrCreateConfigMapWithNamespaceIfNotExists(ctx, clientFactory.Static(), namespace, configMapName, true); err != nil {
				return nil, fmt.Errorf("unable to create cm/%s in ns/%s: %w", configMapName, namespace, err)
			}

			store := optimistic_locking_store.NewKubernetesResourceAnnotationsStore(
				clientFactory.Dynamic(), schema.GroupVersionResource{
					Group:    "",
					Version:  "v1",
					Resource: "configmaps",
				}, fmt.Sprintf("werf-%s", clientID), "werf-synchronization",
			)
			return distributed_locker.NewOptimisticLockingStorageBasedBackend(store), nil
		}
	} else {
		distributedLockerBackendFactoryFunc = func(clientID string) (distributed_locker.DistributedLockerBackend, error) {
			store := optimistic_locking_store.NewInMemoryStore()
			return distributed_locker.NewOptimisticLockingStorageBasedBackend(store), nil
		}
	}

	return server.Run(ctx, host, port, distributedLockerBackendFactoryFunc)
}
