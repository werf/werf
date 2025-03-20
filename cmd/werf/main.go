package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	helm_v3 "github.com/werf/3p-helm/cmd/helm"
	"github.com/werf/logboek"
	"github.com/werf/nelm/pkg/resrcchangcalc"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/cmd/werf/root"
	"github.com/werf/werf/v2/pkg/process_exterminator"
)

func main() {
	ctx := common.GetContextWithLogger()

	root.PrintStackTraces()

	shouldTerminate, err := common.ContainerBackendProcessStartupHook()
	if err != nil {
		common.TerminateWithError(err.Error(), 1)
	}
	if shouldTerminate {
		return
	}

	common.EnableTerminationSignalsTrap()
	log.SetOutput(logboek.OutStream())
	logrus.StandardLogger().SetOutput(logboek.OutStream())

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
			os.Exit(2)
		} else {
			common.ShutdownTelemetry(ctx, 1)
			common.TerminateWithError(err.Error(), 1)
		}
	}

	common.ShutdownTelemetry(ctx, 0)
}
