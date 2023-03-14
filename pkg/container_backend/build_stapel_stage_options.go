package container_backend

import (
	"fmt"
	"io"
)

type AddDataArchiveOptions struct {
	Owner, Group string
}

type BuildStapelStageOptionsInterface interface {
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

	AddDataArchive(archive io.ReadCloser, archiveType ArchiveType, to string, o AddDataArchiveOptions) BuildStapelStageOptionsInterface
	RemoveData(removeType RemoveType, paths, keepParentDirs []string) BuildStapelStageOptionsInterface
	AddDependencyImport(imageName, fromPath, toPath string, includePaths, excludePaths []string, owner, group string) BuildStapelStageOptionsInterface
}

type BuildStapelStageOptions struct {
	TargetPlatform string

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

	DataArchiveSpecs      []DataArchiveSpec
	RemoveDataSpecs       []RemoveDataSpec
	DependencyImportSpecs []DependencyImportSpec
}

type ArchiveType int

//go:generate stringer -type=ArchiveType
const (
	FileArchive ArchiveType = iota
	DirectoryArchive
)

type DataArchiveSpec struct {
	Archive      io.ReadCloser
	Type         ArchiveType
	To           string
	Owner, Group string
}

type RemoveType int

//go:generate stringer -type=RemoveType
const (
	RemoveExactPath RemoveType = iota
	RemoveExactPathWithEmptyParentDirs
	RemoveInsidePath
)

type RemoveDataSpec struct {
	Type           RemoveType
	Paths          []string
	KeepParentDirs []string
}

type DependencyImportSpec struct {
	ImageName    string
	FromPath     string
	ToPath       string
	IncludePaths []string
	ExcludePaths []string
	Owner        string
	Group        string
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

func (opts *BuildStapelStageOptions) AddDataArchive(archive io.ReadCloser, archiveType ArchiveType, to string, o AddDataArchiveOptions) BuildStapelStageOptionsInterface {
	opts.DataArchiveSpecs = append(opts.DataArchiveSpecs, DataArchiveSpec{
		Archive: archive,
		Type:    archiveType,
		To:      to,
		Owner:   o.Owner,
		Group:   o.Group,
	})
	return opts
}

func (opts *BuildStapelStageOptions) RemoveData(removeType RemoveType, paths, keepParentDirs []string) BuildStapelStageOptionsInterface {
	opts.RemoveDataSpecs = append(opts.RemoveDataSpecs, RemoveDataSpec{
		Type:           removeType,
		Paths:          paths,
		KeepParentDirs: keepParentDirs,
	})
	return opts
}

func (opts *BuildStapelStageOptions) AddDependencyImport(imageName, fromPath, toPath string, includePaths, excludePaths []string, owner, group string) BuildStapelStageOptionsInterface {
	opts.DependencyImportSpecs = append(opts.DependencyImportSpecs, DependencyImportSpec{
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
