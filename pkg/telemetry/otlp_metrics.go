package telemetry

import (
	"context"
	"fmt"
	neturl "net/url"
	"runtime"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// Metrics push exporter (OTLP). NOP-safe when disabled.

// MetricsHandle is an opaque handle for a single InitMetrics/MetricsEnd/
// MetricsShutdown lifecycle. Passing it explicitly (instead of a package-level
// global) keeps the pipeline testable.
type MetricsHandle struct {
	mu       sync.Mutex
	enabled  bool
	provider *sdkmetric.MeterProvider
	runs     metric.Int64Counter
	dur      metric.Float64Histogram
	start    time.Time
	command  string
}

// InitMetrics initializes the OTLP metrics pipeline when enabled and returns a
// handle for the rest of the lifecycle. When disabled, it returns a No-Op
// handle so callers can unconditionally call MetricsStart/End/Shutdown.
// endpoint example: http://collector:4318/v1/metrics or https://collector/v1/metrics
func InitMetrics(ctx context.Context, enabled bool, endpoint string) (*MetricsHandle, error) {
	h := &MetricsHandle{}
	if !enabled {
		return h, nil
	}
	if endpoint == "" {
		return h, fmt.Errorf("otlp metrics enabled but endpoint is not set")
	}

	urlObj, err := neturl.Parse(endpoint)
	if err != nil {
		return h, fmt.Errorf("parse otlp endpoint: %w", err)
	}
	switch urlObj.Scheme {
	case "http", "https":
	default:
		return h, fmt.Errorf("parse otlp endpoint: unsupported scheme %q (must be http or https)", urlObj.Scheme)
	}
	if urlObj.Host == "" {
		return h, fmt.Errorf("parse otlp endpoint: missing host in %q", endpoint)
	}

	var opts []otlpmetrichttp.Option
	if urlObj.Scheme == "http" {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}
	opts = append(opts,
		otlpmetrichttp.WithEndpoint(urlObj.Host),
		otlpmetrichttp.WithURLPath(urlObj.Path),
		otlpmetrichttp.WithTimeout(5*time.Second),
	)

	exporter, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		return h, fmt.Errorf("create otlp metric exporter: %w", err)
	}

	// No WithInterval: this is a short-lived CLI process, so the only
	// guaranteed delivery point is the forced flush inside provider.Shutdown.
	reader := sdkmetric.NewPeriodicReader(exporter)
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))

	meter := provider.Meter("github.com/werf/werf")
	runs, err := meter.Int64Counter("werf_runs")
	if err != nil {
		return h, fmt.Errorf("create werf_runs counter: %w", err)
	}
	dur, err := meter.Float64Histogram("werf_run_duration_seconds")
	if err != nil {
		return h, fmt.Errorf("create werf_run_duration_seconds histogram: %w", err)
	}

	otel.SetMeterProvider(provider)
	h.provider = provider
	h.runs = runs
	h.dur = dur
	h.enabled = true
	return h, nil
}

// MetricsStart marks the beginning of a command execution. Safe to call on a
// nil handle (no-op), so callers don't need to guard InitMetrics failures.
func (h *MetricsHandle) MetricsStart(_ context.Context, command string) {
	if h == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.enabled {
		return
	}
	h.command = command
	h.start = time.Now()
}

// MetricsEnd records final counters/histograms with exit code and duration.
// Safe to call on a nil handle (no-op).
func (h *MetricsHandle) MetricsEnd(ctx context.Context, exitCode int) {
	if h == nil {
		return
	}
	h.mu.Lock()
	enabled := h.enabled
	runs := h.runs
	dur := h.dur
	command := h.command
	start := h.start
	h.mu.Unlock()

	if !enabled {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("command", command),
		attribute.String("os", runtime.GOOS),
		attribute.String("arch", runtime.GOARCH),
		attribute.Int("exit_code", exitCode),
	}

	runs.Add(ctx, 1, metric.WithAttributes(attrs...))
	if !start.IsZero() {
		dur.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(attrs...))
	}
}

// MetricsShutdown flushes and shuts down the provider. Safe to call on a nil
// handle (no-op).
func (h *MetricsHandle) MetricsShutdown(ctx context.Context) error {
	if h == nil {
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.enabled || h.provider == nil {
		return nil
	}
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return h.provider.Shutdown(cctx)
}
