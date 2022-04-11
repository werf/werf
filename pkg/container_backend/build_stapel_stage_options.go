package container_backend

import (
	"fmt"
	"io"
)

type BuildStapelStageOptionsInterface interface {
	SetBaseImage(baseImage string) BuildStapelStageOptionsInterface

	AddLabels(labels map[string]string) BuildStapelStageOptionsInterface
	AddVolumes(volumes []string) BuildStapelStageOptionsInterface
	AddExpose(expose []string) BuildStapelStageOptionsInterface
	AddEnvs(envs map[string]string) BuildStapelStageOptionsInterface
	SetCmd(cmd []string) BuildStapelStageOptionsInterface
	SetEntrypoint(entrypoint []string) BuildStapelStageOptionsInterface
	SetUser(user string) BuildStapelStageOptionsInterface
	SetWorkdir(workdir string) BuildStapelStageOptionsInterface
	SetHealthcheck(healthcheck string) BuildStapelStageOptionsInterface

	AddBuildVolumes(volumes ...string) BuildStapelStageOptionsInterface
	AddCommands(commands ...string) BuildStapelStageOptionsInterface

	AddDataArchives(archives ...DataArchive) BuildStapelStageOptionsInterface
	AddPathsToRemove(paths ...string) BuildStapelStageOptionsInterface
	AddDependencyImport(imageName, fromPath, toPath string, includePaths, excludePaths []string, owner, group string) BuildStapelStageOptionsInterface
}

type BuildStapelStageOptions struct {
	BaseImage string

	Labels      []string
	Volumes     []string
	Expose      []string
	Envs        map[string]string
	Cmd         []string
	Entrypoint  []string
	User        string
	Workdir     string
	Healthcheck string

	BuildVolumes []string
	Commands     []string

	DataArchives        []DataArchive
	PathsToRemove       []string
	DependenciesImports []DependencyImport
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

type DependencyImport struct {
	ImageName    string
	FromPath     string
	ToPath       string
	IncludePaths []string
	ExcludePaths []string
	Owner        string
	Group        string
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

func (opts *BuildStapelStageOptions) AddCommands(commands ...string) BuildStapelStageOptionsInterface {
	opts.Commands = append(opts.Commands, commands...)
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

func (opts *BuildStapelStageOptions) AddDependencyImport(imageName, fromPath, toPath string, includePaths, excludePaths []string, owner, group string) BuildStapelStageOptionsInterface {
	opts.DependenciesImports = append(opts.DependenciesImports, DependencyImport{
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
