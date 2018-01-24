package config

type RubyType interface {
	TagYAML() string
}

type Config struct {
	Dimg      []Dimg      `yaml:"_dimg"`
	DimgGroup []DimgGroup `yaml:"_dimg_group"`
}

func (cfg Config) TagYAML() string {
	return "!ruby/object:Dapp::Config::Config"
}

type DimgGroup struct {
	Dimg      []Dimg      `yaml:"_dimg"`
	DimgGroup []DimgGroup `yaml:"_dimg_group"`
}

func (cfg DimgGroup) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::DimgGroup"
}

type Dimg struct {
	Name        string           `yaml:"_name"`
	Docker      DockerDimg       `yaml:"_docker"`
	Builder     string           `yaml:"_builder"`
	Shell       *ShellDimg       `yaml:"_shell"`
	Chef        Chef             `yaml:"_chef"`
	Artifact    []ArtifactExport `yaml:"_artifact"`
	GitArtifact GitArtifact      `yaml:"_git_artifact"`
	Mount       []Mount          `yaml:"_mount"`
}

func (cfg Dimg) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::Dimg"
}

type ArtifactDimg struct {
	Dimg
	Docker DockerArtifact `yaml:"_docker"`
	Shell  *ShellArtifact `yaml:"_shell"`
}

func (cfg ArtifactDimg) TagYAML() string {
	return "!ruby/hash:Dapp::Dimg::Config::Directive::ArtifactDimg"
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

func (cfg DockerDimg) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::Docker::Dimg"
}

type DockerArtifact struct {
	DockerBase
}

func (cfg DockerArtifact) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::Docker::Artifact"
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

func (cfg ShellDimg) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::Shell::Dimg"
}

type ShellArtifact struct {
	ShellDimg
	BuildArtifact StageCommand `yaml:"_build_artifact"`
}

func (cfg ShellArtifact) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::Shell::Artifact"
}

type StageCommand struct {
	Version string   `yaml:"_version"`
	Run     []string `yaml:"_run"`
}

func (cfg StageCommand) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::GitArtifactLocal::Export::StageDependencies"
}

type Chef struct {
	Dimod      []string       `yaml:"_dimod"`
	Recipe     []string       `yaml:"_recipe"`
	Attributes ChefAttributes `yaml:"_attributes"`
}

func (cfg Chef) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::Chef"
}

type ChefAttributes map[interface{}]interface{}

func (cfg ChefAttributes) TagYAML() string {
	return "!ruby/hash:Dapp::Dimg::Config::Directive::Chef::Attributes"
}

type ArtifactExport struct {
	ArtifactBaseExport
	Config ArtifactDimg `yaml:"_config"`
	Before string       `yaml:"_before"`
	After  string       `yaml:"_after"`
}

func (cfg ArtifactExport) TagYAML() string {
	return "!ruby/hash:Dapp::Dimg::Config::Directive::Artifact::Export"
}

type GitArtifact struct {
	Local  []GitArtifactLocal  `yaml:"_local"`
	Remote []GitArtifactRemote `yaml:"_remote"`
}

func (cfg GitArtifact) TagYAML() string {
	return "!ruby/hash:Dapp::Dimg::Config::Directive::Dimg::InstanceMethods::GitArtifact"
}

type GitArtifactLocal struct {
	Export []GitArtifactLocalExport `yaml:"_export"`
}

func (cfg GitArtifactLocal) TagYAML() string {
	return "!ruby/hash:Dapp::Dimg::Config::Directive::GitArtifactLocal"
}

type GitArtifactLocalExport struct {
	ArtifactBaseExport
	As                string            `yaml:"_as"`
	StageDependencies StageDependencies `yaml:"_stage_dependencies"`
}

func (cfg GitArtifactLocalExport) TagYAML() string {
	return "!ruby/hash:Dapp::Dimg::Config::Directive::GitArtifactLocal::Export"
}

type StageDependencies struct {
	Install       []string `yaml:"_install"`
	Setup         []string `yaml:"_setup"`
	BeforeSetup   []string `yaml:"_before_setup"`
	BuildArtifact []string `yaml:"_build_artifact"`
}

func (cfg StageDependencies) TagYAML() string {
	return "!ruby/hash:Dapp::Dimg::Config::Directive::Shell::Dimg::StageCommand"
}

type GitArtifactRemote struct {
	Export []GitArtifactRemoteExport `yaml:"_export"`
}

func (cfg GitArtifactRemote) TagYAML() string {
	return "!ruby/hash:Dapp::Dimg::Config::Directive::GitArtifactRemote"
}

type GitArtifactRemoteExport struct {
	GitArtifactLocalExport
	Url    string `yaml:"_url"`
	Name   string `yaml:"_name"`
	Branch string `yaml:"_branch"`
	Commit string `yaml:"_commit"`
}

func (cfg GitArtifactRemoteExport) TagYAML() string {
	return "!ruby/hash:Dapp::Dimg::Config::Directive::GitArtifactRemote::Export"
}

type ArtifactBaseExport struct {
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

func (cfg Mount) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::Mount"
}
