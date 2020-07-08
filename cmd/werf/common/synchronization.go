package common

import (
	"fmt"
	"os"
	"strings"

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

func SetupSynchronizationKubeConfig(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.SynchronizationKubeConfig = new(string)

	defaultValue := os.Getenv("WERF_SYNCHRONIZATION_KUBE_CONFIG")

	cmd.Flags().StringVarP(cmdData.SynchronizationKubeConfig, "synchronization-kube-config", "", defaultValue, "Kubernetes config to use for synchronization. This option should be explicitly set when --kube-config has been explicitly specified.")
}

func SetupSynchronizationKubeContext(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.SynchronizationKubeContext = new(string)

	defaultValue := os.Getenv("WERF_SYNCHRONIZATION_KUBE_CONTEXT")

	cmd.Flags().StringVarP(cmdData.SynchronizationKubeContext, "synchronization-kube-context", "", defaultValue, "Kubernetes context to use for synchronization.")
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

	KubeConfig    string
	KubeContext   string
	KubeNamespace string
}

func checkSynchronizationKubernetesParamsForWarnings(cmdData *CmdData) {
	var doPrintWarn bool

	specifiedKubeConfig := *cmdData.KubeConfig
	if kubeConfigEnv := os.Getenv("KUBECONFIG"); kubeConfigEnv != "" {
		specifiedKubeConfig = kubeConfigEnv
	}

	if specifiedKubeConfig != "" && *cmdData.SynchronizationKubeConfig == "" {
		if !doPrintWarn {
			werf.GlobalWarningLn(`##########################################################################################################################`)
		}
		doPrintWarn = true

		werf.GlobalWarningLn(`##  Required --synchronization-kube-config (or WERF_SYNCHRONIZATION_KUBE_CONFIG env var) param to be specified explicitly,`)
		werf.GlobalWarningLn(fmt.Sprintf(`##  because --kube-config (or WERF_KUBE_CONFIG or KUBECONFIG env var) = %s has been specified explicitly.`, *cmdData.KubeConfig))
	}

	if *cmdData.KubeContext != "" && *cmdData.SynchronizationKubeContext == "" {
		if !doPrintWarn {
			werf.GlobalWarningLn(`##########################################################################################################################`)
		} else {
			werf.GlobalWarningLn(`##  `)
		}
		doPrintWarn = true

		werf.GlobalWarningLn(`##  Required --synchronization-kube-context (or WERF_SYNCHRONIZATION_KUBE_CONTEXT env var) param to be specified explicitly,`)
		werf.GlobalWarningLn(fmt.Sprintf(`##  because --kube-context (or WERF_KUBE_CONTEXT env var) = %s has been specified explicitly.`, *cmdData.KubeContext))
	}

	if doPrintWarn {
		werf.GlobalWarningLn(`##  `)
		werf.GlobalWarningLn(`##  IMPORTANT: all invocations of the werf for any single project should use the same`)
		werf.GlobalWarningLn(`##  --synchronization-kube-config (or WERF_SYNCHRONIZATION_KUBE_CONFIG env var) and`)
		werf.GlobalWarningLn(`##  --synchronization-kube-context (or WERF_SYNCHRONIZATION_KUBE_CONTEXT env var) params values`)
		werf.GlobalWarningLn(`##  to prevent inconsistency of the werf setup for this project.`)
		werf.GlobalWarningLn(`##  `)
		werf.GlobalWarningLn(`##  By default werf stores synchronization data using --synchronization=kubernetes://werf-synchronization namespace`)
		werf.GlobalWarningLn(`##  with default kube-config and kube-context.`)
		werf.GlobalWarningLn(`##  `)
		werf.GlobalWarningLn(`##  For example, configure werf synchronization with the following settings:`)
		werf.GlobalWarningLn(`##  `)
		werf.GlobalWarningLn(`##      export WERF_SYNCHRONIZATION_KUBE_CONFIG=~/.kube/config`)
		werf.GlobalWarningLn(`##      export WERF_SYNCHRONIZATION_KUBE_CONTEXT=werf-synchronization`)
		werf.GlobalWarningLn(`##  `)
		werf.GlobalWarningLn(`##  More info about synchronization: https://werf.io/documentation/reference/stages_and_images.html#synchronization-locks-and-stages-storage-cache`)
		werf.GlobalWarningLn(`##########################################################################################################################`)
	}
}

func GetSynchronization(cmdData *CmdData, stagesStorageAddress string) (*SynchronizationParams, error) {
	synchronizationKubeConfig := *cmdData.SynchronizationKubeConfig
	if synchronizationKubeConfig == "" {
		synchronizationKubeConfig = *cmdData.KubeConfig
	}

	synchronizationKubeContext := *cmdData.SynchronizationKubeContext
	if synchronizationKubeContext == "" {
		synchronizationKubeContext = *cmdData.KubeContext
	}

	if *cmdData.Synchronization == "" {
		if stagesStorageAddress == storage.LocalStorageAddress {
			return &SynchronizationParams{Address: storage.LocalStorageAddress, SynchronizationType: LocalSynchronization}, nil
		} else {
			checkSynchronizationKubernetesParamsForWarnings(cmdData)

			return &SynchronizationParams{
				Address:             storage.DefaultKubernetesStorageAddress,
				SynchronizationType: KubernetesSynchronization,
				KubeConfig:          synchronizationKubeConfig,
				KubeContext:         synchronizationKubeContext,
				KubeNamespace:       strings.TrimPrefix(storage.DefaultKubernetesStorageAddress, "kubernetes://"),
			}, nil
		}
	} else {
		if *cmdData.Synchronization == storage.LocalStorageAddress {
			return &SynchronizationParams{Address: *cmdData.Synchronization, SynchronizationType: LocalSynchronization}, nil
		} else if strings.HasPrefix(*cmdData.Synchronization, "kubernetes://") {
			checkSynchronizationKubernetesParamsForWarnings(cmdData)

			return &SynchronizationParams{
				Address:             *cmdData.Synchronization,
				SynchronizationType: KubernetesSynchronization,
				KubeConfig:          synchronizationKubeConfig,
				KubeContext:         synchronizationKubeContext,
				KubeNamespace:       strings.TrimPrefix(*cmdData.Synchronization, "kubernetes://"),
			}, nil
		} else if strings.HasPrefix(*cmdData.Synchronization, "http://") || strings.HasPrefix(*cmdData.Synchronization, "https://") {
			return &SynchronizationParams{Address: *cmdData.Synchronization, SynchronizationType: HttpSynchronization}, nil
		} else {
			return nil, fmt.Errorf("only --synchronization=%s or --synchronization=kubernetes://NAMESPACE or --synchronization=http[s]://HOST:PORT/CLIENT_ID is supported, got %q", storage.LocalStorageAddress, *cmdData.Synchronization)
		}
	}
}

func GetStagesStorageCache(synchronization *SynchronizationParams) (storage.StagesStorageCache, error) {
	switch synchronization.SynchronizationType {
	case LocalSynchronization:
		return storage.NewFileStagesStorageCache(werf.GetStagesStorageCacheDir()), nil
	case KubernetesSynchronization:
		if config, err := kube.GetKubeConfig(kube.KubeConfigOptions{KubeConfig: synchronization.KubeConfig, KubeContext: synchronization.KubeContext}); err != nil {
			return nil, fmt.Errorf("unable to load synchronization kube config %q (context %q)", synchronization.KubeConfig, synchronization.KubeContext)
		} else if client, err := kubernetes.NewForConfig(config.Config); err != nil {
			return nil, fmt.Errorf("unable to create synchronization kubernetes client: %s", err)
		} else {
			return storage.NewKubernetesStagesStorageCache(synchronization.KubeNamespace, client), nil
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
		if config, err := kube.GetKubeConfig(kube.KubeConfigOptions{KubeConfig: synchronization.KubeConfig, KubeContext: synchronization.KubeContext}); err != nil {
			return nil, fmt.Errorf("unable to load synchronization kube config %q (context %q)", synchronization.KubeConfig, synchronization.KubeContext)
		} else if dynamicClient, err := dynamic.NewForConfig(config.Config); err != nil {
			return nil, fmt.Errorf("unable to create synchronization kubernetes dynamic client: %s", err)
		} else if client, err := kubernetes.NewForConfig(config.Config); err != nil {
			return nil, fmt.Errorf("unable to create synchronization kubernetes client: %s", err)
		} else {
			return storage.NewKubernetesLockManager(synchronization.KubeNamespace, client, dynamicClient), nil
		}
	case HttpSynchronization:
		return synchronization_server.NewLockManagerHttpClient(fmt.Sprintf("%s/lock-manager", synchronization.Address)), nil
	default:
		panic(fmt.Sprintf("unsupported synchronization address %q", synchronization.Address))
	}
}
