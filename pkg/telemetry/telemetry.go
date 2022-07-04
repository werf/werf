package telemetry

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/util"
)

const (
	TracesURL = "http://localhost:4318/v1/traces"
)

var telemetrywerfio *TelemetryWerfIO

func GetTelemetryWerfIO() TelemetryWerfIOInterface {
	if telemetrywerfio == nil {
		return &NoTelemetryWerfIO{}
	}
	return telemetrywerfio
}

func Init(ctx context.Context) error {
	if !IsEnabled() {
		return nil
	}

	if t, err := NewTelemetryWerfIO(TracesURL); err != nil {
		return fmt.Errorf("unable to setup telemetry.werf.io exporter: %w", err)
	} else {
		telemetrywerfio = t
	}

	return nil
}

func Shutdown(ctx context.Context) error {
	if !IsEnabled() {
		return nil
	}
	return telemetrywerfio.Shutdown(ctx)
}

func IsEnabled() bool {
	return util.GetBoolEnvironmentDefaultFalse("WERF_TELEMETRY")
}
