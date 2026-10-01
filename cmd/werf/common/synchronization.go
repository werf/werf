package common

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"
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

	if params.ServerAddress != "" {
		logboek.Context(ctx).LogF("Using sync server: %s\n", synchronizationAddressForLog(params.ServerAddress))
	}

	if params.ServerAddress == "" {
		return initDefault(ctx, params)
	} else if protocolIsLocal(params.ServerAddress) {
		return lock_manager.NewLocalSynchronization(ctx, params)
	} else if protocolIsKube(params.ServerAddress) {
		return lock_manager.NewKubernetesSynchronization(ctx, params)
	} else if protocolIsHttpOrHttps(params.ServerAddress) {
		return lock_manager.NewHttpSynchronization(ctx, params)
	} else {
		return nil, fmt.Errorf("unsupported synchronization address; use :local, kubernetes://NAMESPACE[:CONTEXT][@CONFIG], or http[s]://HOST:PORT")
	}
}

func synchronizationAddressForLog(address string) string {
	if protocolIsKube(address) {
		if prefix, _, found := strings.Cut(address, "@base64:"); found {
			return prefix + "@base64:[REDACTED]"
		}
		return address
	}
	if protocolIsLocal(address) {
		return address
	}
	parsed, err := url.Parse(address)
	if err != nil || !protocolIsHttpOrHttps(address) {
		return "[invalid address]"
	}
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.ForceQuery = false
	parsed.Fragment = ""
	return parsed.String()
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
	return lock_manager.NewHttpSynchronization(ctx, params)
}
