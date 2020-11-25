package common

import (
	"fmt"
	"os"
	"strings"

	"github.com/werf/werf/pkg/werf/locker_with_retry"

	"github.com/werf/lockgate/pkg/distributed_locker"
	"github.com/werf/logboek"

	"github.com/spf13/cobra"

	"k8s.io/client-go/dynamic"

	"k8s.io/client-go/kubernetes"

	"github.com/werf/kubedog/pkg/kube"

	"github.com/werf/werf/pkg/storage/synchronization_server"

	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/werf"
)

func SetupSynchronization(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Synchronization = new(string)

	defaultValue := os.Getenv("WERF_SYNCHRONIZATION")

	cmd.Flags().StringVarP(cmdData.Synchronization, "synchronization", "S", defaultValue, fmt.Sprintf("Address of synchronizer for multiple werf processes to work with a single stages storage (default :local if --stages-storage=:local or %s if non-local stages-storage specified or $WERF_SYNCHRONIZATION if set). The same address should be specified for all werf processes that work with a single stages storage. :local address allows execution of werf processes from a single host only.", storage.DefaultKubernetesStorageAddress))
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

	doPrintWarning := false
	if *cmdData.KubeConfigBase64 != "" {
		doPrintWarning = true
		werf.GlobalWarningLn(`###`)
		werf.GlobalWarningLn(`##  Required --synchronization param (or WERF_SYNCHRONIZATION env var) to be specified explicitly,`)
		werf.GlobalWarningLn(fmt.Sprintf(`##  because --kube-config-base64=%s (or WERF_KUBE_CONFIG_BASE64, or WERF_KUBECONFIG_BASE64, or $KUBECONFIG_BASE64 env var) has been specified explicitly.`, *cmdData.KubeConfigBase64))
	} else if kubeConfigEnv := os.Getenv("KUBECONFIG"); kubeConfigEnv != "" {
		doPrintWarning = true
		werf.GlobalWarningLn(`###`)
		werf.GlobalWarningLn(`##  Required --synchronization param (or WERF_SYNCHRONIZATION env var) to be specified explicitly,`)
		werf.GlobalWarningLn(fmt.Sprintf(`##  because KUBECONFIG=%s env var has been specified explicitly.`, kubeConfigEnv))
	} else if *cmdData.KubeConfig != "" {
		doPrintWarning = true
		werf.GlobalWarningLn(`###`)
		werf.GlobalWarningLn(`##  Required --synchronization param (or WERF_SYNCHRONIZATION env var) to be specified explicitly,`)
		werf.GlobalWarningLn(fmt.Sprintf(`##  because --kube-config=%s (or WERF_KUBE_CONFIG, or WERF_KUBECONFIG, or KUBECONFIG env var) has been specified explicitly.`, kubeConfigEnv))
	} else if *cmdData.KubeContext != "" {
		doPrintWarning = true
		werf.GlobalWarningLn(`###`)
		werf.GlobalWarningLn(`##  Required --synchronization param (or WERF_SYNCHRONIZATION env var) to be specified explicitly,`)
		werf.GlobalWarningLn(fmt.Sprintf(`##  because --kube-context=%s (or WERF_KUBE_CONTEXT env var) has been specified explicitly.`, kubeConfigEnv))
	}

	if doPrintWarning {
		werf.GlobalWarningLn(`##  `)
		werf.GlobalWarningLn(`##  IMPORTANT: all invocations of the werf for any single project should use the same`)
		werf.GlobalWarningLn(`##  --synchronization param (or WERF_SYNCHRONIZATION env var) value`)
		werf.GlobalWarningLn(`##  to prevent inconsistency of the werf setup for this project.`)
		werf.GlobalWarningLn(`##  `)
		werf.GlobalWarningLn(`##  Format of the synchronization param: kubernetes://NAMESPACE[:CONTEXT][@(base64:BASE64_CONFIG_DATA)|CONFIG_PATH]`)
		werf.GlobalWarningLn(`##  `)
		werf.GlobalWarningLn(`##  By default werf stores synchronization data using --synchronization=kubernetes://werf-synchronization namespace`)
		werf.GlobalWarningLn(`##  with default kube-config and kube-context.`)
		werf.GlobalWarningLn(`##  `)
		werf.GlobalWarningLn(`##  For example, configure werf synchronization with the following settings:`)
		werf.GlobalWarningLn(`##  `)
		werf.GlobalWarningLn(`##      export WERF_SYNCHRONIZATION=kubernetes://werf-synchronization:mycontext@/root/.kube/custom-config`)
		werf.GlobalWarningLn(`##  `)
		werf.GlobalWarningLn(`##  — these same settings required to be used in every werf invocation for your project.`)
		werf.GlobalWarningLn(`###`)
	}
}

func GetSynchronization(cmdData *CmdData, projectName string, stagesStorage storage.StagesStorage) (*SynchronizationParams, error) {
	var defaultKubernetesSynchronization string
	if *cmdData.Synchronization == "" {
		defaultKubernetesSynchronization = storage.DefaultKubernetesStorageAddress

		if *cmdData.KubeContext != "" {
			defaultKubernetesSynchronization += fmt.Sprintf(":%s", *cmdData.KubeContext)
		}

		if *cmdData.KubeConfigBase64 != "" {
			defaultKubernetesSynchronization += fmt.Sprintf("@base64:%s", *cmdData.KubeConfigBase64)
		} else if *cmdData.KubeConfig != "" {
			defaultKubernetesSynchronization += fmt.Sprintf("@%s", *cmdData.KubeConfig)
		}
	}

	getKubeParamsFunc := func(address string) (*SynchronizationParams, error) {
		res := &SynchronizationParams{}
		res.SynchronizationType = KubernetesSynchronization
		res.Address = address

		if params, err := storage.ParseKubernetesSynchronization(res.Address); err != nil {
			return nil, fmt.Errorf("unable to parse synchronization address %s: %s", res.Address, err)
		} else {
			res.KubeParams = params
			return res, nil
		}
	}

	getHttpParamsFunc := func(synchronization string, stagesStorage storage.StagesStorage) (*SynchronizationParams, error) {
		var address string
		if err := logboek.Default.LogProcess(fmt.Sprintf("Getting client id for the http syncrhonization server"), logboek.LevelLogProcessOptions{}, func() error {
			if clientID, err := synchronization_server.GetOrCreateClientID(projectName, synchronization_server.NewSynchronizationClient(synchronization), stagesStorage); err != nil {
				return fmt.Errorf("unable to get synchronization client id: %s", err)
			} else {
				address = fmt.Sprintf("%s/%s", synchronization, clientID)
				logboek.Default.LogF("Using clientID %q for http synchronization server at address %s\n", clientID, address)
				return nil
			}
		}); err != nil {
			return nil, err
		}

		return &SynchronizationParams{Address: address, SynchronizationType: HttpSynchronization}, nil
	}

	if *cmdData.Synchronization == "" {
		if stagesStorage.Address() == storage.LocalStorageAddress {
			return &SynchronizationParams{SynchronizationType: LocalSynchronization, Address: storage.LocalStorageAddress}, nil
		} else {
			return getHttpParamsFunc("https://synchronization.werf.io", stagesStorage)
		}
	} else if *cmdData.Synchronization == storage.LocalStorageAddress {
		return &SynchronizationParams{Address: *cmdData.Synchronization, SynchronizationType: LocalSynchronization}, nil
	} else if strings.HasPrefix(*cmdData.Synchronization, "kubernetes://") {
		checkSynchronizationKubernetesParamsForWarnings(cmdData)
		return getKubeParamsFunc(*cmdData.Synchronization)
	} else if strings.HasPrefix(*cmdData.Synchronization, "http://") || strings.HasPrefix(*cmdData.Synchronization, "https://") {
		return getHttpParamsFunc(*cmdData.Synchronization, stagesStorage)
	} else {
		return nil, fmt.Errorf("only --synchronization=%s or --synchronization=kubernetes://NAMESPACE or --synchronization=http[s]://HOST:PORT/CLIENT_ID is supported, got %q", storage.LocalStorageAddress, *cmdData.Synchronization)
	}
}

func GetStagesStorageCache(synchronization *SynchronizationParams) (storage.StagesStorageCache, error) {
	switch synchronization.SynchronizationType {
	case LocalSynchronization:
		return storage.NewFileStagesStorageCache(werf.GetStagesStorageCacheDir()), nil
	case KubernetesSynchronization:
		if config, err := kube.GetKubeConfig(kube.KubeConfigOptions{
			ConfigPath:       synchronization.KubeParams.ConfigPath,
			ConfigDataBase64: synchronization.KubeParams.ConfigDataBase64,
			Context:          synchronization.KubeParams.ConfigContext,
		}); err != nil {
			return nil, fmt.Errorf("unable to load synchronization kube config %q (context %q)", synchronization.KubeParams.ConfigPath, synchronization.KubeParams.ConfigContext)
		} else if client, err := kubernetes.NewForConfig(config.Config); err != nil {
			return nil, fmt.Errorf("unable to create synchronization kubernetes client: %s", err)
		} else {
			return storage.NewKubernetesStagesStorageCache(synchronization.KubeParams.Namespace, client, func(projectName string) string {
				return fmt.Sprintf("werf-%s", projectName)
			}), nil
		}
	case HttpSynchronization:
		return synchronization_server.NewStagesStorageCacheHttpClient(fmt.Sprintf("%s/stages-storage-cache", synchronization.Address)), nil
	default:
		panic(fmt.Sprintf("unsupported synchronization address %q", synchronization.Address))
	}
}

func GetStorageLockManager(synchronization *SynchronizationParams) (storage.LockManager, error) {
	switch synchronization.SynchronizationType {
	case LocalSynchronization:
		return storage.NewGenericLockManager(werf.GetHostLocker()), nil
	case KubernetesSynchronization:
		if config, err := kube.GetKubeConfig(kube.KubeConfigOptions{
			ConfigPath:       synchronization.KubeParams.ConfigPath,
			ConfigDataBase64: synchronization.KubeParams.ConfigDataBase64,
			Context:          synchronization.KubeParams.ConfigContext,
		}); err != nil {
			return nil, fmt.Errorf("unable to load synchronization kube config %q (context %q)", synchronization.KubeParams.ConfigPath, synchronization.KubeParams.ConfigContext)
		} else if dynamicClient, err := dynamic.NewForConfig(config.Config); err != nil {
			return nil, fmt.Errorf("unable to create synchronization kubernetes dynamic client: %s", err)
		} else if client, err := kubernetes.NewForConfig(config.Config); err != nil {
			return nil, fmt.Errorf("unable to create synchronization kubernetes client: %s", err)
		} else {
			return storage.NewKubernetesLockManager(synchronization.KubeParams.Namespace, client, dynamicClient, func(projectName string) string {
				return fmt.Sprintf("werf-%s", projectName)
			}), nil
		}
	case HttpSynchronization:
		locker := distributed_locker.NewHttpLocker(fmt.Sprintf("%s/locker", synchronization.Address))
		lockerWithRetry := locker_with_retry.NewLockerWithRetry(locker, locker_with_retry.LockerWithRetryOptions{MaxAcquireAttempts: 10, MaxReleaseAttempts: 10})
		return storage.NewGenericLockManager(lockerWithRetry), nil
	default:
		panic(fmt.Sprintf("unsupported synchronization address %q", synchronization.Address))
	}
}
