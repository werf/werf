package common

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/lockgate/pkg/distributed_locker"
	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/storage/synchronization_server"
	"github.com/werf/werf/pkg/werf"
	"github.com/werf/werf/pkg/werf/global_warnings"
	"github.com/werf/werf/pkg/werf/locker_with_retry"
)

func SetupSynchronization(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Synchronization = new(string)

	defaultValue := os.Getenv("WERF_SYNCHRONIZATION")

	cmd.Flags().StringVarP(cmdData.Synchronization, "synchronization", "S", defaultValue, fmt.Sprintf(`Address of synchronizer for multiple werf processes to work with a single repo.

Default:
 - $WERF_SYNCHRONIZATION, or
 - :local if --repo is not specified, or
 - %s if --repo has been specified.

The same address should be specified for all werf processes that work with a single repo. :local address allows execution of werf processes from a single host only`, storage.DefaultHttpSynchronizationServer))
}

type SynchronizationType string

const (
	LocalSynchronization      SynchronizationType = "LocalSynchronization"
	KubernetesSynchronization SynchronizationType = "KubernetesSynchronization"
	HttpSynchronization       SynchronizationType = "HttpSynchronization"
)

type SynchronizationParams struct {
	Address             string
	SynchronizationType SynchronizationType
	KubeParams          *storage.KubernetesSynchronizationParams
}

func checkSynchronizationKubernetesParamsForWarnings(cmdData *CmdData) {
	if *cmdData.Synchronization != "" {
		return
	}

	ctx := GetContextWithLogger()
	doPrintWarning := false
	kubeConfigEnv := os.Getenv("KUBECONFIG")
	switch {
	case *cmdData.KubeConfigBase64 != "":
		doPrintWarning = true
		global_warnings.GlobalWarningLn(ctx, `###`)
		global_warnings.GlobalWarningLn(ctx, `##  Required --synchronization param (or WERF_SYNCHRONIZATION env var) to be specified explicitly,`)
		global_warnings.GlobalWarningLn(ctx, fmt.Sprintf(`##  because --kube-config-base64=%s (or WERF_KUBE_CONFIG_BASE64, or WERF_KUBECONFIG_BASE64, or $KUBECONFIG_BASE64 env var) has been specified explicitly.`, *cmdData.KubeConfigBase64))
	case kubeConfigEnv != "":
		doPrintWarning = true
		global_warnings.GlobalWarningLn(ctx, `###`)
		global_warnings.GlobalWarningLn(ctx, `##  Required --synchronization param (or WERF_SYNCHRONIZATION env var) to be specified explicitly,`)
		global_warnings.GlobalWarningLn(ctx, fmt.Sprintf(`##  because KUBECONFIG=%s env var has been specified explicitly.`, kubeConfigEnv))
	case *cmdData.KubeConfig != "":
		doPrintWarning = true
		global_warnings.GlobalWarningLn(ctx, `###`)
		global_warnings.GlobalWarningLn(ctx, `##  Required --synchronization param (or WERF_SYNCHRONIZATION env var) to be specified explicitly,`)
		global_warnings.GlobalWarningLn(ctx, fmt.Sprintf(`##  because --kube-config=%s (or WERF_KUBE_CONFIG, or WERF_KUBECONFIG, or KUBECONFIG env var) has been specified explicitly.`, *cmdData.KubeConfig))
	case *cmdData.KubeContext != "":
		doPrintWarning = true
		global_warnings.GlobalWarningLn(ctx, `###`)
		global_warnings.GlobalWarningLn(ctx, `##  Required --synchronization param (or WERF_SYNCHRONIZATION env var) to be specified explicitly,`)
		global_warnings.GlobalWarningLn(ctx, fmt.Sprintf(`##  because --kube-context=%s (or WERF_KUBE_CONTEXT env var) has been specified explicitly.`, *cmdData.KubeContext))
	}

	if doPrintWarning {
		global_warnings.GlobalWarningLn(ctx, `##  `)
		global_warnings.GlobalWarningLn(ctx, `##  IMPORTANT: all invocations of the werf for any single project should use the same`)
		global_warnings.GlobalWarningLn(ctx, `##  --synchronization param (or WERF_SYNCHRONIZATION env var) value`)
		global_warnings.GlobalWarningLn(ctx, `##  to prevent inconsistency of the werf setup for this project.`)
		global_warnings.GlobalWarningLn(ctx, `##  `)
		global_warnings.GlobalWarningLn(ctx, `##  Format of the synchronization param: kubernetes://NAMESPACE[:CONTEXT][@(base64:BASE64_CONFIG_DATA)|CONFIG_PATH]`)
		global_warnings.GlobalWarningLn(ctx, `##  `)
		global_warnings.GlobalWarningLn(ctx, `##  By default werf stores synchronization data using --synchronization=kubernetes://werf-synchronization namespace`)
		global_warnings.GlobalWarningLn(ctx, `##  with default kube-config and kube-context.`)
		global_warnings.GlobalWarningLn(ctx, `##  `)
		global_warnings.GlobalWarningLn(ctx, `##  For example, configure werf synchronization with the following settings:`)
		global_warnings.GlobalWarningLn(ctx, `##  `)
		global_warnings.GlobalWarningLn(ctx, `##      export WERF_SYNCHRONIZATION=kubernetes://werf-synchronization:mycontext@/root/.kube/custom-config`)
		global_warnings.GlobalWarningLn(ctx, `##  `)
		global_warnings.GlobalWarningLn(ctx, `##  — these same settings required to be used in every werf invocation for your project.`)
		global_warnings.GlobalWarningLn(ctx, `##  `)
		global_warnings.GlobalWarningLn(ctx, `##  More info about synchronization: https://werf.io/documentation/v1.2/usage/build/process.html`)
		global_warnings.GlobalWarningLn(ctx, `###`)
	}
}

func GetSynchronization(ctx context.Context, cmdData *CmdData, projectName string, stagesStorage storage.StagesStorage) (*SynchronizationParams, error) {
	getKubeParamsFunc := func(address string, commonKubeInitializer *OndemandKubeInitializer) (*SynchronizationParams, error) {
		res := &SynchronizationParams{}
		res.SynchronizationType = KubernetesSynchronization
		res.Address = address

		if params, err := storage.ParseKubernetesSynchronization(res.Address); err != nil {
			return nil, fmt.Errorf("unable to parse synchronization address %s: %w", res.Address, err)
		} else {
			res.KubeParams = params
		}

		if res.KubeParams.ConfigPath == "" {
			res.KubeParams.ConfigPath = commonKubeInitializer.KubeConfig
		}
		if res.KubeParams.ConfigContext == "" {
			res.KubeParams.ConfigContext = commonKubeInitializer.KubeContext
		}
		if res.KubeParams.ConfigDataBase64 == "" {
			res.KubeParams.ConfigDataBase64 = commonKubeInitializer.KubeConfigBase64
		}
		if res.KubeParams.ConfigPathMergeList == nil {
			res.KubeParams.ConfigPathMergeList = commonKubeInitializer.KubeConfigPathMergeList
		}

		return res, nil
	}

	getHttpParamsFunc := func(synchronization string, stagesStorage storage.StagesStorage) (*SynchronizationParams, error) {
		var address string
		if err := logboek.Info().LogProcess(fmt.Sprintf("Getting client id for the http synchronization server")).
			DoError(func() error {
				if clientID, err := synchronization_server.GetOrCreateClientID(ctx, projectName, synchronization_server.NewSynchronizationClient(synchronization), stagesStorage); err != nil {
					return fmt.Errorf("unable to get synchronization client id: %w", err)
				} else {
					address = fmt.Sprintf("%s/%s", synchronization, clientID)
					logboek.Info().LogF("Using clientID %q for http synchronization server at address %s\n", clientID, address)
					return nil
				}
			}); err != nil {
			return nil, err
		}

		return &SynchronizationParams{Address: address, SynchronizationType: HttpSynchronization}, nil
	}

	switch {
	case *cmdData.Synchronization == "":
		if stagesStorage.Address() == storage.LocalStorageAddress {
			return &SynchronizationParams{SynchronizationType: LocalSynchronization, Address: storage.LocalStorageAddress}, nil
		}

		return getHttpParamsFunc(storage.DefaultHttpSynchronizationServer, stagesStorage)
	case *cmdData.Synchronization == storage.LocalStorageAddress:
		return &SynchronizationParams{Address: *cmdData.Synchronization, SynchronizationType: LocalSynchronization}, nil
	case strings.HasPrefix(*cmdData.Synchronization, "kubernetes://"):
		checkSynchronizationKubernetesParamsForWarnings(cmdData)
		return getKubeParamsFunc(*cmdData.Synchronization, GetOndemandKubeInitializer())
	case strings.HasPrefix(*cmdData.Synchronization, "http://") || strings.HasPrefix(*cmdData.Synchronization, "https://"):
		return getHttpParamsFunc(*cmdData.Synchronization, stagesStorage)
	default:
		return nil, fmt.Errorf("only --synchronization=%s or --synchronization=kubernetes://NAMESPACE or --synchronization=http[s]://HOST:PORT/CLIENT_ID is supported, got %q", storage.LocalStorageAddress, *cmdData.Synchronization)
	}
}

func GetStorageLockManager(ctx context.Context, synchronization *SynchronizationParams) (storage.LockManager, error) {
	switch synchronization.SynchronizationType {
	case LocalSynchronization:
		return storage.NewGenericLockManager(werf.GetHostLocker()), nil
	case KubernetesSynchronization:
		if config, err := kube.GetKubeConfig(kube.KubeConfigOptions{
			ConfigPath:          synchronization.KubeParams.ConfigPath,
			ConfigDataBase64:    synchronization.KubeParams.ConfigDataBase64,
			ConfigPathMergeList: synchronization.KubeParams.ConfigPathMergeList,
			Context:             synchronization.KubeParams.ConfigContext,
		}); err != nil {
			return nil, fmt.Errorf("unable to load synchronization kube config %q (context %q): %w", synchronization.KubeParams.ConfigPath, synchronization.KubeParams.ConfigContext, err)
		} else if dynamicClient, err := dynamic.NewForConfig(config.Config); err != nil {
			return nil, fmt.Errorf("unable to create synchronization kubernetes dynamic client: %w", err)
		} else if client, err := kubernetes.NewForConfig(config.Config); err != nil {
			return nil, fmt.Errorf("unable to create synchronization kubernetes client: %w", err)
		} else {
			return storage.NewKubernetesLockManager(synchronization.KubeParams.Namespace, client, dynamicClient, func(projectName string) string {
				return fmt.Sprintf("werf-%s", projectName)
			}), nil
		}
	case HttpSynchronization:
		locker := distributed_locker.NewHttpLocker(fmt.Sprintf("%s/locker", synchronization.Address))
		lockerWithRetry := locker_with_retry.NewLockerWithRetry(ctx, locker, locker_with_retry.LockerWithRetryOptions{MaxAcquireAttempts: 10, MaxReleaseAttempts: 10})
		return storage.NewGenericLockManager(lockerWithRetry), nil
	default:
		panic(fmt.Sprintf("unsupported synchronization address %q", synchronization.Address))
	}
}
