package telemetry

import (
	"context"
	"fmt"
	neturl "net/url"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
)

var (
	ctrl     *controller.Controller
	exporter *otlpmetric.Exporter
)

func SetupController(ctx context.Context, url string) error {
	{
		e, err := NewExporter(url)
		if err != nil {
			return fmt.Errorf("unable to create telemetry exporter: %w", err)
		}
		exporter = e
	}
	ctrl = NewController(exporter)

	if err := exporter.Start(ctx); err != nil {
		return fmt.Errorf("error starting telemetry exporter: %w", err)
	}

	return nil
}

func NewExporter(url string) (*otlpmetric.Exporter, error) {
	urlObj, err := neturl.Parse(url)
	if err != nil {
		return nil, fmt.Errorf("bad url: %w", err)
	}

	var opts []otlpmetrichttp.Option

	if urlObj.Scheme == "http" {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}

	opts = append(opts,
		otlpmetrichttp.WithEndpoint(urlObj.Host),
		otlpmetrichttp.WithURLPath(urlObj.Path),
		otlpmetrichttp.WithRetry(otlpmetrichttp.RetryConfig{Enabled: false}),
		otlpmetrichttp.WithTimeout(5*time.Second),
	)

	client := otlpmetrichttp.NewClient(opts...)

	return otlpmetric.NewUnstarted(client), nil
}

func NewController(exporter *otlpmetric.Exporter) *controller.Controller {
	checkpointerFactory := processor.NewFactory(simple.NewWithHistogramDistribution(), exporter)

	return controller.New(
		checkpointerFactory,
		controller.WithExporter(exporter),
		controller.WithCollectTimeout(5*time.Second),
		controller.WithCollectPeriod(0),
		controller.WithPushTimeout(5*time.Second),
	)
}
