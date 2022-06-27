package common

import (
	"context"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/telemetry"
)

func InitTelemetry(ctx context.Context) {
	// TODO: append error to the ~/.werf/telemetry_errors.log
	err := telemetry.Init(ctx)
	logboek.Context(ctx).Debug().LogF("Telemetry: init error: %s\n", err)
}

func ShutdownTelemetry(ctx context.Context, exitCode int) {
	telemetry.ChangeMetrics(func(metrics *telemetry.Metrics) {
		metrics.ExitCode = exitCode
	})

	// TODO: append error to the ~/.werf/telemetry_errors.log
	err := telemetry.Shutdown(ctx)

	logboek.Context(ctx).Debug().LogF("Telemetry: shutdown error: %s\n", err)
}
