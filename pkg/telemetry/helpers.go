package telemetry

import (
	"fmt"
	neturl "net/url"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
)

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
		otlptracehttp.WithTimeout(1300*time.Millisecond),
	)

	client := otlptracehttp.NewClient(opts...)

	return otlptrace.NewUnstarted(client), nil
}
