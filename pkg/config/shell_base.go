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

	Raw *RawShell
}

func (c *ShellBase) Validate() error {
	return nil
}
