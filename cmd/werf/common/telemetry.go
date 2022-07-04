package common

import (
	"context"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/telemetry"
)

func InitTelemetry(ctx context.Context) {
	// TODO: append error to the ~/.werf/telemetry_errors.log
	if err := telemetry.Init(ctx); err != nil {
		logboek.Context(ctx).Debug().LogF("Telemetry: init error: %s\n", err)
	}
}

func ShutdownTelemetry(ctx context.Context, exitCode int) {
	// TODO: append error to the ~/.werf/telemetry_errors.log
	if err := telemetry.Shutdown(ctx); err != nil {
		logboek.Context(ctx).Debug().LogF("Telemetry: shutdown error: %s\n", err)
	}
}
