package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
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
			return fmt.Errorf("Unsupported ansible task!\n\n%s\n%s", DumpConfigSection(c), DumpConfigDoc(c.RawAnsible.RawDimg.Doc))
		}
	}

	return nil
}

func (c *RawAnsibleTask) BlockDefined() bool {
	return c.Block != nil || c.Rescue != nil || c.Always != nil
}

func supportedModules() []string {
	return []string{"command", "shell", "copy", "debug"}
}

func (c *RawAnsibleTask) ToDirective() (interface{}, error) {
	marshal, err := yaml.Marshal(c)
	if err != nil {
		return nil, err
	}

	var unmarshal map[string]interface{}
	err = yaml.Unmarshal(marshal, &unmarshal)
	if err != nil {
		return nil, err
	}

	return unmarshal, nil
}
