package config_test

import (
	"strings"

	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/flant/werf/pkg/testing/utils"
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
		extraArgs:      []string{},
		expectedOutput: strings.ReplaceAll("\n\nproject: none\nconfigVersion: 1.0\n---\nimage: image_a\nfrom: ubuntu\ngit:\n- to: /app\nansible:\n  beforeInstall:  \n  - command: gpg --keyserver hkp://keys.gnupg.net --recv-keys 409B6B1796C275462A1703113804BB82D39DC0E3\n  - get_url:\n      url: https://raw.githubusercontent.com/rvm/rvm/master/binscripts/rvm-installer\n      dest: /tmp/rvm-installer\n  - name: \"Install rvm\"\n    command: bash -e /tmp/rvm-installer\n  - name: \"Install ruby 2.3.4\"\n    raw: bash -lec {{ item | quote }}\n    with_items:\n      - rvm install 2.3.4\n      - rvm use --default 2.3.4\n      - gem install bundler --no-ri --no-rdoc\n      - rvm cleanup all\n---\nimage: image_b\nfrom: ubuntu\nansible:\n  beforeInstall:  \n  - command: gpg --keyserver hkp://keys.gnupg.net --recv-keys 409B6B1796C275462A1703113804BB82D39DC0E3\n  - get_url:\n      url: https://raw.githubusercontent.com/rvm/rvm/master/binscripts/rvm-installer\n      dest: /tmp/rvm-installer\n  - name: \"Install rvm\"\n    command: bash -e /tmp/rvm-installer\n  - name: \"Install ruby 2.3.4\"\n    raw: bash -lec {{ item | quote }}\n    with_items:\n      - rvm install 2.3.4\n      - rvm use --default 2.3.4\n      - gem install bundler --no-ri --no-rdoc\n      - rvm cleanup all\n", "\n", utils.LineBreak),
	}),
	Entry("image_a", renderEntry{
		extraArgs:      []string{"image_a"},
		expectedOutput: strings.ReplaceAll("image: image_a\nfrom: ubuntu\ngit:\n- to: /app\nansible:\n  beforeInstall:  \n  - command: gpg --keyserver hkp://keys.gnupg.net --recv-keys 409B6B1796C275462A1703113804BB82D39DC0E3\n  - get_url:\n      url: https://raw.githubusercontent.com/rvm/rvm/master/binscripts/rvm-installer\n      dest: /tmp/rvm-installer\n  - name: \"Install rvm\"\n    command: bash -e /tmp/rvm-installer\n  - name: \"Install ruby 2.3.4\"\n    raw: bash -lec {{ item | quote }}\n    with_items:\n      - rvm install 2.3.4\n      - rvm use --default 2.3.4\n      - gem install bundler --no-ri --no-rdoc\n      - rvm cleanup all\n", "\n", utils.LineBreak),
	}))
