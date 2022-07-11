package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/werf/werf/pkg/werf"
)

const (
	spanName      = "telemetry.werf.io"
	schemaVersion = 1
)

type TelemetryWerfIOInterface interface {
	SetProjectID(ctx context.Context, projectID string)
	SetCommand(ctx context.Context, command string)
	CommandStarted(ctx context.Context)
}

type TelemetryWerfIO struct {
	handleErrorFunc func(err error)
	tracerProvider  *sdktrace.TracerProvider
	traceExporter   *otlptrace.Exporter

	executionID string
	projectID   string
	command     string
}

type TelemetryWerfIOOptions struct {
	HandleErrorFunc func(err error)
}

func NewTelemetryWerfIO(url string, opts TelemetryWerfIOOptions) (*TelemetryWerfIO, error) {
	e, err := NewTraceExporter(url)
	if err != nil {
		return nil, fmt.Errorf("unable to create telemetry trace exporter: %w", err)
	}

	return &TelemetryWerfIO{
		handleErrorFunc: opts.HandleErrorFunc,
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

func (t *TelemetryWerfIO) SetCommand(ctx context.Context, command string) {
	t.command = command
}

func (t *TelemetryWerfIO) SetProjectID(ctx context.Context, projectID string) {
	t.projectID = projectID
}

func (t *TelemetryWerfIO) CommandStarted(ctx context.Context) {
	t.sendEvent(ctx, CommandStartedEvent, nil)
}

func (t *TelemetryWerfIO) getAttributes() map[string]interface{} {
	attributes := map[string]interface{}{
		"version": werf.Version,
		"os":      runtime.GOOS,
		"arch":    runtime.GOARCH,
	}
	if val := os.Getenv("TRDL_USE_WERF_GROUP_CHANNEL"); val != "" {
		attributes["groupChannel"] = val
	}

	return attributes
}

func (t *TelemetryWerfIO) sendEvent(ctx context.Context, eventType EventType, eventData interface{}) error {
	trc := t.getTracer()
	_, span := trc.Start(ctx, spanName)

	ts := time.Now().UnixMilli()

	span.SetAttributes(attribute.Key("ts").Int64(ts))
	span.SetAttributes(attribute.Key("executionID").String(t.executionID))
	span.SetAttributes(attribute.Key("projectID").String(t.projectID))
	span.SetAttributes(attribute.Key("command").String(t.command))

	rawAttributes, err := json.Marshal(t.getAttributes())
	if err != nil {
		return fmt.Errorf("unable to marshal attributes: %w", err)
	}
	span.SetAttributes(attribute.Key("attributes").String(string(rawAttributes)))

	span.SetAttributes(attribute.Key("eventType").String(string(eventType)))

	rawEventData, err := json.Marshal(eventData)
	if err != nil {
		return fmt.Errorf("unable to marshal event data: %w", err)
	}
	span.SetAttributes(attribute.Key("eventData").String(string(rawEventData)))
	span.SetAttributes(attribute.Key("schemaVersion").Int64(schemaVersion))

	span.End()

	if IsLogsEnabled() {
		fmt.Printf("Telemetry: sent event: ts=%d executionID=%q projectID=%q command=%q attributes=%q eventType=%q eventData=%q schemaVersion=%d\n", ts, t.executionID, t.projectID, t.command, string(rawAttributes), string(eventType), string(rawEventData), schemaVersion)
	}

	return nil
}
