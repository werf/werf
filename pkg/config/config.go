package config

type Dimg struct {
	Dimg     interface{}      `yaml:"dimg,omitempty"`
	Artifact string           `yaml:"artifact,omitempty"`
	From     string           `yaml:"from,omitempty"`
	Git      []Git            `yaml:"git,omitempty"`
	Shell    Shell            `yaml:"shell,omitempty"`
	Chef     Chef             `yaml:"chef,omitempty"`
	Mount    []Mount          `yaml:"mount,omitempty"`
	Docker   Docker           `yaml:"docker,omitempty"`
	Import   []ArtifactImport `yaml:"import,omitempty"`
}

type Git struct {
	ExportBase        `yaml:",inline"`
	As                string            `yaml:"as,omitempty"`
	Url               string            `yaml:"url,omitempty"`
	Branch            string            `yaml:"branch,omitempty"`
	Commit            string            `yaml:"commit,omitempty"`
	StageDependencies StageDependencies `yaml:"stageDependencies,omitempty"`
}

type StageDependencies struct {
	Install       interface{} `yaml:"install,omitempty"`
	Setup         interface{} `yaml:"setup,omitempty"`
	BeforeSetup   interface{} `yaml:"before_setup,omitempty"`
	BuildArtifact interface{} `yaml:"build_artifact,omitempty"`
}

type ArtifactImport struct {
	ExportBase   `yaml:",inline"`
	ArtifactName string `yaml:"artifact,omitempty"`
	Before       string `yaml:"before,omitempty"`
	After        string `yaml:"after,omitempty"`
}

type ExportBase struct {
	Add          string      `yaml:"add,omitempty"`
	To           string      `yaml:"to,omitempty"`
	IncludePaths interface{} `yaml:"includePaths,omitempty"`
	ExcludePaths interface{} `yaml:"excludePaths,omitempty"`
	Owner        string      `yaml:"owner,omitempty"`
	Group        string      `yaml:"group,omitempty"`
}

type Docker struct {
	Volume     interface{}       `yaml:"VOLUME,omitempty"`
	Expose     interface{}       `yaml:"EXPOSE,omitempty"`
	Env        map[string]string `yaml:"ENV,omitempty"`
	Label      map[string]string `yaml:"LABEL,omitempty"`
	Cmd        interface{}       `yaml:"CMD,omitempty"`
	Onbuild    interface{}       `yaml:"ONBUILD,omitempty"`
	Workdir    string            `yaml:"WORKDIR,omitempty"`
	User       string            `yaml:"USER,omitempty"`
	Entrypoint interface{}       `yaml:"ENTRYPOINT,omitempty"`
}

type Chef struct {
	Cookbook           string         `yaml:"cookbook,omitempty"`
	Recipe             interface{}    `yaml:"recipe,omitempty"`
	AdditionalPackages interface{}    `yaml:"additional_packages,omitempty"`
	Attributes         ChefAttributes `yaml:"attributes,omitempty"`
}
type ChefAttributes map[interface{}]interface{}

type Shell struct {
	BeforeInstall interface{} `yaml:"before_install,omitempty"`
	Install       interface{} `yaml:"install,omitempty"`
	BeforeSetup   interface{} `yaml:"before_setup,omitempty"`
	Setup         interface{} `yaml:"setup,omitempty"`
	BuildArtifact interface{} `yaml:"build_artifact,omitempty"`
}

type Mount struct {
	From string `yaml:"from,omitempty"`
	To   string `yaml:"to,omitempty"`
}
