package common

import (
	"context"
	"fmt"
	"os"

	"github.com/werf/werf/pkg/telemetry"
	"github.com/werf/werf/pkg/util"
)

func InitTelemetry(ctx context.Context) {
	if err := telemetry.Init(ctx, telemetry.TelemetryOptions{
		ErrorHandlerFunc: func(err error) {
			logTelemetryError(ctx, err.Error())
		},
	}); err != nil {
		logTelemetryError(ctx, fmt.Sprintf("unable to init: %s", err))
	}
}

func ShutdownTelemetry(ctx context.Context, exitCode int) {
	if err := telemetry.Shutdown(ctx); err != nil {
		logTelemetryError(ctx, fmt.Sprintf("unable to shutdown: %s", err))
	}
}

func logTelemetryError(ctx context.Context, msg string) {
	if !util.GetBoolEnvironmentDefaultFalse("WERF_TELEMETRY_LOGS") {
		return
	}
	fmt.Fprintf(os.Stderr, "Telemetry error: %s\n", msg)
}
