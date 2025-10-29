package common

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/werf/werf/v2/pkg/logging"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/storage/synchronization/lock_manager"
	"github.com/werf/werf/v2/pkg/storage/synchronization/server"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
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

func checkSynchronizationKubernetesParamsForWarnings(ctx context.Context, cmdData *CmdData) {
	if *cmdData.Synchronization != "" {
		return
	}

	ctx = logging.WithLogger(ctx)

	doPrintWarning := false
	kubeConfigEnv := os.Getenv("KUBECONFIG")
	switch {
	case cmdData.KubeConfigBase64 != "":
		doPrintWarning = true
		global_warnings.GlobalWarningLn(ctx, fmt.Sprintf(`###
##  Required --synchronization param (or WERF_SYNCHRONIZATION env var) to be specified explicitly,
##  because --kube-config-base64=%s (or WERF_KUBE_CONFIG_BASE64, or WERF_KUBECONFIG_BASE64, or $KUBECONFIG_BASE64 env var) has been specified explicitly.`, cmdData.KubeConfigBase64))
	case kubeConfigEnv != "":
		doPrintWarning = true
		global_warnings.GlobalWarningLn(ctx, fmt.Sprintf(`###
##  Required --synchronization param (or WERF_SYNCHRONIZATION env var) to be specified explicitly,
##  because KUBECONFIG=%s env var has been specified explicitly.`, kubeConfigEnv))
	case cmdData.LegacyKubeConfigPath != "":
		doPrintWarning = true
		global_warnings.GlobalWarningLn(ctx, fmt.Sprintf(`###
##  Required --synchronization param (or WERF_SYNCHRONIZATION env var) to be specified explicitly,
##  because --kube-config=%s (or WERF_KUBE_CONFIG, or WERF_KUBECONFIG, or KUBECONFIG env var) has been specified explicitly.`, cmdData.LegacyKubeConfigPath))
	case cmdData.KubeContextCurrent != "":
		doPrintWarning = true
		global_warnings.GlobalWarningLn(ctx, fmt.Sprintf(`###
##  Required --synchronization param (or WERF_SYNCHRONIZATION env var) to be specified explicitly,
##  because --kube-context=%s (or WERF_KUBE_CONTEXT env var) has been specified explicitly.`, cmdData.KubeContextCurrent))
	}

	if doPrintWarning {
		global_warnings.GlobalWarningLn(ctx, `##
##  IMPORTANT: all invocations of the werf for any single project should use the same
##  --synchronization param (or WERF_SYNCHRONIZATION env var) value
##  to prevent inconsistency of the werf setup for this project.
##
##  Format of the synchronization param: kubernetes://NAMESPACE[:CONTEXT][@(base64:BASE64_CONFIG_DATA)|CONFIG_PATH]
##
##  By default werf stores synchronization data using --synchronization=kubernetes://werf-synchronization namespace
##  with default kube-config and kube-context.
##
##  For example, configure werf synchronization with the following settings:
##
##      export WERF_SYNCHRONIZATION=kubernetes://werf-synchronization:mycontext@/root/.kube/custom-config
##
##  — these same settings required to be used in every werf invocation for your project.
###`)
	}
}

// GetSynchronization determines the type of synchronization server
func GetSynchronization(ctx context.Context, cmdData *CmdData, projectName string, stagesStorage storage.StagesStorage) (Synchronization, error) {
	params := lock_manager.SynchronizationParams{
		ProjectName:   projectName,
		ServerAddress: *cmdData.Synchronization,
		StagesStorage: stagesStorage,
	}

	if params.ServerAddress == "" {
		return initDefault(ctx, params)
	} else if protocolIsLocal(params.ServerAddress) {
		return lock_manager.NewLocalSynchronization(ctx, params)
	} else if protocolIsKube(params.ServerAddress) {
		checkSynchronizationKubernetesParamsForWarnings(ctx, cmdData)
		return initKube(ctx, params)
	} else if protocolIsHttpOrHttps(params.ServerAddress) {
		return lock_manager.NewHttpSynchronization(ctx, params)
	} else {
		return nil, fmt.Errorf("only --synchronization=%s or --synchronization=kubernetes://NAMESPACE or --synchronization=http[s]://HOST:PORT/CLIENT_ID is supported, got %q", storage.LocalStorageAddress, *cmdData.Synchronization)
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
	if params.StagesStorage.Address() == storage.LocalStorageAddress {
		return lock_manager.NewLocalSynchronization(ctx, params)
	}
	params.ServerAddress = server.DefaultAddress
	return lock_manager.NewHttpSynchronization(ctx, params)
}

func initKube(ctx context.Context, params lock_manager.SynchronizationParams) (Synchronization, error) {
	ondemandKubeInitializer := GetOndemandKubeInitializer()
	params.CommonKubeInitializer = &lock_manager.SynchronizationKubeParams{
		KubeContext:             ondemandKubeInitializer.KubeContext,
		KubeConfig:              ondemandKubeInitializer.KubeConfig,
		KubeConfigBase64:        ondemandKubeInitializer.KubeConfigBase64,
		KubeConfigPathMergeList: ondemandKubeInitializer.KubeConfigPathMergeList,
	}
	return lock_manager.NewKubernetesSynchronization(ctx, params)
}
