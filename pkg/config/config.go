package config

type RubyType interface {
	GetRubyTypeTag() string
}

type Dimg struct {
	Docker      DockerDimg  `yaml:"_docker"`
	Builder     string      `yaml:"_builder"`
	Shell       ShellDimg   `yaml:"_shell"`
	Chef        Chef        `yaml:"_chef"`
	Artifact    []Artifact  `yaml:"_artifact"`
	GitArtifact GitArtifact `yaml:"_git_artifact"`
	Mount       []Mount     `yaml:"_mount"`
}

func (cfg *Dimg) GetRubyTypeTag() string {
	return "ruby/object:Dapp::Dimg::Config::Directive::Dimg"
}

type ArtifactDimg struct {
	Dimg
	Docker DockerArtifact `yaml:"_shell"`
	Shell  ShellArtifact  `yaml:"_shell"`
}

func (cfg *ArtifactDimg) GetRubyTypeTag() string {
	return "ruby/hash:Dapp::Dimg::Config::Directive::ArtifactDimg"
}

type DockerDimg struct {
	DockerBase
	Volume     []string          `yaml:"_volume"`
	Expose     []string          `yaml:"_expose"`
	Env        map[string]string `yaml:"_env"`
	Label      map[string]string `yaml:"_label"`
	Cmd        []string          `yaml:"_cmd"`
	Onbuild    []string          `yaml:"_onbuild"`
	Workdir    string            `yaml:"_workdir"`
	User       string            `yaml:"_user"`
	Entrypoint []string          `yaml:"_entrypoint"`
}

func (cfg *DockerDimg) GetRubyTypeTag() string {
	return "ruby/object:Dapp::Dimg::Config::Directive::Docker::Dimg"
}

type DockerArtifact struct {
	DockerBase
}

func (cfg *DockerArtifact) GetRubyTypeTag() string {
	return "ruby/object:Dapp::Dimg::Config::Directive::Docker::Artifact"
}

type DockerBase struct {
	From             string `yaml:"_from"`
	FromCacheVersion string `yaml:"_from_cache_version"`
}

type ShellDimg struct {
	Version       string       `yaml:"_version"`
	BeforeInstall StageCommand `yaml:"_before_install"`
	BeforeSetup   StageCommand `yaml:"_before_setup"`
	Install       StageCommand `yaml:"_install"`
	Setup         StageCommand `yaml:"_setup"`
}

func (cfg *ShellDimg) GetRubyTypeTag() string {
	return "ruby/object:Dapp::Dimg::Config::Directive::Shell::Dimg"
}

type ShellArtifact struct {
	ShellDimg
	BuildArtifact StageCommand `yaml:"_build_artifact"`
}

func (cfg *ShellArtifact) GetRubyTypeTag() string {
	return "ruby/object:Dapp::Dimg::Config::Directive::Shell::Artifact"
}

type StageCommand struct {
	Version string   `yaml:"_version"`
	Run     []string `yaml:"_run"`
}

func (cfg *StageCommand) GetRubyTypeTag() string {
	return "ruby/object:Dapp::Dimg::Config::Directive::GitArtifactLocal::Export::StageDependencies"
}

type Chef struct {
	Dimod      []string         `yaml:"_dimod"`
	Recipe     []string         `yaml:"_recipe"`
	Attributes []ChefAttributes `yaml:"_attributes"`
}

func (cfg *Chef) GetRubyTypeTag() string {
	return "ruby/object:Dapp::Dimg::Config::Directive::Chef"
}

type ChefAttributes map[interface{}]interface{}

func (cfg *ChefAttributes) GetRubyTypeTag() string {
	return "ruby/hash:Dapp::Dimg::Config::Directive::Chef::Attributes"
}

type Artifact struct {
	ArtifactBase
	Config ArtifactDimg `yaml:"_config"`
	Before string       `yaml:"_before"`
	After  string       `yaml:"_after"`
}

func (cfg *Artifact) GetRubyTypeTag() string {
	return "ruby/hash:Dapp::Dimg::Config::Directive::Artifact::Export"
}

type GitArtifact struct {
	Local  []GitArtifactLocal  `yaml:"_local"`
	Remote []GitArtifactRemote `yaml:"_remote"`
}

func (cfg *GitArtifact) GetRubyTypeTag() string {
	return "ruby/hash:Dapp::Dimg::Config::Directive::Dimg::InstanceMethods::GitArtifact"
}

type GitArtifactLocal struct {
	ArtifactBase
	As                string            `yaml:"_as"`
	StageDependencies StageDependencies `yaml:"_stage_dependencies"`
}

func (cfg *GitArtifactLocal) GetRubyTypeTag() string {
	return "ruby/hash:Dapp::Dimg::Config::Directive::GitArtifactLocal"
}

type StageDependencies struct {
	Install       []string `yaml:"_install"`
	Setup         []string `yaml:"_setup"`
	BeforeSetup   []string `yaml:"_before_setup"`
	BuildArtifact []string `yaml:"_build_artifact"`
}

func (cfg *StageDependencies) GetRubyTypeTag() string {
	return "ruby/hash:Dapp::Dimg::Config::Directive::Shell::Dimg::StageCommand"
}

type GitArtifactRemote struct {
	GitArtifactLocal
	Url    string `yaml:"_url"`
	Name   string `yaml:"_name"`
	Branch string `yaml:"_branch"`
	Commit string `yaml:"_commit"`
}

func (cfg *GitArtifactRemote) GetRubyTypeTag() string {
	return "ruby/hash:Dapp::Dimg::Config::Directive::GitArtifactRemote"
}

type ArtifactBase struct {
	Cwd          string   `yaml:"_cwd"`
	To           string   `yaml:"_to"`
	IncludePaths []string `yaml:"_include_paths"`
	ExcludePaths []string `yaml:"_exclude_paths"`
	Owner        string   `yaml:"_owner"`
	Group        string   `yaml:"_group"`
}

type Mount struct {
	To   string `yaml:"_to"`
	From string `yaml:"_from"`
	Type string `yaml:"_type"`
}

func (cfg *Mount) GetRubyTypeTag() string {
	return "ruby/object:Dapp::Dimg::Config::Directive::Mount"
}
