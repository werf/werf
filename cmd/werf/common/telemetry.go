package common

import (
	"context"
	"fmt"
	"os"

	"github.com/werf/werf/pkg/telemetry"
)

func InitTelemetry(ctx context.Context) {
	if err := telemetry.Init(ctx, telemetry.TelemetryOptions{
		ErrorHandlerFunc: func(err error) {
			if err == nil {
				return
			}
			logTelemetryError(err.Error())
		},
	}); err != nil {
		logTelemetryError(fmt.Sprintf("unable to init: %s", err))
	}
}

func ShutdownTelemetry(ctx context.Context, exitCode int) {
	if err := telemetry.Shutdown(ctx); err != nil {
		logTelemetryError(fmt.Sprintf("unable to shutdown: %s", err))
	}
}

func logTelemetryError(msg string) {
	if !telemetry.IsLogsEnabled() {
		return
	}
	fmt.Fprintf(os.Stderr, "Telemetry error: %s\n", msg)
}
