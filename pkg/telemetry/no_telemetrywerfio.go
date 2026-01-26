package telemetry

import "context"

type NoTelemetryWerfIO struct{}

func (t *NoTelemetryWerfIO) CommandStarted(context.Context)                                  {}
func (t *NoTelemetryWerfIO) SetUserID(context.Context, string)                               {}
func (t *NoTelemetryWerfIO) SetProjectID(context.Context, string)                            {}
func (t *NoTelemetryWerfIO) SetCommand(context.Context, string)                              {}
func (t *NoTelemetryWerfIO) CommandExited(context.Context, int)                              {}
func (t *NoTelemetryWerfIO) SetCommandOptions(context.Context, []CommandOption)              {}
func (t *NoTelemetryWerfIO) UnshallowFailed(context.Context, error)                          {}
func (t *NoTelemetryWerfIO) BuildStarted(context.Context, int)                               {}
func (t *NoTelemetryWerfIO) BuildFinished(context.Context, bool)                             {}
func (t *NoTelemetryWerfIO) ImageBuildFinished(context.Context, string, int64, bool)         {}
func (t *NoTelemetryWerfIO) StageBuildFinished(context.Context, string, string, int64, bool) {}
