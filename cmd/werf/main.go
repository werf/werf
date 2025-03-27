package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	helm_v3 "github.com/werf/3p-helm/cmd/helm"
	"github.com/werf/logboek"
	"github.com/werf/nelm/pkg/action"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/cmd/werf/root"
	"github.com/werf/werf/v2/pkg/background"
	"github.com/werf/werf/v2/pkg/graceful"
	"github.com/werf/werf/v2/pkg/process_exterminator"
)

func main() {
	// IMPORTANT. In background mode we MUST take host-lock to prevent parallel "werf host cleanup" processes.
	// The processes write data to the same log files that causes "data race".
	// We don't need to release the lock manually, because it does automatically when the background process will end.
	if background.IsBackgroundModeEnabled() && !background.TryLock() {
		return
	}

	defer graceful.Shutdown()

	ctx := common.GetContextWithLogger()

	root.PrintStackTraces()

	shouldTerminate, err := common.ContainerBackendProcessStartupHook()
	if err != nil {
		graceful.Terminate(err.Error(), 1)
	}
	if shouldTerminate {
		return
	}

	common.EnableTerminationSignalsTrap()
	log.SetOutput(logboek.OutStream())
	logrus.StandardLogger().SetOutput(logboek.OutStream())

	if err := process_exterminator.Init(); err != nil {
		graceful.Terminate(fmt.Sprintf("process exterminator initialization failed: %s", err), 1)
	}

	rootCmd, err := root.ConstructRootCmd(ctx)
	if err != nil {
		graceful.Terminate(err.Error(), 1)
	}

	root.SetupTelemetryInit(rootCmd)

	// WARNING this behaviour could be changed
	// https://github.com/spf13/cobra/pull/2167 is not accepted yet
	cobra.EnableErrorOnUnknownSubcommand = true

	if err := rootCmd.Execute(); err != nil {
		if helm_v3.IsPluginError(err) {
			common.ShutdownTelemetry(ctx, helm_v3.PluginErrorCode(err))
			graceful.Terminate(err.Error(), helm_v3.PluginErrorCode(err))
		} else if errors.Is(err, action.ErrChangesPlanned) {
			common.ShutdownTelemetry(ctx, 2)
			graceful.Terminate(fmt.Sprintf("nelm: %v", action.ErrChangesPlanned), 2)
		} else {
			common.ShutdownTelemetry(ctx, 1)
			graceful.Terminate(err.Error(), 1)
		}
	}

	common.ShutdownTelemetry(ctx, 0)
}
