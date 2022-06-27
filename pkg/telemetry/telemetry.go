package telemetry

import (
	"context"
	"fmt"
	"runtime"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

const (
	URL = "http://localhost:4318/v1/metrics"
)

var metrics *Metrics

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

	metrics = &Metrics{
		Version: werf.Version,
		Command: "TODO",
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
	}

	if err := SetupController(ctx, URL); err != nil {
		return fmt.Errorf("unable to setup telemetry controller for url %q: %w", URL, err)
	}

	return nil
}

func Shutdown(ctx context.Context) error {
	if !IsEnabled() {
		return nil
	}

	if err := metrics.WriteOpenTelemetry(ctx, ctrl.Meter("werf")); err != nil {
		return fmt.Errorf("unable to write open telemetry data: %w", err)
	}

	if err := ctrl.Collect(ctx); err != nil {
		return fmt.Errorf("otel controller collection failed: %w", err)
	}

	if err := exporter.Export(ctx, ctrl.Resource(), ctrl); err != nil {
		return fmt.Errorf("unable to export telemetry: %w", err)
	}

	logboek.Context(ctx).Debug().LogF("Telemetry collection done\n")

	return nil
}

func IsEnabled() bool {
	return util.GetBoolEnvironmentDefaultFalse("WERF_TELEMETRY")
}
