package config

type RawAnsible struct {
	BeforeInstall []RawAnsibleTask `yaml:"beforeInstall"`
	Install       []RawAnsibleTask `yaml:"install"`
	BeforeSetup   []RawAnsibleTask `yaml:"beforeSetup"`
	Setup         []RawAnsibleTask `yaml:"setup"`

	RawDimg *RawDimg `yaml:"-"` // parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *RawAnsible) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := ParentStack.Peek().(*RawDimg); ok {
		c.RawDimg = parent
	}

	ParentStack.Push(c)
	type plain RawAnsible
	err := unmarshal((*plain)(c))
	ParentStack.Pop()
	if err != nil {
		return err
	}

	if err := CheckOverflow(c.UnsupportedAttributes, c, c.RawDimg.Doc); err != nil {
		return err
	}

	return nil
}

func (c *RawAnsible) ToDirective() (ansible *Ansible, err error) {
	ansible = &Ansible{}

	for _, someTask := range c.BeforeInstall {
		if task, err := someTask.ToDirective(); err != nil {
			return nil, err
		} else {
			ansible.BeforeInstall = append(ansible.BeforeInstall, task)
		}
	}

	for _, someTask := range c.Install {
		if task, err := someTask.ToDirective(); err != nil {
			return nil, err
		} else {
			ansible.Install = append(ansible.Install, task)
		}
	}

	for _, someTask := range c.BeforeSetup {
		if task, err := someTask.ToDirective(); err != nil {
			return nil, err
		} else {
			ansible.BeforeSetup = append(ansible.BeforeSetup, task)
		}
	}

	for _, someTask := range c.Setup {
		if task, err := someTask.ToDirective(); err != nil {
			return nil, err
		} else {
			ansible.Setup = append(ansible.Setup, task)
		}
	}

	ansible.Raw = c

	if err := c.ValidateDirective(ansible); err != nil {
		return nil, err
	}

	return ansible, nil
}

func (c *RawAnsible) ValidateDirective(ansible *Ansible) (err error) {
	if err := ansible.Validate(); err != nil {
		return err
	}

	return nil
}
