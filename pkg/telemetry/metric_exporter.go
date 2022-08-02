package telemetry

/* Metric exporter initialization example

var (
	ctrl           *controller.Controller
	metricExporter *otlpmetric.Exporter
)

func SetupMetricExporter(ctx context.Context, url string) error {
	{
		e, err := NewMetricExporter(url)
		if err != nil {
			return fmt.Errorf("unable to create telemetry metric exporter: %w", err)
		}
		metricExporter = e
	}
	ctrl = NewController(metricExporter)

	if err := metricExporter.Start(ctx); err != nil {
		return fmt.Errorf("error starting telemetry metric exporter: %w", err)
	}

	return nil
}

func NewMetricExporter(url string) (*otlpmetric.Exporter, error) {
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
*/

/* Metric exporter usage example

type Metrics struct {
	Version  string
	Command  string
	OS       string
	Arch     string
	ExitCode int
}

func (m *Metrics) WriteOpenTelemetry(ctx context.Context, meter metric.Meter) error {
	labels := []attribute.KeyValue{
		attribute.String("version", m.Version),
		attribute.String("command", m.Command),
		attribute.String("os", m.OS),
		attribute.String("arch", m.Arch),
		attribute.Int("exit_code", m.ExitCode),
	}

	runCounter, err := meter.SyncInt64().Counter("runs", instrument.WithDescription("werf runs counter"))
	if err != nil {
		return fmt.Errorf("unable to record runs counter: %w", err)
	}
	runCounter.Add(ctx, 1, labels...)
	logboek.Context(ctx).Debug().LogF("Telemetry: incremented runs counter: %#v\n", runCounter)

	return nil
}
*/
