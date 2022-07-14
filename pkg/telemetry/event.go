package telemetry

type EventType string

const (
	CommandStartedEvent EventType = "CommandStarted"
	CommandExitedEvent  EventType = "CommandExited"
)

type Event interface {
	GetType() EventType
	GetData() interface{}
}

func NewCommandStarted(commandOptions []CommandOption) *CommandStarted {
	return &CommandStarted{commandOptions: commandOptions}
}

type CommandStarted struct {
	commandOptions []CommandOption
}

func (e *CommandStarted) GetType() EventType { return CommandStartedEvent }
func (e *CommandStarted) GetData() interface{} {
	if len(e.commandOptions) > 0 {
		return map[string]interface{}{"commandOptions": e.commandOptions}
	}
	return nil
}

func NewCommandExited(exitCode int) *CommandExited { return &CommandExited{exitCode: exitCode} }

type CommandExited struct {
	exitCode int
}

func (e *CommandExited) GetType() EventType   { return CommandExitedEvent }
func (e *CommandExited) GetData() interface{} { return map[string]interface{}{"exitCode": e.exitCode} }
