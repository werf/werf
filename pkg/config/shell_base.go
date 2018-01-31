package config

type ShellBase struct {
	BeforeInstall []string
	Install       []string
	BeforeSetup   []string
	Setup         []string

	Raw *RawShell
}

func (c *ShellBase) Validate() error {
	return nil
}
