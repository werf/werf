package giterminism_test

import (
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/integration/pkg/utils"
)

var _ = Describe("config", func() {
	BeforeEach(func() {
		gitInit()
		utils.CopyIn(utils.FixturePath("config"), SuiteData.TestDirPath)
		gitAddAndCommit("werf-giterminism.yaml")
	})

	const minimalConfigContent = `
configVersion: 1
project: none
---
`

	Context("regular files", func() {
		type entry struct {
			allowUncommitted        bool
			addConfig               bool
			commitConfig            bool
			changeConfigAfterCommit bool
			expectedErrSubstring    string
		}

		DescribeTable("config.allowUncommitted",
			func(e entry) {
				var contentToAppend string
				if e.allowUncommitted {
					contentToAppend = `
config:
  allowUncommitted: true`
				} else {
					contentToAppend = `
config:
  allowUncommitted: false`
				}
				fileCreateOrAppend("werf-giterminism.yaml", contentToAppend)
				gitAddAndCommit("werf-giterminism.yaml")

				if e.addConfig {
					fileCreateOrAppend("werf.yaml", minimalConfigContent)
				}

				if e.commitConfig {
					gitAddAndCommit("werf.yaml")
				}

				if e.changeConfigAfterCommit {
					fileCreateOrAppend("werf.yaml", "\n")
				}

				output, err := utils.RunCommand(
					SuiteData.TestDirPath,
					SuiteData.WerfBinPath,
					"config", "render",
				)

				if e.expectedErrSubstring != "" {
					Ω(err).Should(HaveOccurred())
					Ω(string(output)).Should(ContainSubstring(e.expectedErrSubstring))
				} else {
					Ω(err).ShouldNot(HaveOccurred())
				}
			},
			Entry("the config file not found", entry{
				expectedErrSubstring: `unable to read werf config: the file "werf.yaml" not found in the project git repository`,
			}),
			Entry("the config file not committed", entry{
				addConfig:            true,
				expectedErrSubstring: `unable to read werf config: the file "werf.yaml" must be committed`,
			}),
			Entry("the config file committed", entry{
				addConfig:    true,
				commitConfig: true,
			}),
			Entry("the config file changed after commit", entry{
				addConfig:               true,
				commitConfig:            true,
				changeConfigAfterCommit: true,
				expectedErrSubstring:    `unable to read werf config: the file "werf.yaml" must be committed`,
			}),
			Entry("config.allowUncommitted allows not committed config file", entry{
				allowUncommitted: true,
				addConfig:        true,
			}),
			Entry("config.allowUncommitted allows committed file", entry{
				allowUncommitted: true,
				addConfig:        true,
				commitConfig:     true,
			}),
		)
	})

	Context("symlinks", func() {
		configFilePath := "dir/werf.yaml"

		type entry struct {
			allowUncommitted          bool
			addConfigFile             bool
			commitConfigFile          bool
			addSymlinks               map[string]string
			addAndCommitSymlinks      map[string]string
			changeSymlinksAfterCommit map[string]string
			expectedErrSubstring      string
			skipOnWindows             bool
		}

		DescribeTable("config.allowUncommitted",
			func(e entry) {
				if e.skipOnWindows && runtime.GOOS == "windows" {
					Skip("skip on windows")
				}

				var contentToAppend string
				if e.allowUncommitted {
					contentToAppend = `
config:
  allowUncommitted: true`
				} else {
					contentToAppend = `
config:
  allowUncommitted: false`
				}
				fileCreateOrAppend("werf-giterminism.yaml", contentToAppend)
				gitAddAndCommit("werf-giterminism.yaml")

				if e.addConfigFile {
					fileCreateOrAppend(configFilePath, minimalConfigContent)
				}

				if e.commitConfigFile {
					gitAddAndCommit(configFilePath)
				}

				for path, link := range e.addSymlinks {
					symlinkFileCreateOrModify(path, link)
				}

				for path, link := range e.addAndCommitSymlinks {
					symlinkFileCreateOrModifyAndAdd(path, link)
					gitAddAndCommit(path)
				}

				for path, link := range e.changeSymlinksAfterCommit {
					symlinkFileCreateOrModify(path, link)
				}

				output, err := utils.RunCommand(
					SuiteData.TestDirPath,
					SuiteData.WerfBinPath,
					"config", "render",
				)

				if e.expectedErrSubstring != "" {
					Ω(err).Should(HaveOccurred())
					Ω(string(output)).Should(ContainSubstring(e.expectedErrSubstring))
				} else {
					Ω(err).ShouldNot(HaveOccurred())
				}
			},
			Entry("the config file committed: werf.yaml -> a -> dir/werf.yaml", entry{
				commitConfigFile: true,
				addConfigFile:    true,
				addAndCommitSymlinks: map[string]string{
					"werf.yaml": "a",
					"a":         configFilePath,
				},
			}),
			Entry("the config file not committed: werf.yaml -> a -> dir/werf.yaml (not committed)", entry{
				skipOnWindows: true,
				addConfigFile: true,
				addAndCommitSymlinks: map[string]string{
					"werf.yaml": "a",
					"a":         configFilePath,
				},
				expectedErrSubstring: `unable to read werf config: symlink "werf.yaml" check failed: the file "dir/werf.yaml" must be committed`,
			}),
			Entry("the symlink to the config file not committed: werf.yaml (not committed) -> a (not committed) -> dir/werf.yaml", entry{
				addConfigFile:    true,
				commitConfigFile: true,
				addSymlinks: map[string]string{
					"werf.yaml": "a",
					"a":         configFilePath,
				},
				expectedErrSubstring: `unable to read werf config: the file "werf.yaml" must be committed`,
			}),
			Entry("the symlink to the config file not committed: werf.yaml -> a (not committed) -> dir/werf.yaml", entry{
				skipOnWindows:    true,
				addConfigFile:    true,
				commitConfigFile: true,
				addAndCommitSymlinks: map[string]string{
					"werf.yaml": "a",
				},
				addSymlinks: map[string]string{
					"a": configFilePath,
				},
				expectedErrSubstring: ` unable to read werf config: symlink "werf.yaml" check failed: the file "a" must be committed`,
			}),
			Entry("the symlink to the config file changed after commit: werf.yaml (changed) -> a -> dir/werf.yaml", entry{
				addConfigFile:    true,
				commitConfigFile: true,
				addAndCommitSymlinks: map[string]string{
					"werf.yaml": "a",
					"a":         configFilePath,
				},
				changeSymlinksAfterCommit: map[string]string{
					"werf.yaml": configFilePath,
				},
				expectedErrSubstring: `unable to read werf config: the file "werf.yaml" must be committed`,
			}),
			Entry("config.allowUncommitted allows not committed config file", entry{
				skipOnWindows:        true,
				allowUncommitted:     true,
				addConfigFile:        true,
				addAndCommitSymlinks: map[string]string{"werf.yaml": configFilePath},
			}),
			Entry("config.allowUncommitted allows committed config file", entry{
				skipOnWindows:        true,
				allowUncommitted:     true,
				addConfigFile:        true,
				commitConfigFile:     true,
				addAndCommitSymlinks: map[string]string{"werf.yaml": configFilePath},
			}),
			Entry("the broken symlink in fs: werf.yaml -> werf.yaml", entry{
				skipOnWindows:        true,
				allowUncommitted:     true,
				addSymlinks:          map[string]string{"werf.yaml": "werf.yaml"},
				expectedErrSubstring: `unable to read werf config: accepted symlink "werf.yaml" check failed: too many levels of symbolic links`,
			}),
			Entry("the broken symlink in commit: werf.yaml -> werf.yaml", entry{
				addAndCommitSymlinks: map[string]string{"werf.yaml": "werf.yaml"},
				expectedErrSubstring: `unable to read werf config: symlink "werf.yaml" check failed: too many levels of symbolic links`,
			}),
			Entry("the linked file outside the project repository: werf.yaml -> a -> ../../../werf.yaml", entry{
				skipOnWindows:        true,
				addAndCommitSymlinks: map[string]string{"werf.yaml": "a", "a": "../../../werf.yaml"},
				expectedErrSubstring: `unable to read werf config: symlink "werf.yaml" check failed: commit tree entry "../../../werf.yaml" not found in the repository`,
			}),
		)
	})
})
