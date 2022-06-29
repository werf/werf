package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

func MessageEvent(ctx context.Context, msg string) error {
	if !IsEnabled() {
		return nil
	}

	trc := tracerProvider.Tracer("telemetry.werf.io")

	_, span := trc.Start(ctx, "message")

	span.SetAttributes(attribute.Key("ts").Int64(time.Now().UnixMilli()))
	span.SetAttributes(attribute.Key("executionID").String(executionID))
	span.SetAttributes(attribute.Key("projectID").String(projectID))
	span.SetAttributes(attribute.Key("command").String(command))
	span.SetAttributes(attribute.Key("eventType").String("message"))

	data, err := json.Marshal(map[string]interface{}{"message": msg})
	if err != nil {
		return fmt.Errorf("unable to marshal message data: %w", err)
	}
	span.SetAttributes(attribute.Key("data").String(string(data)))
	span.End()

	return nil
}
