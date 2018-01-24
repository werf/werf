package ruby_marshal_config

type RubyType interface {
	TagYAML() string
}

type Config struct {
	Dimg      []Dimg      `yaml:"_dimg,omitempty"`
	DimgGroup []DimgGroup `yaml:"_dimg_group,omitempty"`
}

func (cfg Config) TagYAML() string {
	return "!ruby/object:Dapp::Config::Config"
}

type DimgGroup struct {
	Dimg      []Dimg      `yaml:"_dimg,omitempty"`
	DimgGroup []DimgGroup `yaml:"_dimg_group,omitempty"`
}

func (cfg DimgGroup) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::DimgGroup"
}

type Dimg struct {
	Name        string           `yaml:"_name,omitempty"`
	Docker      DockerDimg       `yaml:"_docker,omitempty"`
	Builder     string           `yaml:"_builder,omitempty"`
	Shell       ShellDimg        `yaml:"_shell,omitempty"`
	Chef        Chef             `yaml:"_chef,omitempty"`
	Artifact    []ArtifactExport `yaml:"_artifact,omitempty"`
	GitArtifact GitArtifact      `yaml:"_git_artifact,omitempty"`
	Mount       []Mount          `yaml:"_mount,omitempty"`
}

func (cfg Dimg) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::Dimg"
}

type ArtifactDimg struct {
	Dimg   `yaml:",inline"`
	Docker DockerArtifact `yaml:"_docker,omitempty"`
	Shell  *ShellArtifact `yaml:"_shell,omitempty"`
}

func (cfg ArtifactDimg) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::ArtifactDimg"
}

type DockerDimg struct {
	DockerBase `yaml:",inline"`
	Volume     []string          `yaml:"_volume,omitempty"`
	Expose     []string          `yaml:"_expose,omitempty"`
	Env        map[string]string `yaml:"_env,omitempty"`
	Label      map[string]string `yaml:"_label,omitempty"`
	Cmd        []string          `yaml:"_cmd,omitempty"`
	Onbuild    []string          `yaml:"_onbuild,omitempty"`
	Workdir    string            `yaml:"_workdir,omitempty"`
	User       string            `yaml:"_user,omitempty"`
	Entrypoint []string          `yaml:"_entrypoint,omitempty"`
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
	From             string `yaml:"_from,omitempty"`
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
	Run     []string `yaml:"_run,omitempty"`
}

func (cfg StageCommand) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::Shell::Dimg::StageCommand"
}

type Chef struct {
	Dimod      []string       `yaml:"_dimod,omitempty"`
	Recipe     []string       `yaml:"_recipe,omitempty"`
	Attributes ChefAttributes `yaml:"_attributes,omitempty"`
}

func (cfg Chef) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::Chef"
}

type ChefAttributes map[interface{}]interface{}

func (cfg ChefAttributes) TagYAML() string {
	return "!ruby/hash:Dapp::Dimg::Config::Directive::Chef::Attributes"
}

type ArtifactExport struct {
	ArtifactBaseExport `yaml:",inline"`
	Config             ArtifactDimg `yaml:"_config,omitempty"`
	Before             string       `yaml:"_before,omitempty"`
	After              string       `yaml:"_after,omitempty"`
}

func (cfg ArtifactExport) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::Artifact::Export"
}

type GitArtifact struct {
	Local  []GitArtifactLocal  `yaml:"_local,omitempty"`
	Remote []GitArtifactRemote `yaml:"_remote,omitempty"`
}

func (cfg GitArtifact) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::Dimg::InstanceMethods::GitArtifact"
}

type GitArtifactLocal struct {
	Export []GitArtifactLocalExport `yaml:"_export,omitempty"`
}

func (cfg GitArtifactLocal) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::GitArtifactLocal"
}

type GitArtifactLocalExport struct {
	ArtifactBaseExport `yaml:",inline"`
	As                 string            `yaml:"_as,omitempty"`
	StageDependencies  StageDependencies `yaml:"_stage_dependencies,omitempty"`
}

func (cfg GitArtifactLocalExport) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::GitArtifactLocal::Export"
}

type StageDependencies struct {
	Install       []string `yaml:"_install,omitempty"`
	Setup         []string `yaml:"_setup,omitempty"`
	BeforeSetup   []string `yaml:"_before_setup,omitempty"`
	BuildArtifact []string `yaml:"_build_artifact,omitempty"`
}

func (cfg StageDependencies) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::GitArtifactLocal::Export::StageDependencies"
}

type GitArtifactRemote struct {
	Export []GitArtifactRemoteExport `yaml:"_export,omitempty"`
}

func (cfg GitArtifactRemote) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::GitArtifactRemote"
}

type GitArtifactRemoteExport struct {
	GitArtifactLocalExport `yaml:",inline"`
	Url                    string `yaml:"_url,omitempty"`
	Name                   string `yaml:"_name,omitempty"`
	Branch                 string `yaml:"_branch,omitempty"`
	Commit                 string `yaml:"_commit,omitempty"`
}

func (cfg GitArtifactRemoteExport) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::GitArtifactRemote::Export"
}

type ArtifactBaseExport struct {
	Cwd          string   `yaml:"_cwd,omitempty"`
	To           string   `yaml:"_to,omitempty"`
	IncludePaths []string `yaml:"_include_paths,omitempty"`
	ExcludePaths []string `yaml:"_exclude_paths,omitempty"`
	Owner        string   `yaml:"_owner,omitempty"`
	Group        string   `yaml:"_group,omitempty"`
}

type Mount struct {
	To   string `yaml:"_to,omitempty"`
	From string `yaml:"_from,omitempty"`
	Type string `yaml:"_type,omitempty"`
}

func (cfg Mount) TagYAML() string {
	return "!ruby/object:Dapp::Dimg::Config::Directive::Mount"
}
