package telemetry

import (
	"context"
	"fmt"
	neturl "net/url"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/trace"
)

var (
	tracerProvider *trace.TracerProvider
	traceExporter  *otlptrace.Exporter
)

func SetupTraceExporter(ctx context.Context, url string) error {
	{
		e, err := NewTraceExporter(url)
		if err != nil {
			return fmt.Errorf("unable to create telemetry trace exporter: %w", err)
		}
		traceExporter = e
	}

	tracerProvider = trace.NewTracerProvider(trace.WithSyncer(traceExporter))

	if err := traceExporter.Start(ctx); err != nil {
		return fmt.Errorf("error starting telemetry trace exporter: %w", err)
	}

	return nil
}

func NewTraceExporter(url string) (*otlptrace.Exporter, error) {
	urlObj, err := neturl.Parse(url)
	if err != nil {
		return nil, fmt.Errorf("bad url: %w", err)
	}

	var opts []otlptracehttp.Option

	if urlObj.Scheme == "http" {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	opts = append(opts,
		otlptracehttp.WithEndpoint(urlObj.Host),
		otlptracehttp.WithURLPath(urlObj.Path),
		otlptracehttp.WithRetry(otlptracehttp.RetryConfig{Enabled: false}),
		otlptracehttp.WithTimeout(5*time.Second),
	)

	client := otlptracehttp.NewClient(opts...)

	return otlptrace.NewUnstarted(client), nil
}
