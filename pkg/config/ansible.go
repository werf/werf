package config

import (
	"context"

	"github.com/werf/werf/pkg/werf/global_warnings"
)

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
	global_warnings.GlobalDeprecationWarningLn(context.Background(), "The `ansible` directive is deprecated and will be removed in v3!")

	return nil
}
