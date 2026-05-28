package lock_manager

import (
	"context"
	"fmt"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/werf"
)

type SynchronizationParams struct {
	ProjectName   string
	ServerAddress string
	MetaStorage   storage.MetaStorage
}

type LocalSynchronization struct {
	address string
}

// NewLocalSynchronization returns local synchronization parameters
func NewLocalSynchronization(ctx context.Context, params SynchronizationParams) (*LocalSynchronization, error) {
	serverAddress := storage.LocalStorageAddress
	if ForceSyncServerRepo == "true" {
		repoSyncServer, err := checkRepoSyncServer(ctx, params.ProjectName, serverAddress, params.MetaStorage)
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
	return NewGeneric(werf.HostLocker().Locker()), nil
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
				repoSyncServer, err := checkRepoSyncServer(ctx, params.ProjectName, serverAddress, params.MetaStorage)
				if err != nil {
					return err
				}
				serverAddress = repoSyncServer
			}

			clientID, err = GetHttpClientID(ctx, params.ProjectName, serverAddress, params.MetaStorage)
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

func checkRepoSyncServer(ctx context.Context, projectName, serverAddress string, metaStorage storage.MetaStorage) (string, error) {
	logboek.Info().LogProcess("Checking synchronization server")
	repoSyncServer, err := GetOrCreateSyncServer(ctx, projectName, serverAddress, metaStorage)
	if err != nil {
		return "", fmt.Errorf("unable to get synchronization server address: %w", err)
	}

	if repoSyncServer != serverAddress {
		err := PromptRewriteSyncRepoServer(ctx, serverAddress, repoSyncServer)
		if err != nil {
			return "", err
		}
		err = OverwriteSyncServerRepo(ctx, projectName, serverAddress, metaStorage)
		if err != nil {
			return "", fmt.Errorf("unable to overwrite synchronization server: %w", err)
		}
		repoSyncServer = serverAddress
	}

	return repoSyncServer, nil
}
