package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"

	"github.com/werf/werf/pkg/util"
)

const (
	TracesURL = "https://telemetry.werf.io/v1/traces"
)

var telemetrywerfio *TelemetryWerfIO

func GetTelemetryWerfIO() TelemetryWerfIOInterface {
	if telemetrywerfio == nil {
		return &NoTelemetryWerfIO{}
	}
	return telemetrywerfio
}

type TelemetryOptions struct {
	ErrorHandlerFunc func(err error)
}

func Init(ctx context.Context, opts TelemetryOptions) error {
	if !IsEnabled() {
		return nil
	}

	if t, err := NewTelemetryWerfIO(TracesURL, TelemetryWerfIOOptions{
		HandleErrorFunc: opts.ErrorHandlerFunc,
	}); err != nil {
		return fmt.Errorf("unable to setup telemetry.werf.io exporter: %w", err)
	} else {
		telemetrywerfio = t
	}

	otel.SetErrorHandler(&callFuncErrorHandler{f: opts.ErrorHandlerFunc})

	if err := telemetrywerfio.Start(ctx); err != nil {
		return fmt.Errorf("unable to start telemetry.werf.io exporter: %w", err)
	}

	return nil
}

type callFuncErrorHandler struct{ f func(error) }

func (h *callFuncErrorHandler) Handle(err error) {
	if h.f != nil {
		h.f(err)
	}
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

func IsLogsEnabled() bool {
	return util.GetBoolEnvironmentDefaultFalse("WERF_TELEMETRY_LOGS")
}
