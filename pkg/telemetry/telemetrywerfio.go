package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

const (
	spanName      = "telemetry.werf.io"
	schemaVersion = 2
)

type TelemetryWerfIOInterface interface {
	SetUserID(ctx context.Context, projectID string)
	SetProjectID(ctx context.Context, projectID string)
	SetCommand(ctx context.Context, command string)
	SetCommandOptions(ctx context.Context, options []CommandOption)

	CommandStarted(ctx context.Context)
	CommandExited(ctx context.Context, exitCode int)
	UnshallowFailed(ctx context.Context, err error)
}

type TelemetryWerfIO struct {
	handleErrorFunc func(err error)
	tracerProvider  *sdktrace.TracerProvider
	traceExporter   *otlptrace.Exporter

	startedAt      time.Time
	executionID    string
	userID         string
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
				sdktrace.WithExportTimeout(1300*time.Millisecond),
			),
		),
		traceExporter: e,
		executionID:   uuid.New().String(),
		startedAt:     time.Now(),
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

func (t *TelemetryWerfIO) SetUserID(_ context.Context, userID string) {
	t.userID = userID
}

func (t *TelemetryWerfIO) SetProjectID(ctx context.Context, projectID string) {
	t.projectID = projectID
}

func (t *TelemetryWerfIO) CommandStarted(ctx context.Context) {
	t.sendEvent(ctx, NewCommandStarted(t.commandOptions))
}

func (t *TelemetryWerfIO) CommandExited(ctx context.Context, exitCode int) {
	duration := time.Now().Sub(t.startedAt)
	t.sendEvent(ctx, NewCommandExited(exitCode, int64(duration/time.Millisecond)))
}

func (t *TelemetryWerfIO) UnshallowFailed(ctx context.Context, err error) {
	t.sendEvent(ctx, NewUnshallowFailed(err.Error()))
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
	attributes["expDeployEngine"] = helm.IsExperimentalEngine()

	{
		if isCI := util.GetBoolEnvironmentDefaultFalse("CI"); isCI {
			attributes["ci"] = true
		}
		if isGitlabCI := util.GetBoolEnvironmentDefaultFalse("GITLAB_CI"); isGitlabCI {
			attributes["ciName"] = "gitlab"
		}
		if val := os.Getenv("JENKINS_URL"); val != "" {
			attributes["ciName"] = "jenkins"
		}
		if isTravis := util.GetBoolEnvironmentDefaultFalse("TRAVIS"); isTravis {
			attributes["ciName"] = "travis"
		}
		if isGithubActions := util.GetBoolEnvironmentDefaultFalse("GITHUB_ACTIONS"); isGithubActions {
			attributes["ciName"] = "github-actions"
		}
		if isCircleCI := util.GetBoolEnvironmentDefaultFalse("CIRCLECI"); isCircleCI {
			attributes["ciName"] = "circleci"
		}
		if val := os.Getenv("TEAMCITY_VERSION"); val != "" {
			attributes["ciName"] = "teamcity"
		}
		if isBuddy := util.GetBoolEnvironmentDefaultFalse("BUDDY"); isBuddy {
			attributes["ciName"] = "buddy"
		}
		if val := os.Getenv("GO_SERVER_URL"); val != "" {
			attributes["ciName"] = "gocd"
		}
	}

	// add extra attributes
	{
		extraAttributes := map[string]interface{}{}
		env := os.Environ()
		sort.Strings(env)
		for _, keyValue := range env {
			parts := strings.SplitN(keyValue, "=", 2)
			if strings.HasPrefix(parts[0], "WERF_TELEMETRY_EXTRA_ATTRIBUTE_") {
				valueParts := strings.SplitN(parts[1], "=", 2)

				if len(valueParts) != 2 {
					LogF("env with extra attribute %q is not valid", keyValue)
					continue
				}

				attrKey := valueParts[0]
				attrVal := valueParts[1]
				extraAttributes[attrKey] = attrVal
			}
		}

		attributes["extra"] = extraAttributes
	}

	return attributes
}

func (t *TelemetryWerfIO) sendEvent(ctx context.Context, event Event) error {
	LogF("start sending event")

	trc := t.getTracer()
	_, span := trc.Start(ctx, spanName)

	var attributes []attribute.KeyValue
	{
		ts := time.Now().UnixMilli()

		rawAttributes, err := json.Marshal(t.getAttributes())
		if err != nil {
			return fmt.Errorf("unable to marshal attributes: %w", err)
		}

		rawEventData, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("unable to marshal event data: %w", err)
		}

		attributes = []attribute.KeyValue{
			attribute.Key("ts").Int64(ts),
			attribute.Key("executionID").String(t.executionID),
			attribute.Key("userID").String(t.userID),
			attribute.Key("projectID").String(t.projectID),
			attribute.Key("command").String(t.command),
			attribute.Key("eventType").String(string(event.GetType())),
			attribute.Key("eventData").String(string(rawEventData)),
			attribute.Key("attributes").String(string(rawAttributes)),
			attribute.Key("schemaVersion").Int64(schemaVersion),
		}
	}

	span.SetAttributes(attributes...)
	span.End()

	var attributesString string
	{
		for i, attr := range attributes {
			if i != 0 {
				attributesString += ", "
			}

			var valueFormat string
			{
				if attr.Value.Type() == attribute.STRING {
					valueFormat = "=%q"
				} else {
					valueFormat = "=%v"
				}
			}

			attributesString += fmt.Sprintf("%s", attr.Key)
			attributesString += fmt.Sprintf(valueFormat, attr.Value.AsInterface())
		}
	}

	LogF("sent event: %s", attributesString)

	return nil
}
