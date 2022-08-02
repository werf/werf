package config

type AnsibleTask struct {
	Config interface{}

	raw *rawAnsibleTask
}

func (c *AnsibleTask) GetDumpConfigSection() string {
	return dumpConfigSection(c.raw)
}
