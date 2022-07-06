package telemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type TelemetryWerfIOInterface interface {
	SetProjectID(projectID string)
	SetCommand(command string)
	CommandStarted(ctx context.Context)
}

type TelemetryWerfIO struct {
	tracerProvider *sdktrace.TracerProvider
	traceExporter  *otlptrace.Exporter

	executionID string
	projectID   string
	command     string
}

func NewTelemetryWerfIO(url string) (*TelemetryWerfIO, error) {
	e, err := NewTraceExporter(url)
	if err != nil {
		return nil, fmt.Errorf("unable to create telemetry trace exporter: %w", err)
	}

	return &TelemetryWerfIO{
		tracerProvider: sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(e,
				sdktrace.WithBatchTimeout(0),
				sdktrace.WithExportTimeout(3*time.Second),
			),
		),
		traceExporter: e,
		executionID:   uuid.New().String(),
	}, nil
}

// TODO: start background procedure
func (t *TelemetryWerfIO) Start(ctx context.Context) error {
	if err := t.traceExporter.Start(ctx); err != nil {
		return fmt.Errorf("error starting telemetry trace exporter: %w", err)
	}
	return nil
}

func (t *TelemetryWerfIO) Shutdown(ctx context.Context) error {
	if err := t.tracerProvider.ForceFlush(ctx); err != nil {
		return fmt.Errorf("unable to force flush tracer provider: %w", err)
	}

	if err := t.traceExporter.Shutdown(ctx); err != nil {
		return fmt.Errorf("unable to shutdown trace exporter: %w", err)
	}

	if err := t.tracerProvider.Shutdown(ctx); err != nil {
		return fmt.Errorf("unable to shutdown trace provider: %w", err)
	}

	return nil
}

func (t *TelemetryWerfIO) getTracer() trace.Tracer {
	return t.tracerProvider.Tracer("telemetry.werf.io")
}

func (t *TelemetryWerfIO) SetCommand(command string) {
	t.command = command
}

func (t *TelemetryWerfIO) SetProjectID(projectID string) {
	t.projectID = projectID
}

func (t *TelemetryWerfIO) CommandStarted(ctx context.Context) {
	trc := t.getTracer()

	_, span := trc.Start(ctx, "telemetry.werf.io")
	span.SetAttributes(attribute.Key("ts").Int64(time.Now().UnixMilli()))
	span.SetAttributes(attribute.Key("executionID").String(t.executionID))
	span.SetAttributes(attribute.Key("projectID").String(t.projectID))
	span.SetAttributes(attribute.Key("command").String(t.command))
	span.SetAttributes(attribute.Key("eventType").String("CommandStarted"))
	span.SetAttributes(attribute.Key("data").String(`{}`))
	span.End()
}
