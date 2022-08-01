package giterminism_test

import (
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/utils"
)

var _ = Describe("config", func() {
	const minimalConfigContent = `
configVersion: 1
project: none
---
`

	for _, value := range []string{"", "dir/custom_project_dir"} {
		projectRelPath := value

		relativeToWorkTreeDir := func(path string) string {
			return filepath.Join(projectRelPath, path)
		}

		werfGiterminismRelPath := relativeToWorkTreeDir("werf-giterminism.yaml")
		werfConfigRelPath := relativeToWorkTreeDir("werf.yaml")

		type symlinkEntry struct {
			allowUncommitted          bool
			addConfigFile             bool
			commitConfigFile          bool
			addSymlinks               map[string]string
			addAndCommitSymlinks      map[string]string
			changeSymlinksAfterCommit map[string]string
			expectedErrSubstring      string
			skipOnWindows             bool
		}

		symlinkBodyFunc := func(configFilePath string) func(e symlinkEntry) {
			return func(e symlinkEntry) {
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

				fileCreateOrAppend(werfGiterminismRelPath, contentToAppend)
				gitAddAndCommit(werfGiterminismRelPath)

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
					filepath.Join(SuiteData.TestDirPath, projectRelPath),
					SuiteData.WerfBinPath,
					"config", "render",
				)

				if e.expectedErrSubstring != "" {
					Ω(err).Should(HaveOccurred())
					Ω(string(output)).Should(ContainSubstring(e.expectedErrSubstring))
				} else {
					Ω(err).ShouldNot(HaveOccurred())
				}
			}
		}

		runCommonTests := func() {
			BeforeEach(func() {
				gitInit()
				utils.CopyIn(utils.FixturePath("config"), filepath.Join(SuiteData.TestDirPath, projectRelPath))
				gitAddAndCommit(werfGiterminismRelPath)
			})

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

						fileCreateOrAppend(werfGiterminismRelPath, contentToAppend)
						gitAddAndCommit(werfGiterminismRelPath)

						if e.addConfig {
							fileCreateOrAppend(werfConfigRelPath, minimalConfigContent)
						}

						if e.commitConfig {
							gitAddAndCommit(werfConfigRelPath)
						}

						if e.changeConfigAfterCommit {
							fileCreateOrAppend(werfConfigRelPath, "\n")
						}

						output, err := utils.RunCommand(
							filepath.Join(SuiteData.TestDirPath, projectRelPath),
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
					Entry("the config file not tracked", entry{
						addConfig:            true,
						expectedErrSubstring: `unable to read werf config: the untracked file "werf.yaml" must be committed`,
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
					Entry("config.allowUncommitted allows not tracked config file", entry{
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
				configFilePath := relativeToWorkTreeDir("dir/werf.yaml")
				aFilePath := relativeToWorkTreeDir("a")

				DescribeTable("config.allowUncommitted",
					symlinkBodyFunc(configFilePath),
					Entry("the config file committed: werf.yaml -> a -> dir/werf.yaml", symlinkEntry{
						commitConfigFile: true,
						addConfigFile:    true,
						addAndCommitSymlinks: map[string]string{
							werfConfigRelPath: getLinkTo(werfConfigRelPath, aFilePath),
							aFilePath:         getLinkTo(aFilePath, configFilePath),
						},
					}),
					Entry("the config file not tracked: werf.yaml -> a -> dir/werf.yaml (not tracked)", symlinkEntry{
						skipOnWindows: true,
						addConfigFile: true,
						addAndCommitSymlinks: map[string]string{
							werfConfigRelPath: getLinkTo(werfConfigRelPath, aFilePath),
							aFilePath:         getLinkTo(aFilePath, configFilePath),
						},
						expectedErrSubstring: `unable to read werf config: symlink "werf.yaml" check failed: the untracked file "dir/werf.yaml" must be committed`,
					}),
					Entry("the symlink to the config file not tracked: werf.yaml (not tracked) -> a (not tracked) -> dir/werf.yaml", symlinkEntry{
						addConfigFile:    true,
						commitConfigFile: true,
						addSymlinks: map[string]string{
							werfConfigRelPath: getLinkTo(werfConfigRelPath, aFilePath),
							aFilePath:         getLinkTo(aFilePath, configFilePath),
						},
						expectedErrSubstring: `unable to read werf config: the untracked file "werf.yaml" must be committed`,
					}),
					Entry("the symlink to the config file not tracked: werf.yaml -> a (not tracked) -> dir/werf.yaml", symlinkEntry{
						skipOnWindows:    true,
						addConfigFile:    true,
						commitConfigFile: true,
						addAndCommitSymlinks: map[string]string{
							werfConfigRelPath: getLinkTo(werfConfigRelPath, aFilePath),
						},
						addSymlinks: map[string]string{
							aFilePath: configFilePath,
						},
						expectedErrSubstring: ` unable to read werf config: symlink "werf.yaml" check failed: the untracked file "a" must be committed`,
					}),
					Entry("the symlink to the config file changed after commit: werf.yaml (changed) -> a -> dir/werf.yaml", symlinkEntry{
						addConfigFile:    true,
						commitConfigFile: true,
						addAndCommitSymlinks: map[string]string{
							werfConfigRelPath: getLinkTo(werfConfigRelPath, aFilePath),
							aFilePath:         getLinkTo(aFilePath, configFilePath),
						},
						changeSymlinksAfterCommit: map[string]string{
							werfConfigRelPath: getLinkTo(werfConfigRelPath, configFilePath),
						},
						expectedErrSubstring: `unable to read werf config: the untracked file "werf.yaml" must be committed`,
					}),
					Entry("config.allowUncommitted allows not tracked config file", symlinkEntry{
						skipOnWindows:        true,
						allowUncommitted:     true,
						addConfigFile:        true,
						addAndCommitSymlinks: map[string]string{werfConfigRelPath: getLinkTo(werfConfigRelPath, configFilePath)},
					}),
					Entry("config.allowUncommitted allows committed config file", symlinkEntry{
						skipOnWindows:        true,
						allowUncommitted:     true,
						addConfigFile:        true,
						commitConfigFile:     true,
						addAndCommitSymlinks: map[string]string{werfConfigRelPath: getLinkTo(werfConfigRelPath, configFilePath)},
					}),
					Entry("the broken symlink in fs: werf.yaml -> werf.yaml", symlinkEntry{
						skipOnWindows:        true,
						allowUncommitted:     true,
						addSymlinks:          map[string]string{werfConfigRelPath: getLinkTo(werfConfigRelPath, werfConfigRelPath)},
						expectedErrSubstring: `unable to read werf config: accepted file "werf.yaml" check failed: too many levels of symbolic links`,
					}),
					Entry("the broken symlink in commit: werf.yaml -> werf.yaml", symlinkEntry{
						addAndCommitSymlinks: map[string]string{werfConfigRelPath: getLinkTo(werfConfigRelPath, werfConfigRelPath)},
						expectedErrSubstring: `unable to read werf config: symlink "werf.yaml" check failed: too many levels of symbolic links`,
					}),
					Entry("the linked file outside the project repository: werf.yaml -> a -> ../../../werf.yaml", symlinkEntry{
						skipOnWindows:        true,
						addAndCommitSymlinks: map[string]string{werfConfigRelPath: getLinkTo(werfConfigRelPath, aFilePath), aFilePath: "../../../werf.yaml"},
						expectedErrSubstring: `unable to read werf config: symlink "werf.yaml" check failed: commit tree entry "../../../werf.yaml" not found in the repository`,
					}),
				)
			})
		}

		if projectRelPath == "" {
			{
				runCommonTests()
			}
		} else {
			Context("custom project directory", func() {
				runCommonTests()

				It("the symlink to the config file outside project directory: werf.yaml -> ../../werf.yaml", func() {
					symlinkBodyFunc("werf.yaml")(symlinkEntry{
						addConfigFile:    true,
						commitConfigFile: true,
						addAndCommitSymlinks: map[string]string{
							relativeToWorkTreeDir("werf.yaml"): getLinkTo(relativeToWorkTreeDir("werf.yaml"), "werf.yaml"),
						},
					})
				})

				It("config.allowUncommitted allows the symlink to the config file outside project directory: werf.yaml -> ../../werf.yaml", func() {
					symlinkBodyFunc("werf.yaml")(symlinkEntry{
						skipOnWindows:    true,
						allowUncommitted: true,
						addConfigFile:    true,
						addSymlinks: map[string]string{
							relativeToWorkTreeDir("werf.yaml"): getLinkTo(relativeToWorkTreeDir("werf.yaml"), "werf.yaml"),
						},
					})
				})
			})
		}
	}
})
