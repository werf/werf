package giterminism_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/utils"
)

var _ = Describe("config stapel", func() {
	BeforeEach(CommonBeforeEach)

	Context("fromLatest", func() {
		type entry struct {
			allowStapelFromLatest bool
			expectedErrSubstring  string
		}

		DescribeTable("config.stapel.allowFromLatest",
			func(e entry) {
				fileCreateOrAppend("werf.yaml", `
fromLatest: true
image: test
from: alpine
`)
				gitAddAndCommit("werf.yaml")

				if e.allowStapelFromLatest {
					contentToAppend := `
config:
  stapel:
    allowFromLatest: true`
					fileCreateOrAppend("werf-giterminism.yaml", contentToAppend)
					gitAddAndCommit("werf-giterminism.yaml")
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
			Entry("the from latest directive not allowed", entry{
				expectedErrSubstring: "the configuration with potential external dependency found in the werf config: fromLatest directive not allowed by giterminism",
			}),
			Entry("the from latest directive allowed", entry{
				allowStapelFromLatest: true,
			}),
		)
	})

	Context("git.branch", func() {
		type entry struct {
			allowStapelGitBranch bool
			expectedErrSubstring string
		}

		DescribeTable("config.stapel.git.allowBranch",
			func(e entry) {
				fileCreateOrAppend("werf.yaml", `
image: test
from: alpine
git:
- url: https://github.com/werf/werf.git
  branch: test
  to: /app
`)
				gitAddAndCommit("werf.yaml")

				if e.allowStapelGitBranch {
					contentToAppend := `
config:
  stapel:
    git:
      allowBranch: true`
					fileCreateOrAppend("werf-giterminism.yaml", contentToAppend)
					gitAddAndCommit("werf-giterminism.yaml")
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
			Entry("the remote git branch not allowed", entry{
				expectedErrSubstring: "the configuration with potential external dependency found in the werf config: git branch directive not allowed by giterminism",
			}),
			Entry("the remote git branch allowed", entry{
				allowStapelGitBranch: true,
			}),
		)
	})

	Context("mount build_dir", func() {
		type entry struct {
			allowStapelMountBuildDir bool
			expectedErrSubstring     string
		}

		DescribeTable("config.stapel.mount.allowBuildDir",
			func(e entry) {
				fileCreateOrAppend("werf.yaml", `
image: test
from: alpine
mount:
- from: build_dir
  to: /test
`)
				gitAddAndCommit("werf.yaml")

				if e.allowStapelMountBuildDir {
					contentToAppend := `
config:
  stapel:
    mount:
      allowBuildDir: true`
					fileCreateOrAppend("werf-giterminism.yaml", contentToAppend)
					gitAddAndCommit("werf-giterminism.yaml")
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
			Entry("the build_dir mount not allowed", entry{
				expectedErrSubstring: `the configuration with potential external dependency found in the werf config: "mount { from: build_dir, ... }" not allowed by giterminism`,
			}),
			Entry("the build_dir mount allowed", entry{
				allowStapelMountBuildDir: true,
			}),
		)
	})

	Context("mount fromPath", func() {
		type entry struct {
			allowStapelMountFromPathsGlob string
			fromPath                      string
			expectedErrSubstring          string
		}

		DescribeTable("config.stapel.mount.allowFromPaths",
			func(e entry) {
				fileCreateOrAppend("werf.yaml", fmt.Sprintf(`
image: test
from: alpine
mount:
- fromPath: %s
  to: /test
`, e.fromPath))
				gitAddAndCommit("werf.yaml")

				if e.allowStapelMountFromPathsGlob != "" {
					contentToAppend := fmt.Sprintf(`
config:
  stapel:
    mount:
      allowFromPaths: ["%s"]`, e.allowStapelMountFromPathsGlob)
					fileCreateOrAppend("werf-giterminism.yaml", contentToAppend)
					gitAddAndCommit("werf-giterminism.yaml")
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
			Entry("the from path /a/b/c not allowed", entry{
				fromPath:             "/a/b/c",
				expectedErrSubstring: `the configuration with potential external dependency found in the werf config: "mount { fromPath: /a/b/c, ... }" not allowed by giterminism`,
			}),
			Entry("config.stapel.mount.allowFromPaths (/a/b/c) covers the from path /a/b/c", entry{
				allowStapelMountFromPathsGlob: "/a/b/c",
				fromPath:                      "/a/b/c",
			}),
			Entry("config.stapel.mount.allowFromPaths (**/*) covers the from path /a/b/c", entry{
				allowStapelMountFromPathsGlob: "**/*",
				fromPath:                      "/a/b/c",
			}),
		)
	})
})
