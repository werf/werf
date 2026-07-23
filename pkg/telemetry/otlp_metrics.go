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
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// Metrics push exporter (OTLP). NOP-safe when disabled.

var metricsState struct {
	mu       sync.Mutex
	enabled  bool
	provider *sdkmetric.MeterProvider
	meter    metric.Meter
	start    time.Time
	command  string
}

// InitMetrics initializes the OTLP metrics pipeline when enabled.
// endpoint example: http://collector:4318/v1/metrics or https://collector/v1/metrics
func InitMetrics(ctx context.Context, enabled bool, endpoint string) error {
	metricsState.mu.Lock()
	defer metricsState.mu.Unlock()

	if !enabled {
		metricsState.enabled = false
		return nil
	}
	if endpoint == "" {
		return fmt.Errorf("telemetry enabled but --telemetry-otlp-endpoint is empty")
	}

	urlObj, err := neturl.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("bad otlp endpoint: %w", err)
	}

	var opts []otlpmetrichttp.Option
	if urlObj.Scheme == "http" {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}
	opts = append(opts,
		otlpmetrichttp.WithEndpoint(urlObj.Host),
		otlpmetrichttp.WithURLPath(urlObj.Path),
		otlpmetrichttp.WithTimeout(5*time.Second),
		otlpmetrichttp.WithRetry(otlpmetrichttp.RetryConfig{Enabled: false}),
	)

	client := otlpmetrichttp.NewClient(opts...)
	exporter, err := otlpmetric.New(ctx, client)
	if err != nil {
		return fmt.Errorf("create otlp metric exporter: %w", err)
	}

	reader := sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(2*time.Second))
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))

	otel.SetMeterProvider(provider)
	metricsState.provider = provider
	metricsState.meter = otel.Meter("github.com/werf/werf")
	metricsState.enabled = true
	return nil
}

// MetricsStart marks the beginning of a command execution.
func MetricsStart(_ context.Context, command string) {
	metricsState.mu.Lock()
	defer metricsState.mu.Unlock()
	if !metricsState.enabled {
		return
	}
	metricsState.command = command
	metricsState.start = time.Now()
}

// MetricsEnd records final counters/histograms with exit code and duration.
func MetricsEnd(ctx context.Context, exitCode int) {
	metricsState.mu.Lock()
	enabled := metricsState.enabled
	meter := metricsState.meter
	command := metricsState.command
	start := metricsState.start
	metricsState.mu.Unlock()

	if !enabled {
		return
	}

	runs, _ := meter.Int64Counter("werf_runs")
	dur, _ := meter.Float64Histogram("werf_run_duration_seconds")

	attrs := []attribute.KeyValue{
		attribute.String("command", command),
		attribute.String("os", runtime.GOOS),
		attribute.String("arch", runtime.GOARCH),
		attribute.Int("exit_code", exitCode),
	}

	runs.Add(ctx, 1, attrs...)
	if !start.IsZero() {
		dur.Record(ctx, time.Since(start).Seconds(), attrs...)
	}
}

// MetricsShutdown flushes and shuts down the provider.
func MetricsShutdown(ctx context.Context) error {
	metricsState.mu.Lock()
	defer metricsState.mu.Unlock()
	if !metricsState.enabled || metricsState.provider == nil {
		return nil
	}
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return metricsState.provider.Shutdown(cctx)
}
