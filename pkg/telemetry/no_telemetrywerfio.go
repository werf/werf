package telemetry

import "context"

type NoTelemetryWerfIO struct{}

func (t *NoTelemetryWerfIO) CommandStarted(context.Context)                     {}
func (t *NoTelemetryWerfIO) SetUserID(context.Context, string)                  {}
func (t *NoTelemetryWerfIO) SetProjectID(context.Context, string)               {}
func (t *NoTelemetryWerfIO) SetCommand(context.Context, string)                 {}
func (t *NoTelemetryWerfIO) CommandExited(context.Context, int)                 {}
func (t *NoTelemetryWerfIO) SetCommandOptions(context.Context, []CommandOption) {}
func (t *NoTelemetryWerfIO) UnshallowFailed(context.Context, error)             {}
