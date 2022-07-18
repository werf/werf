package telemetry

type EventType string

const (
	CommandStartedEvent EventType = "CommandStarted"
	CommandExitedEvent  EventType = "CommandExited"
)

type Event interface {
	GetType() EventType
}

type CommandOption struct {
	Name  string `json:"name"`
	AsCli bool   `json:"asCli"`
	AsEnv bool   `json:"asEnv"`
	Count int    `json:"count"`
}

func NewCommandStarted(commandOptions []CommandOption) *CommandStarted {
	return &CommandStarted{CommandOptions: commandOptions}
}

type CommandStarted struct {
	CommandOptions []CommandOption `json:"commandOptions,omitempty"`
}

func (e *CommandStarted) GetType() EventType { return CommandStartedEvent }

func NewCommandExited(exitCode int, durationMs int64) *CommandExited {
	return &CommandExited{ExitCode: exitCode, DurationMs: durationMs}
}

type CommandExited struct {
	ExitCode   int   `json:"exitCode"`
	DurationMs int64 `json:"durationMs"`
}

func (e *CommandExited) GetType() EventType { return CommandExitedEvent }
