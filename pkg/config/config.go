package config

import (
	"fmt"
	"strings"
)

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

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

type Git struct {
	ExportBase        `yaml:",inline"`
	As                string            `yaml:"as,omitempty"`
	Url               string            `yaml:"url,omitempty"`
	Branch            string            `yaml:"branch,omitempty"`
	Commit            string            `yaml:"commit,omitempty"`
	StageDependencies StageDependencies `yaml:"stageDependencies,omitempty"`

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

type StageDependencies struct {
	Install       interface{} `yaml:"install,omitempty"`
	Setup         interface{} `yaml:"setup,omitempty"`
	BeforeSetup   interface{} `yaml:"beforeSetup,omitempty"`
	BuildArtifact interface{} `yaml:"buildArtifact,omitempty"`

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

type ArtifactImport struct {
	ExportBase   `yaml:",inline"`
	ArtifactName string `yaml:"artifact,omitempty"`
	Before       string `yaml:"before,omitempty"`
	After        string `yaml:"after,omitempty"`

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
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

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

type Chef struct {
	Cookbook   string         `yaml:"cookbook,omitempty"`
	Recipe     interface{}    `yaml:"recipe,omitempty"`
	Attributes ChefAttributes `yaml:"attributes,omitempty"`

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}
type ChefAttributes map[interface{}]interface{}

type Shell struct {
	BeforeInstall interface{} `yaml:"beforeInstall,omitempty"`
	Install       interface{} `yaml:"install,omitempty"`
	BeforeSetup   interface{} `yaml:"beforeSetup,omitempty"`
	Setup         interface{} `yaml:"setup,omitempty"`
	BuildArtifact interface{} `yaml:"buildArtifact,omitempty"`

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

type Mount struct {
	From string `yaml:"from,omitempty"`
	To   string `yaml:"to,omitempty"`

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *Dimg) directive() string              { return "dimg" }
func (c *Git) directive() string               { return "git" }
func (c *StageDependencies) directive() string { return "git[].stage_dependencies" }
func (c *ArtifactImport) directive() string    { return "artifactImport" }
func (c *Docker) directive() string            { return "docker" }
func (c *Chef) directive() string              { return "chef" }
func (c *Shell) directive() string             { return "shell" }
func (c *Mount) directive() string             { return "mount" }

func (c *Dimg) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Dimg
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := checkOverflow(c.UnsupportedAttributes, c.directive()); err != nil {
		return err
	}

	return nil
}

func (c *StageDependencies) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain StageDependencies
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := checkOverflow(c.UnsupportedAttributes, c.directive()); err != nil {
		return err
	}

	return nil
}

func (c *Git) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Git
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := checkOverflow(c.UnsupportedAttributes, c.directive()); err != nil {
		return err
	}

	return nil
}

func (c *ArtifactImport) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain ArtifactImport
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := checkOverflow(c.UnsupportedAttributes, c.directive()); err != nil {
		return err
	}

	return nil
}

func (c *Docker) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Docker
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := checkOverflow(c.UnsupportedAttributes, c.directive()); err != nil {
		return err
	}

	return nil
}

func (c *Chef) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Chef
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := checkOverflow(c.UnsupportedAttributes, c.directive()); err != nil {
		return err
	}

	return nil
}

func (c *Shell) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Shell
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := checkOverflow(c.UnsupportedAttributes, c.directive()); err != nil {
		return err
	}

	return nil
}

func (c *Mount) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Mount
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := checkOverflow(c.UnsupportedAttributes, c.directive()); err != nil {
		return err
	}

	return nil
}

func checkOverflow(m map[string]interface{}, directive string) error {
	if len(m) > 0 {
		var keys []string
		for k := range m {
			keys = append(keys, strings.Join([]string{directive, k}, "."))
		}
		return fmt.Errorf("Unsupported directives:\n%s", strings.Join(keys, "\n"))
	}
	return nil
}
