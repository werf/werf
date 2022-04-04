package container_backend

import "fmt"

type StapelStageType int

//go:generate stringer -type=StapelStageType
const (
	FromStage StapelStageType = iota
	UserCommandsStage
	DockerInstructionsStage
)

type BuildStapelStageOptionsInterface interface {
	SetBaseImage(baseImage string) BuildStapelStageOptionsInterface
	AddLabels(labels map[string]string) BuildStapelStageOptionsInterface
	AddBuildVolumes(volumes ...string) BuildStapelStageOptionsInterface

	UserCommandsStage() UserCommandsStageOptionsInterface
}

type UserCommandsStageOptionsInterface interface {
	AddUserCommands(commands ...string) UserCommandsStageOptionsInterface
}

type BuildStapelStageOptions struct {
	BaseImage    string
	Labels       []string
	BuildVolumes []string

	UserCommandsStageOptions
}

func (opts *BuildStapelStageOptions) SetBaseImage(baseImage string) BuildStapelStageOptionsInterface {
	opts.BaseImage = baseImage
	return opts
}

func (opts *BuildStapelStageOptions) AddLabels(labels map[string]string) BuildStapelStageOptionsInterface {
	for k, v := range labels {
		opts.Labels = append(opts.Labels, fmt.Sprintf("%s=%s", k, v))
	}
	return opts
}

func (opts *BuildStapelStageOptions) AddBuildVolumes(volumes ...string) BuildStapelStageOptionsInterface {
	opts.BuildVolumes = append(opts.BuildVolumes, volumes...)
	return opts
}

func (opts *BuildStapelStageOptions) UserCommandsStage() UserCommandsStageOptionsInterface {
	return &opts.UserCommandsStageOptions
}

type UserCommandsStageOptions struct {
	Commands []string
}

func (opts *UserCommandsStageOptions) AddUserCommands(commands ...string) UserCommandsStageOptionsInterface {
	opts.Commands = append(opts.Commands, commands...)
	return opts
}
