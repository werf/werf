package telemetry

import "os"

type EventType string

const (
	CommandStartedEvent  EventType = "CommandStarted"
	CommandExitedEvent   EventType = "CommandExited"
	UnshallowFailedEvent EventType = "UnshallowFailed"
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

func NewUnshallowFailed(errorMessage string) *UnshallowFailed {
	return &UnshallowFailed{
		ErrorMessage:        errorMessage,
		GitlabRunnerVersion: os.Getenv("CI_RUNNER_VERSION"),
		GitlabServerVersion: os.Getenv("CI_SERVER_VERSION"),
	}
}

type UnshallowFailed struct {
	ErrorMessage        string `json:"errorMessage"`
	GitlabRunnerVersion string `json:"gitlabRunnerVersion"`
	GitlabServerVersion string `json:"gitlabServerVersion"`
}

func (*UnshallowFailed) GetType() EventType { return UnshallowFailedEvent }
