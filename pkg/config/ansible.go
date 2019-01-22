package config

type Ansible struct {
	BeforeInstall             []*AnsibleTask
	Install                   []*AnsibleTask
	BeforeSetup               []*AnsibleTask
	Setup                     []*AnsibleTask
	CacheVersion              string
	BeforeInstallCacheVersion string
	InstallCacheVersion       string
	BeforeSetupCacheVersion   string
	SetupCacheVersion         string

	raw *rawAnsible
}

func (c *Ansible) GetDumpConfigSection() string {
	return dumpConfigDoc(c.raw.rawImage.doc)
}

func (c *Ansible) validate() error {
	return nil
}
