package config

type Shell interface{}

type ShellBase struct {
	Shell
	BeforeInstall             []string
	Install                   []string
	BeforeSetup               []string
	Setup                     []string
	CacheVersion              string
	BeforeInstallCacheVersion string
	InstallCacheVersion       string
	BeforeSetupCacheVersion   string
	SetupCacheVersion         string

	raw *rawShell
}

func (c *ShellBase) validate() error {
	return nil
}
