package telemetry

import "context"

type NoTelemetryWerfIO struct{}

func (t *NoTelemetryWerfIO) CommandStarted(context.Context) {}
func (t *NoTelemetryWerfIO) SetProjectID(string)            {}
func (t *NoTelemetryWerfIO) SetCommand(string)              {}
