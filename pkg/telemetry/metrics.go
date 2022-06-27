package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"

	"github.com/werf/logboek"
)

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
