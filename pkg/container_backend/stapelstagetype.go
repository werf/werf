package container_backend

import (
	"fmt"
	"io"
)

type StapelStageType int

//go:generate stringer -type=StapelStageType
const (
	FromStage StapelStageType = iota
	UserCommandsStage
	DockerInstructionsStage
	DependenciesStage
	DataArchivesStage
)

type BuildStapelStageOptionsInterface interface {
	SetBaseImage(baseImage string) BuildStapelStageOptionsInterface
	AddLabels(labels map[string]string) BuildStapelStageOptionsInterface
	AddVolumes(volumes []string) BuildStapelStageOptionsInterface
	AddBuildVolumes(volumes ...string) BuildStapelStageOptionsInterface
	AddExpose(expose []string) BuildStapelStageOptionsInterface
	AddEnvs(envs map[string]string) BuildStapelStageOptionsInterface
	SetCmd(cmd []string) BuildStapelStageOptionsInterface
	SetEntrypoint(entrypoint []string) BuildStapelStageOptionsInterface
	SetUser(user string) BuildStapelStageOptionsInterface
	SetWorkdir(workdir string) BuildStapelStageOptionsInterface
	SetHealthcheck(healthcheck string) BuildStapelStageOptionsInterface

	AddDataArchives(archives ...DataArchive) BuildStapelStageOptionsInterface
	AddPathsToRemove(paths ...string) BuildStapelStageOptionsInterface

	UserCommandsStage() UserCommandsStageOptionsInterface
	DependenciesStage() DependenciesStageOptionsInterface
}

type UserCommandsStageOptionsInterface interface {
	AddUserCommands(commands ...string) UserCommandsStageOptionsInterface
}

type DependenciesStageOptionsInterface interface {
	AddDependencyImport(imageName, fromPath, toPath string, includePaths, excludePaths []string, owner, group string) DependenciesStageOptionsInterface
}

type BuildStapelStageOptions struct {
	BaseImage    string
	Labels       []string
	BuildVolumes []string
	Volumes      []string
	Expose       []string
	Envs         map[string]string
	Cmd          []string
	Entrypoint   []string
	User         string
	Workdir      string
	Healthcheck  string

	DataArchives  []DataArchive
	PathsToRemove []string

	UserCommandsStageOptions
	DependenciesStageOptions
}

type ArchiveType int

//go:generate stringer -type=ArchiveType
const (
	FileArchive ArchiveType = iota
	DirectoryArchive
)

type DataArchive struct {
	Data io.ReadCloser
	Type ArchiveType
	To   string
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

func (opts *BuildStapelStageOptions) AddVolumes(volumes []string) BuildStapelStageOptionsInterface {
	opts.Volumes = append(opts.Volumes, volumes...)
	return opts
}

func (opts *BuildStapelStageOptions) AddExpose(expose []string) BuildStapelStageOptionsInterface {
	opts.Expose = append(opts.Expose, expose...)
	return opts
}

func (opts *BuildStapelStageOptions) AddEnvs(envs map[string]string) BuildStapelStageOptionsInterface {
	if opts.Envs == nil {
		opts.Envs = map[string]string{}
	}

	for k, v := range envs {
		opts.Envs[k] = v
	}

	return opts
}

func (opts *BuildStapelStageOptions) SetCmd(cmd []string) BuildStapelStageOptionsInterface {
	opts.Cmd = cmd
	return opts
}

func (opts *BuildStapelStageOptions) SetEntrypoint(entrypoint []string) BuildStapelStageOptionsInterface {
	opts.Entrypoint = entrypoint
	return opts
}

func (opts *BuildStapelStageOptions) SetUser(user string) BuildStapelStageOptionsInterface {
	opts.User = user
	return opts
}

func (opts *BuildStapelStageOptions) SetWorkdir(workdir string) BuildStapelStageOptionsInterface {
	opts.Workdir = workdir
	return opts
}

func (opts *BuildStapelStageOptions) SetHealthcheck(healthcheck string) BuildStapelStageOptionsInterface {
	opts.Healthcheck = healthcheck
	return opts
}

func (opts *BuildStapelStageOptions) AddBuildVolumes(volumes ...string) BuildStapelStageOptionsInterface {
	opts.BuildVolumes = append(opts.BuildVolumes, volumes...)
	return opts
}

func (opts *BuildStapelStageOptions) AddDataArchives(archives ...DataArchive) BuildStapelStageOptionsInterface {
	opts.DataArchives = append(opts.DataArchives, archives...)
	return opts
}

func (opts *BuildStapelStageOptions) AddPathsToRemove(paths ...string) BuildStapelStageOptionsInterface {
	opts.PathsToRemove = append(opts.PathsToRemove, paths...)
	return opts
}

func (opts *BuildStapelStageOptions) UserCommandsStage() UserCommandsStageOptionsInterface {
	return &opts.UserCommandsStageOptions
}

func (opts *BuildStapelStageOptions) DependenciesStage() DependenciesStageOptionsInterface {
	return &opts.DependenciesStageOptions
}

type UserCommandsStageOptions struct {
	Commands []string
}

func (opts *UserCommandsStageOptions) AddUserCommands(commands ...string) UserCommandsStageOptionsInterface {
	opts.Commands = append(opts.Commands, commands...)
	return opts
}

type DependenciesStageOptions struct {
	Imports []DependencyImport
}

type DependencyImport struct {
	ImageName    string
	FromPath     string
	ToPath       string
	IncludePaths []string
	ExcludePaths []string
	Owner        string
	Group        string
}

func (opts *DependenciesStageOptions) AddDependencyImport(imageName, fromPath, toPath string, includePaths, excludePaths []string, owner, group string) DependenciesStageOptionsInterface {
	opts.Imports = append(opts.Imports, DependencyImport{
		ImageName:    imageName,
		FromPath:     fromPath,
		ToPath:       toPath,
		IncludePaths: includePaths,
		ExcludePaths: excludePaths,
		Owner:        owner,
		Group:        group,
	})
	return opts
}
