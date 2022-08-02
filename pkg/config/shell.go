package config

type Shell struct {
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

func (c *Shell) GetDumpConfigSection() string {
	return dumpConfigDoc(c.raw.rawStapelImage.doc)
}

func (c *Shell) validate() error {
	return nil
}
