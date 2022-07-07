package telemetry

import "context"

type NoTelemetryWerfIO struct{}

func (t *NoTelemetryWerfIO) CommandStarted(context.Context)       {}
func (t *NoTelemetryWerfIO) SetProjectID(context.Context, string) {}
func (t *NoTelemetryWerfIO) SetCommand(context.Context, string)   {}
