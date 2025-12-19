package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"

	helm_v3 "github.com/werf/3p-helm/cmd/helm"
	"github.com/werf/common-go/pkg/graceful"
	"github.com/werf/nelm/pkg/action"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/cmd/werf/root"
	"github.com/werf/werf/v2/pkg/background"
	"github.com/werf/werf/v2/pkg/logging"
	"github.com/werf/werf/v2/pkg/process_exterminator"
)

func main() {
	// IMPORTANT. In background mode we MUST take host-lock to prevent parallel "werf host cleanup" processes.
	// The processes write data to the same log files that causes "data race".
	// We don't need to release the lock manually, because it does automatically when the background process will end.
	if background.IsBackgroundModeEnabled() && !background.TryLock() {
		return
	}

	terminationCtx := graceful.WithTermination(context.Background())
	defer graceful.Shutdown(terminationCtx, onShutdown)

	ctx := logging.WithLogger(terminationCtx)

	spew.Config.DisablePointerAddresses = true
	spew.Config.DisableCapacities = true

	root.PrintStackTraces()

	shouldTerminate, err := common.ContainerBackendProcessStartupHook()
	if err != nil {
		graceful.Terminate(ctx, err, 1)
		return
	} else if shouldTerminate {
		return
	}

	if err := process_exterminator.Init(); err != nil {
		graceful.Terminate(ctx, fmt.Errorf("process exterminator initialization failed: %w", err), 1)
		return
	}

	rootCmd, err := root.ConstructRootCmd(ctx)
	if err != nil {
		graceful.Terminate(ctx, err, 1)
		return
	}

	root.SetupTelemetryInit(rootCmd)

	// WARNING this behavior could be changed
	// https://github.com/spf13/cobra/pull/2167 is not accepted yet
	cobra.EnableErrorOnUnknownSubcommand = true

	// Do early exit if termination is started
	if graceful.IsTerminating(ctx) {
		return
	}

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		if helm_v3.IsPluginError(err) {
			common.ShutdownTelemetry(ctx, helm_v3.PluginErrorCode(err))
			graceful.Terminate(ctx, err, helm_v3.PluginErrorCode(err))
			return
		} else if errors.Is(err, action.ErrChangesPlanned) {
			common.ShutdownTelemetry(ctx, 2)
			graceful.Terminate(ctx, action.ErrChangesPlanned, 2)
			return
		} else if errors.Is(err, action.ErrResourceChangesPlanned) {
			common.ShutdownTelemetry(ctx, 2)
			graceful.Terminate(ctx, action.ErrResourceChangesPlanned, 2)
			return
		} else if errors.Is(err, action.ErrReleaseInstallPlanned) {
			common.ShutdownTelemetry(ctx, 3)
			graceful.Terminate(ctx, action.ErrReleaseInstallPlanned, 3)
			return
		} else {
			common.ShutdownTelemetry(ctx, 1)
			graceful.Terminate(ctx, err, 1)
			return
		}
	}

	common.ShutdownTelemetry(ctx, 0)
}

func onShutdown(_ context.Context, desc graceful.TerminationDescriptor) {
	if desc.Signal() != nil {
		logging.Default(fmt.Sprintf("Signal: %s", desc.Signal()))
		os.Exit(desc.ExitCode())
	} else if desc.Err() != nil {
		logging.Error(desc.Err().Error())
		os.Exit(desc.ExitCode())
	}
}
