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
	SetCommandOptions(ctx context.Context, options []CommandOption)

	CommandStarted(ctx context.Context)
	CommandExited(ctx context.Context, exitCode int)
}

type CommandOption struct {
	Name  string `json:"name"`
	AsCli bool   `json:"asCli"`
	AsEnv bool   `json:"asEnv"`
	Count int    `json:"count"`
}

type TelemetryWerfIO struct {
	handleErrorFunc func(err error)
	tracerProvider  *sdktrace.TracerProvider
	traceExporter   *otlptrace.Exporter

	executionID    string
	projectID      string
	command        string
	commandOptions []CommandOption
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
				sdktrace.WithBatchTimeout(1*time.Millisecond), // send all available events immediately
				sdktrace.WithExportTimeout(800*time.Millisecond),
			),
		),
		traceExporter: e,
		executionID:   uuid.New().String(),
	}, nil
}

func (t *TelemetryWerfIO) Start(ctx context.Context) error {
	LogF("start trace exporter")
	if err := t.traceExporter.Start(ctx); err != nil {
		return fmt.Errorf("error starting telemetry trace exporter: %w", err)
	}
	return nil
}

func (t *TelemetryWerfIO) Shutdown(ctx context.Context) error {
	LogF("start shutdown")

	LogF("flush trace provider")
	if err := t.tracerProvider.ForceFlush(ctx); err != nil {
		return fmt.Errorf("unable to force flush tracer provider: %w", err)
	}

	LogF("shutdown trace exporter")
	if err := t.traceExporter.Shutdown(ctx); err != nil {
		return fmt.Errorf("unable to shutdown trace exporter: %w", err)
	}

	LogF("shutdown trace provider")
	if err := t.tracerProvider.Shutdown(ctx); err != nil {
		return fmt.Errorf("unable to shutdown trace provider: %w", err)
	}

	LogF("shutdown complete")

	return nil
}

func (t *TelemetryWerfIO) getTracer() trace.Tracer {
	return t.tracerProvider.Tracer("telemetry.werf.io")
}

func (t *TelemetryWerfIO) SetCommand(ctx context.Context, command string) {
	t.command = command
}

func (t *TelemetryWerfIO) SetCommandOptions(ctx context.Context, options []CommandOption) {
	t.commandOptions = options
}

func (t *TelemetryWerfIO) SetProjectID(ctx context.Context, projectID string) {
	t.projectID = projectID
}

func (t *TelemetryWerfIO) CommandStarted(ctx context.Context) {
	t.sendEvent(ctx, NewCommandStarted(t.commandOptions))
}

func (t *TelemetryWerfIO) CommandExited(ctx context.Context, exitCode int) {
	t.sendEvent(ctx, NewCommandExited(exitCode))
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

func (t *TelemetryWerfIO) sendEvent(ctx context.Context, event Event) error {
	trc := t.getTracer()
	_, span := trc.Start(ctx, spanName)

	LogF("start sending event")

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

	span.SetAttributes(attribute.Key("eventType").String(string(event.GetType())))

	rawEventData, err := json.Marshal(event.GetData())
	if err != nil {
		return fmt.Errorf("unable to marshal event data: %w", err)
	}
	span.SetAttributes(attribute.Key("eventData").String(string(rawEventData)))
	span.SetAttributes(attribute.Key("schemaVersion").Int64(schemaVersion))
	span.End()

	LogF("sent event: ts=%d executionID=%q projectID=%q command=%q attributes=%q eventType=%q eventData=%q schemaVersion=%d", ts, t.executionID, t.projectID, t.command, string(rawAttributes), string(event.GetType()), string(rawEventData), schemaVersion)

	return nil
}
