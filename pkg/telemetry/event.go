package telemetry

import "os"

type EventType string

const (
	CommandStartedEvent     EventType = "CommandStarted"
	CommandExitedEvent      EventType = "CommandExited"
	UnshallowFailedEvent    EventType = "UnshallowFailed"
	BuildStartedEvent       EventType = "BuildStarted"
	BuildFinishedEvent      EventType = "BuildFinished"
	ImageBuildFinishedEvent EventType = "ImageBuildFinished"
	StageBuildFinishedEvent EventType = "StageBuildFinished"
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

type BuildStarted struct {
	ImagesCount int `json:"imagesCount"`
}

func NewBuildStarted(imagesCount int) *BuildStarted {
	return &BuildStarted{ImagesCount: imagesCount}
}

func (*BuildStarted) GetType() EventType { return BuildStartedEvent }

type BuildFinished struct {
	DurationMs  int64 `json:"durationMs"`
	Success     bool  `json:"success"`
	ImagesCount int   `json:"imagesCount"`
}

func NewBuildFinished(durationMs int64, success bool, imagesCount int) *BuildFinished {
	return &BuildFinished{
		DurationMs:  durationMs,
		Success:     success,
		ImagesCount: imagesCount,
	}
}

func (*BuildFinished) GetType() EventType { return BuildFinishedEvent }

type ImageBuildFinished struct {
	Image      string `json:"image"`
	DurationMs int64  `json:"durationMs"`
	Rebuilt    bool   `json:"rebuilt"`
}

func NewImageBuildFinished(image string, durationMs int64, rebuilt bool) *ImageBuildFinished {
	return &ImageBuildFinished{
		Image:      image,
		DurationMs: durationMs,
		Rebuilt:    rebuilt,
	}
}

func (*ImageBuildFinished) GetType() EventType { return ImageBuildFinishedEvent }

type StageBuildFinished struct {
	Image      string `json:"image"`
	Stage      string `json:"stage"`
	DurationMs int64  `json:"durationMs"`
	FromCache  bool   `json:"fromCache"`
}

func NewStageBuildFinished(image, stage string, durationMs int64, fromCache bool) *StageBuildFinished {
	return &StageBuildFinished{
		Image:      image,
		Stage:      stage,
		DurationMs: durationMs,
		FromCache:  fromCache,
	}
}

func (*StageBuildFinished) GetType() EventType { return StageBuildFinishedEvent }
