package telemetry

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel"

	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

const (
	TracesURL    = "https://telemetry.werf.io/v1/traces"
	TelemetryEnv = "WERF_TELEMETRY"
)

var (
	telemetrywerfio *TelemetryWerfIO
	logFile         *os.File
)

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

	if path := os.Getenv("WERF_TELEMETRY_LOG_FILE"); path != "" {
		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
		if err != nil {
			return fmt.Errorf("unable to open log file %q: %w", path, err)
		}
		logFile = f
	}

	if t, err := NewTelemetryWerfIO(GetTraceUrl(), TelemetryWerfIOOptions{
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
	if telemetrywerfio == nil {
		return nil
	}

	if logFile != nil {
		defer logFile.Close()
	}

	return telemetrywerfio.Shutdown(ctx)
}

func GetTraceUrl() string {
	if url := os.Getenv("WERF_TELEMETRY_TRACE_URL"); url != "" {
		return url
	}

	return TracesURL
}

func IsEnabled() bool {
	val, isSet := util.LookupBoolEnvironment(TelemetryEnv)
	if isSet && val != nil {
		return *val
	}

	if werf.Version == "dev" || werf.Version == "0.0.0" {
		return false
	}
	return true
}

func LogF(f string, args ...interface{}) {
	if logFile == nil {
		return
	}
	fmt.Fprintf(logFile, "[%d][%s] Telemetry: %s\n", os.Getpid(), time.Now(), fmt.Sprintf(f, args...))
}
