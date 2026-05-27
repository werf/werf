package common

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/storage/synchronization/lock_manager"
	"github.com/werf/werf/v2/pkg/storage/synchronization/server"
)

const (
	syncProtocolKube  = "kubernetes://"
	syncProtocolHttp  = "http://"
	syncProtocolHttps = "https://"
)

type Synchronization interface {
	// GetStorageLockManager returns lock manager interface based on synchronization server type
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

The same address should be specified for all werf processes that work with a single repo. :local address allows execution of werf processes from a single host only`, server.DefaultAddress))
}

// GetSynchronization determines the type of synchronization server
func GetSynchronization(ctx context.Context, cmdData *CmdData, projectName string, stagesStorage storage.StagesStorage) (Synchronization, error) {
	params := lock_manager.SynchronizationParams{
		ProjectName:   projectName,
		ServerAddress: *cmdData.Synchronization,
		MetaStorage:   stagesStorage,
	}

	if params.ServerAddress == "" {
		return initDefault(ctx, params)
	} else if protocolIsLocal(params.ServerAddress) {
		return lock_manager.NewLocalSynchronization(ctx, params)
	} else if protocolIsKube(params.ServerAddress) {
		return nil, fmt.Errorf("--synchronization=kubernetes:// no longer supported")
	} else if protocolIsHttpOrHttps(params.ServerAddress) {
		return lock_manager.NewHttpSynchronization(ctx, params)
	} else {
		return nil, fmt.Errorf("only --synchronization=%s or --synchronization=http[s]://HOST:PORT/CLIENT_ID is supported, got %q", storage.LocalStorageAddress, *cmdData.Synchronization)
	}
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
	if params.MetaStorage.Address() == storage.LocalStorageAddress {
		return lock_manager.NewLocalSynchronization(ctx, params)
	}
	params.ServerAddress = server.DefaultAddress
	return lock_manager.NewHttpSynchronization(ctx, params)
}
