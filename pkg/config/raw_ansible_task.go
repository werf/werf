package config

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

type rawAnsibleTask struct {
	Block  []rawAnsibleTask       `yaml:"block,omitempty"`
	Rescue []rawAnsibleTask       `yaml:"rescue,omitempty"`
	Always []rawAnsibleTask       `yaml:"always,omitempty"`
	Fields map[string]interface{} `yaml:",inline"`

	rawAnsible *rawAnsible `yaml:"-"` // parent
}

func (c *rawAnsibleTask) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := parentStack.Peek().(*rawAnsible); ok {
		c.rawAnsible = parent
	} else if parent, ok := parentStack.Peek().(*rawAnsibleTask); ok {
		c.rawAnsible = parent.rawAnsible
	}

	parentStack.Push(c)
	type plain rawAnsibleTask
	err := unmarshal((*plain)(c))
	parentStack.Pop()
	if err != nil {
		return err
	}

	if !c.blockDefined() {
		check := false
		for _, supportedModule := range supportedModules() {
			if c.Fields[supportedModule] != nil {
				if check {
					return newDetailedConfigError("invalid ansible task!", c, c.rawAnsible.rawImage.doc)
				} else {
					check = true
				}
			}
		}

		if !check {
			var supportedModulesString string
			for _, supportedModule := range supportedModules() {
				supportedModulesString += fmt.Sprintf("* %s\n", supportedModule)
			}
			return newConfigError(fmt.Sprintf("unsupported ansible task!\n\n%s\nSupported modules list:\n%s\n%s", dumpConfigSection(c), supportedModulesString, dumpConfigDoc(c.rawAnsible.rawImage.doc)))
		}
	}

	return nil
}

func (c *rawAnsibleTask) blockDefined() bool {
	return c.Block != nil || c.Rescue != nil || c.Always != nil
}

func supportedModules() []string {
	var modules []string
	// No Cloud modules
	// No Clustering modules
	// Commands Modules (no expect, psexec and telnet)
	modules = append(modules, []string{"command", "shell", "raw", "script"}...)
	// Crypto Modules
	modules = append(modules, []string{"openssl_certificate", "openssl_csr", "openssl_privatekey", "openssl_publickey"}...)
	// No Databases modules
	// Files Modules (no fetch, patch, replace, synchronize, template, xml
	modules = append(modules, []string{
		"acl",
		"archive",
		"assemble",
		"blockinfile",
		"copy",
		"file",
		"replace",
		"find",
		"ini_file",
		"iso_extract",
		"lineinfile",
		"stat",
		"tempfile",
		"unarchive",
		"xattr",
	}...)
	// No Identity modules
	// No Inventory modules
	// No Messaging modules
	// No Monitoring messages
	// Net Tools Modules (only Basics)
	modules = append(modules, []string{"get_url", "slurp", "uri"}...)
	// No Network modules
	// No Notification modules
	// Packaging/Language Modules
	modules = append(modules, []string{"bower", "bundler", "composer", "cpanm", "easy_install", "gem", "maven_artifact", "npm", "pear", "pip"}...)
	// Packaging/Os Modules
	modules = append(modules, []string{
		"apk",
		"apt",
		"apt_key",
		"apt_repository",
		"apt_rpm",
		"dnf",
		"dpkg_selections",
		"debconf",
		"homebrew",
		"homebrew_cask",
		"homebrew_tap",
		"layman",
		"macports",
		"openbsd_pkg",
		"opkg",
		"package",
		"pacman",
		"pkg5",
		"pkg5_publisher",
		"pkgin",
		"pkgng",
		"pkgutil",
		"portage",
		"portinstall",
		"pulp_repo",
		"redhat_subscription",
		"rhn_channel",
		"rhn_register",
		"rpm_key",
		"slackpkg",
		"sorcery",
		"svr4pkg",
		"swdepot",
		"swupd",
		"urpmi",
		"xbps",
		"yum",
		"yum_repository",
		"zypper",
		"zypper_repository",
	}...)
	// No Remote Management modules
	// No Source Control modules
	// No Storage modules
	// System Modules (only passwd management and locales)
	modules = append(modules, []string{"cron", "user", "group", "getent", "locale_gen", "timezone"}...)
	// Utilities Modules
	modules = append(modules, []string{"meta", "assert", "debug", "fail", "set_fact", "wait_for"}...)
	// No Web Infrastructure modules
	// No Windows modules

	return modules
}

func (c *rawAnsibleTask) toDirective() (*AnsibleTask, error) {
	ansibleTask := &AnsibleTask{}

	marshal, err := yaml.Marshal(c)
	if err != nil {
		return nil, err
	}

	var unmarshal map[string]interface{}
	err = yaml.Unmarshal(marshal, &unmarshal)
	if err != nil {
		return nil, err
	}

	ansibleTask.Config = unmarshal
	ansibleTask.raw = c

	return ansibleTask, nil
}
