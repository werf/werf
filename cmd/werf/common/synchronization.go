package common

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/spf13/cobra"

	"github.com/werf/werf/v3/pkg/storage"
	"github.com/werf/werf/v3/pkg/storage/synchronization/lock_manager"
	"github.com/werf/werf/v3/pkg/storage/synchronization/server"
)

const (
	syncProtocolKube  = "kubernetes://"
	syncProtocolHttp  = "http://"
	syncProtocolHttps = "https://"
)

var (
	_ Synchronization = (*lock_manager.LocalSynchronization)(nil)
	_ Synchronization = (*lock_manager.HttpSynchronization)(nil)
	_ Synchronization = (*lock_manager.KubernetesSynchronization)(nil)
)

type Synchronization interface {
	GetStorageLockManager(ctx context.Context) (lock_manager.Interface, error)
}

func SetupSynchronization(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Synchronization = new(string)

	defaultValue := os.Getenv("WERF_SYNCHRONIZATION")

	cmd.Flags().StringVarP(cmdData.Synchronization, "synchronization", "S", defaultValue, fmt.Sprintf(`Address of synchronizer for multiple werf processes to work with a single repo.

Default:
 - $WERF_SYNCHRONIZATION, or
 - :local if --repo is not specified, or
 - %s if --repo has been specified.

Without an explicit value, an unavailable public server (DNS, network, timeout, or HTTP 5xx) causes a warning and the build continues without publication locks. An explicit value, including the public server address, is strict. The first publication with an explicit value records a repository marker requiring the same value for subsequent commands.

The same address should be specified for all werf processes that work with a single repo. :local address allows execution of werf processes from a single host only`, server.DefaultAddress))
}

func GetSynchronization(ctx context.Context, cmdData *CmdData, projectName string, stagesStorage storage.StagesStorage) (Synchronization, error) {
	params := lock_manager.SynchronizationParams{
		ProjectName:           projectName,
		ServerAddress:         *cmdData.Synchronization,
		StagesStorage:         stagesStorage,
		KubeConnectionOptions: cmdData.KubeConnectionOptions,
	}
	if params.ServerAddress != "" && !protocolIsLocal(params.ServerAddress) && protocolIsLocal(params.StagesStorage.Address()) {
		return nil, fmt.Errorf("--synchronization (or WERF_SYNCHRONIZATION) is set but --repo (or WERF_REPO) is not specified: --repo is required when using a non-local synchronization server")
	}

	var markerStore *storage.RepoStagesStorage
	if !protocolIsLocal(stagesStorage.Address()) {
		var ok bool
		markerStore, ok = stagesStorage.(*storage.RepoStagesStorage)
		if !ok {
			return nil, fmt.Errorf("synchronization marker requires repository storage")
		}
		if err := validateSynchronizationMarker(ctx, markerStore, projectName, params.ServerAddress); err != nil {
			return nil, err
		}
	}

	var synchronization Synchronization
	var err error
	switch {
	case params.ServerAddress == "":
		synchronization, err = initDefault(ctx, params)
	case protocolIsLocal(params.ServerAddress):
		synchronization, err = lock_manager.NewLocalSynchronization(ctx, params)
	case protocolIsKube(params.ServerAddress):
		synchronization, err = lock_manager.NewKubernetesSynchronization(ctx, params)
	case protocolIsHttpOrHttps(params.ServerAddress):
		synchronization, err = lock_manager.NewHttpSynchronization(ctx, params)
	default:
		return nil, fmt.Errorf("unsupported synchronization address; use :local, kubernetes://NAMESPACE[:CONTEXT][@CONFIG], or http[s]://HOST:PORT")
	}
	if err != nil {
		return nil, err
	}
	if markerStore == nil {
		return synchronization, nil
	}
	return &markerSynchronization{Synchronization: synchronization, store: markerStore, projectName: projectName, address: params.ServerAddress}, nil
}

func protocolIsKube(address string) bool {
	return strings.HasPrefix(address, syncProtocolKube)
}

func protocolIsHttpOrHttps(address string) bool {
	return strings.HasPrefix(address, syncProtocolHttp) || strings.HasPrefix(address, syncProtocolHttps)
}

func protocolIsLocal(address string) bool {
	return address == storage.LocalStorageAddress
}

func initDefault(ctx context.Context, params lock_manager.SynchronizationParams) (Synchronization, error) {
	if params.StagesStorage.Address() == storage.LocalStorageAddress {
		return lock_manager.NewLocalSynchronization(ctx, params)
	}
	params.ServerAddress = server.DefaultAddress
	params.AllowFallback = true
	return lock_manager.NewHttpSynchronization(ctx, params)
}

func synchronizationFingerprint(address string) string {
	fingerprint := sha256.Sum256([]byte(address))
	return hex.EncodeToString(fingerprint[:])
}

func validateSynchronizationMarker(ctx context.Context, store *storage.RepoStagesStorage, projectName, address string) error {
	fingerprint, found, err := store.GetSynchronizationMarker(ctx, projectName)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	if address == "" {
		return fmt.Errorf("project %q requires explicitly configured synchronization; pass --synchronization or WERF_SYNCHRONIZATION matching repository marker %q", projectName, store.SynchronizationMarkerReference(ctx, projectName))
	}
	if fingerprint != synchronizationFingerprint(address) {
		return fmt.Errorf("synchronization configuration differs from repository marker %q for project %q", store.SynchronizationMarkerReference(ctx, projectName), projectName)
	}
	return nil
}

type markerSynchronization struct {
	Synchronization
	store       *storage.RepoStagesStorage
	projectName string
	address     string
}

var _ Synchronization = (*markerSynchronization)(nil)

func (s *markerSynchronization) GetStorageLockManager(ctx context.Context) (lock_manager.Interface, error) {
	manager, err := s.Synchronization.GetStorageLockManager(ctx)
	if err != nil {
		return nil, err
	}
	return &markerLockManager{Interface: manager, store: s.store, projectName: s.projectName, address: s.address}, nil
}

type markerLockManager struct {
	lock_manager.Interface
	once        sync.Once
	markerErr   error
	store       *storage.RepoStagesStorage
	projectName string
	address     string
}

var _ lock_manager.Interface = (*markerLockManager)(nil)

func (manager *markerLockManager) LockStage(ctx context.Context, projectName, digest string) (lock_manager.LockHandle, error) {
	handle, err := manager.Interface.LockStage(ctx, projectName, digest)
	if err != nil {
		return lock_manager.LockHandle{}, err
	}
	manager.once.Do(func() {
		if manager.address == "" {
			manager.markerErr = validateSynchronizationMarker(ctx, manager.store, manager.projectName, manager.address)
		} else {
			manager.markerErr = manager.store.PutSynchronizationMarker(ctx, manager.projectName, synchronizationFingerprint(manager.address))
		}
	})
	if manager.markerErr != nil {
		releaseErr := manager.Interface.Unlock(ctx, handle)
		return lock_manager.LockHandle{}, errors.Join(manager.markerErr, releaseErr)
	}
	return handle, nil
}
