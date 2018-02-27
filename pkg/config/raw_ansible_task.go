package config

import (
	"fmt"
	"gopkg.in/flant/yaml.v2"
)

type RawAnsibleTask struct {
	Block  []RawAnsibleTask       `yaml:"block,omitempty"`
	Rescue []RawAnsibleTask       `yaml:"rescue,omitempty"`
	Always []RawAnsibleTask       `yaml:"always,omitempty"`
	Fields map[string]interface{} `yaml:",inline"`

	RawAnsible *RawAnsible `yaml:"-"` // parent
}

func (c *RawAnsibleTask) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := ParentStack.Peek().(*RawAnsible); ok {
		c.RawAnsible = parent
	} else if parent, ok := ParentStack.Peek().(*RawAnsibleTask); ok {
		c.RawAnsible = parent.RawAnsible
	}

	ParentStack.Push(c)
	type plain RawAnsibleTask
	err := unmarshal((*plain)(c))
	ParentStack.Pop()
	if err != nil {
		return err
	}

	if !c.BlockDefined() {
		check := false
		for _, supportedModule := range supportedModules() {
			if c.Fields[supportedModule] != nil {
				if check {
					return fmt.Errorf("Invalid ansible task!\n\n%s\n%s", DumpConfigSection(c), DumpConfigDoc(c.RawAnsible.RawDimg.Doc))
				} else {
					check = true
				}
			}
		}

		if !check {
			var supportedModulesString string
			for _, supportedModule := range supportedModules() {
				supportedModulesString += fmt.Sprintf("* %s\n", supportedModule)
			}
			return fmt.Errorf("Unsupported ansible task!\n\nSupported modules list:\n%s\n\n%s\n%s", supportedModulesString, DumpConfigSection(c), DumpConfigDoc(c.RawAnsible.RawDimg.Doc))
		}
	}

	return nil
}

func (c *RawAnsibleTask) BlockDefined() bool {
	return c.Block != nil || c.Rescue != nil || c.Always != nil
}

func supportedModules() []string {
	var modules []string
	// Commands modules
	modules = append(modules, []string{"command", "shell", "raw", "script"}...)
	// Files modules
	modules = append(modules, []string{"blockinfile", "lineinfile", "file", "copy", "acl", "xattr"}...)
	// Utilities modules
	modules = append(modules, []string{"assert", "debug", "set_fact"}...)
	// System modules
	modules = append(modules, []string{"user", "group", "getent"}...)
	// Packaging modules
	modules = append(modules, []string{"apk", "apt", "apt_key", "apt_repository", "yum", "yum_repository"}...)
	return modules
}

func (c *RawAnsibleTask) ToDirective() (*AnsibleTask, error) {
	ansibleTask := &AnsibleTask{}

	marshal, err := yaml.Marshal(c)
	if err != nil {
		return nil, err
	}

	var unmarshal map[string]interface{}
	err = yaml.Unmarshal(marshal, &unmarshal)
	if err != nil {
		return nil, err
	}

	ansibleTask.Config = unmarshal
	ansibleTask.Raw = c

	return ansibleTask, nil
}
