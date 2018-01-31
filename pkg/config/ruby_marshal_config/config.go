package ruby_marshal_config

type RubyType interface {
	TagYAML() string
}

type Config struct {
	DimgGroup `yaml:",inline"`
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
	DimgBase `yaml:",inline"`
	Docker   DockerDimg `yaml:"_docker,omitempty"`
	Shell    ShellDimg  `yaml:"_shell,omitempty"`
}

func (cfg Dimg) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::Dimg"
}

type DimgArtifact struct {
	DimgBase `yaml:",inline"`
	Docker   DockerArtifact `yaml:"_docker,omitempty"`
	Shell    ShellArtifact  `yaml:"_shell,omitempty"`
}

func (cfg DimgArtifact) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::ArtifactDimg"
}

type DimgBase struct {
	Name          string          `yaml:"_name"`
	Builder       Symbol          `yaml:"_builder"`
	Chef          Chef            `yaml:"_chef,omitempty"`
	ArtifactGroup []ArtifactGroup `yaml:"_artifact_groups,omitempty"`
	GitArtifact   GitArtifact     `yaml:"_git_artifact,omitempty"`
	Mount         []Mount         `yaml:"_mount,omitempty"`
}

type DockerDimg struct {
	DockerBase `yaml:",inline"`
	Volume     []string          `yaml:"_volume"`
	Expose     []string          `yaml:"_expose"`
	Env        map[Symbol]string `yaml:"_env"`
	Label      map[Symbol]string `yaml:"_label"`
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
	DockerBase `yaml:",inline"`
}

func (cfg DockerArtifact) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::Docker::Artifact"
}

type DockerBase struct {
	From             string `yaml:"_from"`
	FromCacheVersion string `yaml:"_from_cache_version,omitempty"`
}

type ShellDimg struct {
	Version       string       `yaml:"_version,omitempty"`
	BeforeInstall StageCommand `yaml:"_before_install,omitempty"`
	BeforeSetup   StageCommand `yaml:"_before_setup,omitempty"`
	Install       StageCommand `yaml:"_install,omitempty"`
	Setup         StageCommand `yaml:"_setup,omitempty"`
}

func (cfg ShellDimg) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::Shell::Dimg"
}

type ShellArtifact struct {
	ShellDimg     `yaml:",inline"`
	BuildArtifact StageCommand `yaml:"_build_artifact,omitempty"`
}

func (cfg ShellArtifact) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::Shell::Artifact"
}

type StageCommand struct {
	Version string   `yaml:"_version,omitempty"`
	Run     []string `yaml:"_run"`
}

func (cfg StageCommand) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::Shell::Dimg::StageCommand"
}

type Chef struct {
	Dimod      []string       `yaml:"_dimod"`
	Recipe     []string       `yaml:"_recipe"`
	Attributes ChefAttributes `yaml:"_attributes"`
	// TODO: Cookbook   []Cookbook     `yaml:"_cookbook"`
}

func (cfg Chef) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::Chef"
}

type ChefAttributes map[interface{}]interface{}

func (cfg ChefAttributes) TagYAML() string {
	return "!ruby/hash:Dapp::Dimg::Config::Directive::Chef::Attributes"
}

type ArtifactGroup struct {
	Export []ArtifactExport `yaml:"_export"`
}

func (cfg ArtifactGroup) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::ArtifactGroup"
}

type ArtifactExport struct {
	ArtifactBaseExport `yaml:",inline"`
	Config             DimgArtifact `yaml:"_config,omitempty"`
	Before             Symbol       `yaml:"_before,omitempty"`
	After              Symbol       `yaml:"_after,omitempty"`
}

func (cfg ArtifactExport) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::Artifact::Export"
}

type GitArtifact struct {
	Local  []GitArtifactLocal  `yaml:"_local"`
	Remote []GitArtifactRemote `yaml:"_remote"`
}

func (cfg GitArtifact) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::Dimg::InstanceMethods::GitArtifact"
}

type GitArtifactLocal struct {
	As     string                   `yaml:"_as,omitempty"`
	Export []GitArtifactLocalExport `yaml:"_export"`
}

func (cfg GitArtifactLocal) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::GitArtifactLocal"
}

type GitArtifactLocalExport struct {
	ArtifactBaseExport `yaml:",inline"`
	StageDependencies  StageDependencies `yaml:"_stage_dependencies"`
}

func (cfg GitArtifactLocalExport) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::GitArtifactLocal::Export"
}

type StageDependencies struct {
	Install       []string `yaml:"_install"`
	Setup         []string `yaml:"_setup"`
	BeforeSetup   []string `yaml:"_before_setup"`
	BuildArtifact []string `yaml:"_build_artifact"`
}

func (cfg StageDependencies) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::GitArtifactLocal::Export::StageDependencies"
}

type GitArtifactRemote struct {
	Url    string                    `yaml:"_url,omitempty"`
	Name   string                    `yaml:"_name,omitempty"`
	As     string                    `yaml:"_as,omitempty"`
	Export []GitArtifactRemoteExport `yaml:"_export"`
}

func (cfg GitArtifactRemote) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::GitArtifactRemote"
}

type GitArtifactRemoteExport struct {
	GitArtifactLocalExport `yaml:",inline"`
	Branch                 string `yaml:"_branch,omitempty"`
	Commit                 string `yaml:"_commit,omitempty"`
}

func (cfg GitArtifactRemoteExport) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::GitArtifactRemote::Export"
}

type ArtifactBaseExport struct {
	Cwd          string   `yaml:"_cwd,omitempty"`
	To           string   `yaml:"_to,omitempty"`
	IncludePaths []string `yaml:"_include_paths"`
	ExcludePaths []string `yaml:"_exclude_paths"`
	Owner        string   `yaml:"_owner,omitempty"`
	Group        string   `yaml:"_group,omitempty"`
}

type Mount struct {
	To   string `yaml:"_to,omitempty"`
	From string `yaml:"_from,omitempty"`
	Type Symbol `yaml:"_type,omitempty"`
}

func (cfg Mount) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::Mount"
}

type Symbol string

func (cfg Symbol) TagYAML() string {
	return "!ruby/symbol"
}
