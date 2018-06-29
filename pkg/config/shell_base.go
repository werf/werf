package config

type ShellBase struct {
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
