package config_test

import (
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/testing/utils"
)

type renderEntry struct {
	extraArgs      []string
	expectedOutput string
}

var renderItBody = func(entry renderEntry) {
	testDirPath = utils.FixturePath("render")

	werfArgs := []string{"config", "render"}
	werfArgs = append(werfArgs, entry.extraArgs...)

	output := utils.SucceedCommandOutputString(
		testDirPath,
		werfBinPath,
		werfArgs...,
	)

	Ω(output).Should(Equal(entry.expectedOutput))
}

var _ = DescribeTable("config render", renderItBody,
	Entry("all", renderEntry{
		extraArgs: []string{},
		expectedOutput: `

project: none
configVersion: 1.0
---
image: image_a
from: ubuntu
git:
- to: /app
ansible:
  beforeInstall:  
  - command: gpg --keyserver hkp://keys.gnupg.net --recv-keys 409B6B1796C275462A1703113804BB82D39DC0E3
  - get_url:
      url: https://raw.githubusercontent.com/rvm/rvm/master/binscripts/rvm-installer
      dest: /tmp/rvm-installer
  - name: "Install rvm"
    command: bash -e /tmp/rvm-installer
  - name: "Install ruby 2.3.4"
    raw: bash -lec {{ item | quote }}
    with_items:
      - rvm install 2.3.4
      - rvm use --default 2.3.4
      - gem install bundler --no-ri --no-rdoc
      - rvm cleanup all
---
image: image_b
from: ubuntu
ansible:
  beforeInstall:  
  - command: gpg --keyserver hkp://keys.gnupg.net --recv-keys 409B6B1796C275462A1703113804BB82D39DC0E3
  - get_url:
      url: https://raw.githubusercontent.com/rvm/rvm/master/binscripts/rvm-installer
      dest: /tmp/rvm-installer
  - name: "Install rvm"
    command: bash -e /tmp/rvm-installer
  - name: "Install ruby 2.3.4"
    raw: bash -lec {{ item | quote }}
    with_items:
      - rvm install 2.3.4
      - rvm use --default 2.3.4
      - gem install bundler --no-ri --no-rdoc
      - rvm cleanup all
`,
	}),
	Entry("image_a", renderEntry{
		extraArgs: []string{"image_a"},
		expectedOutput: `image: image_a
from: ubuntu
git:
- to: /app
ansible:
  beforeInstall:  
  - command: gpg --keyserver hkp://keys.gnupg.net --recv-keys 409B6B1796C275462A1703113804BB82D39DC0E3
  - get_url:
      url: https://raw.githubusercontent.com/rvm/rvm/master/binscripts/rvm-installer
      dest: /tmp/rvm-installer
  - name: "Install rvm"
    command: bash -e /tmp/rvm-installer
  - name: "Install ruby 2.3.4"
    raw: bash -lec {{ item | quote }}
    with_items:
      - rvm install 2.3.4
      - rvm use --default 2.3.4
      - gem install bundler --no-ri --no-rdoc
      - rvm cleanup all
`}))
