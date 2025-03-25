package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	helm_v3 "github.com/werf/3p-helm/cmd/helm"
	"github.com/werf/nelm/pkg/resrcchangcalc"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/cmd/werf/root"
	"github.com/werf/werf/v2/pkg/logging"
	"github.com/werf/werf/v2/pkg/process_exterminator"
)

func main() {
	ctx, closeOutput := logging.WithLogger(context.Background())
	defer closeOutput()

	root.PrintStackTraces()

	shouldTerminate, err := common.ContainerBackendProcessStartupHook()
	if err != nil {
		common.TerminateWithError(err.Error(), 1)
	}
	if shouldTerminate {
		return
	}

	common.EnableTerminationSignalsTrap()

	if err := process_exterminator.Init(); err != nil {
		common.TerminateWithError(fmt.Sprintf("process exterminator initialization failed: %s", err), 1)
	}

	rootCmd, err := root.ConstructRootCmd(ctx)
	if err != nil {
		common.TerminateWithError(err.Error(), 1)
	}

	root.SetupTelemetryInit(rootCmd)

	// WARNING this behaviour could be changed
	// https://github.com/spf13/cobra/pull/2167 is not accepted yet
	cobra.EnableErrorOnUnknownSubcommand = true

	if err := rootCmd.Execute(); err != nil {
		if helm_v3.IsPluginError(err) {
			common.ShutdownTelemetry(ctx, helm_v3.PluginErrorCode(err))
			common.TerminateWithError(err.Error(), helm_v3.PluginErrorCode(err))
		} else if errors.Is(err, resrcchangcalc.ErrChangesPlanned) {
			common.ShutdownTelemetry(ctx, 2)
			common.TerminateWithError(err.Error(), 2)
		} else {
			common.ShutdownTelemetry(ctx, 1)
			common.TerminateWithError(err.Error(), 1)
		}
	}

	common.ShutdownTelemetry(ctx, 0)
}
