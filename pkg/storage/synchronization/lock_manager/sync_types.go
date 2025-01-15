package lock_manager

import (
	"context"
	"fmt"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	chart "github.com/werf/common-go/pkg/lock"
	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/storage"
)

type SynchronizationParams struct {
	ProjectName           string
	ServerAddress         string
	StagesStorage         storage.StagesStorage
	CommonKubeInitializer *SynchronizationKubeParams
}

type SynchronizationKubeParams struct {
	KubeContext             string
	KubeConfig              string
	KubeConfigBase64        string
	KubeConfigPathMergeList []string
}

type LocalSynchronization struct {
	address string
}

// NewLocalSynchronization returns local synchronization parameters
func NewLocalSynchronization(ctx context.Context, params SynchronizationParams) (*LocalSynchronization, error) {
	serverAddress := storage.LocalStorageAddress
	if ForceSyncServerRepo == "true" {
		repoSyncServer, err := checkRepoSyncServer(ctx, params.ProjectName, serverAddress, params.StagesStorage)
		if err != nil {
			return nil, err
		}
		serverAddress = repoSyncServer
	}
	return &LocalSynchronization{
		address: serverAddress,
	}, nil
}

func (s *LocalSynchronization) GetStorageLockManager(_ context.Context) (Interface, error) {
	return NewGeneric(chart.HostLocker), nil
}

type HttpSynchronization struct {
	address  string
	clientId string
}

// NewHttpSynchronization returns http/https synchronization parameters
func NewHttpSynchronization(ctx context.Context, params SynchronizationParams) (*HttpSynchronization, error) {
	var clientID string
	var err error
	serverAddress := params.ServerAddress
	if err := logboek.Info().LogProcess("Getting client id for the http synchronization server").
		DoError(func() error {
			if ForceSyncServerRepo == "true" {
				repoSyncServer, err := checkRepoSyncServer(ctx, params.ProjectName, serverAddress, params.StagesStorage)
				if err != nil {
					return err
				}
				serverAddress = repoSyncServer
			}

			clientID, err = GetHttpClientID(ctx, params.ProjectName, serverAddress, params.StagesStorage)
			if err != nil {
				return fmt.Errorf("unable to get client id for the http synchronization server: %w", err)
			}

			logboek.Info().LogF("Using clientID %q for http synchronization server at address %s\n", clientID, serverAddress)

			return err
		}); err != nil {
		return nil, fmt.Errorf("unable to init http synchronization: %w", err)
	}

	return &HttpSynchronization{
		address:  serverAddress,
		clientId: clientID,
	}, nil
}

func (s *HttpSynchronization) GetStorageLockManager(ctx context.Context) (Interface, error) {
	return NewHttp(ctx, s.address, s.clientId)
}

type KubernetesSynchronization struct {
	address    string
	kubeParams *KubernetesParams
}

// NewKubernetesSynchronization returns kubernetes synchronization parameters
func NewKubernetesSynchronization(ctx context.Context, params SynchronizationParams) (*KubernetesSynchronization, error) {
	serverAddress := params.ServerAddress
	if ForceSyncServerRepo == "true" {
		repoSyncServer, err := checkRepoSyncServer(ctx, params.ProjectName, serverAddress, params.StagesStorage)
		if err != nil {
			return nil, err
		}
		serverAddress = repoSyncServer
	}
	kubeParams, err := ParseKubernetesParams(serverAddress)
	if err != nil {
		return nil, fmt.Errorf("unable to parse synchronization address %s: %w", serverAddress, err)
	}

	if kubeParams.ConfigPath == "" {
		kubeParams.ConfigPath = params.CommonKubeInitializer.KubeConfig
	}
	if kubeParams.ConfigContext == "" {
		kubeParams.ConfigContext = params.CommonKubeInitializer.KubeContext
	}
	if kubeParams.ConfigDataBase64 == "" {
		kubeParams.ConfigDataBase64 = params.CommonKubeInitializer.KubeConfigBase64
	}
	if kubeParams.ConfigPathMergeList == nil {
		kubeParams.ConfigPathMergeList = params.CommonKubeInitializer.KubeConfigPathMergeList
	}

	return &KubernetesSynchronization{
		address:    serverAddress,
		kubeParams: kubeParams,
	}, nil
}

func (s *KubernetesSynchronization) GetStorageLockManager(_ context.Context) (Interface, error) {
	if config, err := kube.GetKubeConfig(kube.KubeConfigOptions{
		ConfigPath:          s.kubeParams.ConfigPath,
		ConfigDataBase64:    s.kubeParams.ConfigDataBase64,
		ConfigPathMergeList: s.kubeParams.ConfigPathMergeList,
		Context:             s.kubeParams.ConfigContext,
	}); err != nil {
		return nil, fmt.Errorf("unable to load synchronization kube config %q (context %q): %w", s.kubeParams.ConfigPath, s.kubeParams.ConfigContext, err)
	} else if dynamicClient, err := dynamic.NewForConfig(config.Config); err != nil {
		return nil, fmt.Errorf("unable to create synchronization kubernetes dynamic client: %w", err)
	} else if client, err := kubernetes.NewForConfig(config.Config); err != nil {
		return nil, fmt.Errorf("unable to create synchronization kubernetes client: %w", err)
	} else {
		return NewKubernetes(s.kubeParams.Namespace, client, dynamicClient, func(projectName string) string {
			return fmt.Sprintf("werf-%s", projectName)
		}), nil
	}
}

func checkRepoSyncServer(ctx context.Context, projectName, serverAddress string, stagesStorage storage.StagesStorage) (string, error) {
	logboek.Info().LogProcess("Checking synchronization server")
	repoSyncServer, err := GetOrCreateSyncServer(ctx, projectName, serverAddress, stagesStorage)
	if err != nil {
		return "", fmt.Errorf("unable to get synchronization server address: %w", err)
	}

	if repoSyncServer != serverAddress {
		err := PromptRewriteSyncRepoServer(ctx, serverAddress, repoSyncServer)
		if err != nil {
			return "", err
		}
		err = OverwriteSyncServerRepo(ctx, projectName, serverAddress, stagesStorage)
		if err != nil {
			return "", fmt.Errorf("unable to overwrite synchronization server: %w", err)
		}
		repoSyncServer = serverAddress
	}

	return repoSyncServer, nil
}
