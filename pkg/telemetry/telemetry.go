package telemetry

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/google/uuid"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

const (
	MetricsURL = "http://localhost:4318/v1/metrics"
	TracesURL  = "http://localhost:4318/v1/traces"
)

var (
	executionID string
	projectID   string
	command     string
	metrics     *Metrics
)

func ChangeMetrics(changeFunc func(metrics *Metrics)) {
	if !IsEnabled() {
		return
	}
	changeFunc(metrics)
}

func Init(ctx context.Context) error {
	if !IsEnabled() {
		return nil
	}

	executionID = uuid.New().String()
	projectID = "26a38bba-51ed-4f93-8104-e6f51a5454ef"
	command = fmt.Sprintf("%v", os.Args)

	metrics = &Metrics{
		Version: werf.Version,
		Command: "TODO",
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
	}

	if err := SetupTraceExporter(ctx, TracesURL); err != nil {
		return fmt.Errorf("unable to setup telemetry traces exporter for url %q: %w", TracesURL, err)
	}

	return nil
}

func Shutdown(ctx context.Context) error {
	if !IsEnabled() {
		return nil
	}

	if err := traceExporter.Shutdown(ctx); err != nil {
		return fmt.Errorf("unable to shutdown trace exporter: %w", err)
	}

	if err := tracerProvider.Shutdown(ctx); err != nil {
		return fmt.Errorf("unable to shutdown trace provider: %w", err)
	}

	logboek.Context(ctx).Debug().LogF("Telemetry collection done\n")

	return nil
}

func IsEnabled() bool {
	return util.GetBoolEnvironmentDefaultFalse("WERF_TELEMETRY")
}
