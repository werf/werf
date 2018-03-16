package config

import (
	"fmt"
	"gopkg.in/flant/yaml.v2"
)

type RawAnsibleTask struct {
	Block  []RawAnsibleTask       `yaml:"block,omitempty"`
	Rescue []RawAnsibleTask       `yaml:"rescue,omitempty"`
	Always []RawAnsibleTask       `yaml:"always,omitempty"`
	Fields map[string]interface{} `yaml:",inline"`

	RawAnsible *RawAnsible `yaml:"-"` // parent
}

func (c *RawAnsibleTask) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := ParentStack.Peek().(*RawAnsible); ok {
		c.RawAnsible = parent
	} else if parent, ok := ParentStack.Peek().(*RawAnsibleTask); ok {
		c.RawAnsible = parent.RawAnsible
	}

	ParentStack.Push(c)
	type plain RawAnsibleTask
	err := unmarshal((*plain)(c))
	ParentStack.Pop()
	if err != nil {
		return err
	}

	if !c.BlockDefined() {
		check := false
		for _, supportedModule := range supportedModules() {
			if c.Fields[supportedModule] != nil {
				if check {
					return NewDetailedConfigError("Invalid ansible task!", c, c.RawAnsible.RawDimg.Doc)
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
			return NewConfigError(fmt.Sprintf("Unsupported ansible task!\n\n%s\nSupported modules list:\n%s\n%s", DumpConfigSection(c), supportedModulesString, DumpConfigDoc(c.RawAnsible.RawDimg.Doc)))
		}
	}

	return nil
}

func (c *RawAnsibleTask) BlockDefined() bool {
	return c.Block != nil || c.Rescue != nil || c.Always != nil
}

func supportedModules() []string {
	var modules []string
	// Commands Modules
	modules = append(modules, []string{"command", "shell", "raw", "script"}...)
	// Files Modules
	modules = append(modules, []string{"assemble", "archive", "unarchive", "blockinfile", "lineinfile", "file", "find", "tempfile", "copy", "acl", "xattr", "ini_file", "iso_extract"}...)
	// Net Tools Modules
	modules = append(modules, []string{"get_url", "slurp", "uri"}...)

	// Packaging/Language Modules
	modules = append(modules, []string{"bower", "bundler", "composer", "cpanm", "easy_install", "gem", "maven_artifact", "npm", "pear", "pip"}...)

	// Packagning/Os Modules
	modules = append(modules, []string{
		"apk",
		"apt",
		"apt_key",
		"apt_repository",
		"apt_rpm",
		"dnf",
		"dpkg_selections",
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

	// System Modules
	modules = append(modules, []string{"user", "group", "getent", "locale_gen"}...)
	// Utilities Modules
	modules = append(modules, []string{"assert", "debug", "set_fact", "wait_for"}...)
	// Crypto Modules
	modules = append(modules, []string{"openssl_certificate", "openssl_csr", "openssl_privatekey", "openssl_publickey"}...)

	return modules
}

func (c *RawAnsibleTask) ToDirective() (*AnsibleTask, error) {
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
	ansibleTask.Raw = c

	return ansibleTask, nil
}
