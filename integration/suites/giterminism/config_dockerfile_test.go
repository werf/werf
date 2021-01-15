package giterminism_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/integration/pkg/utils"
)

var _ = Describe("config dockerfile", func() {
	BeforeEach(CommonBeforeEach)

	Context("contextAddFile", func() {
		type entry struct {
			configDockerfileContextAddFilesGlob string
			context                             string
			contextAddFile                      string
			expectedErrSubstring                string
		}

		DescribeTable("config.dockerfile.allowContextAddFiles",
			func(e entry) {
				fileCreateOrAppend("werf.yaml", fmt.Sprintf(`
image: test
dockerfile: Dockerfile
context: %s
contextAddFile: [%s]
`, e.context, e.contextAddFile))
				gitAddAndCommit("werf.yaml")

				if e.configDockerfileContextAddFilesGlob != "" {
					contentToAppend := fmt.Sprintf(`
config:
  dockerfile:
    allowContextAddFiles: [%s]`, e.configDockerfileContextAddFilesGlob)
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
			Entry("the contextAddFile a/b/c not allowed", entry{
				contextAddFile:       "a/b/c",
				expectedErrSubstring: "the configuration with external dependency found in the werf config: contextAddFile 'a/b/c' not allowed",
			}),
			Entry("config.dockerfile.allowContextAddFiles (a/b/c) covers the contextAddFile a/b/c", entry{
				configDockerfileContextAddFilesGlob: "a/b/c",
				contextAddFile:                      "a/b/c",
			}),
			Entry("config.dockerfile.allowContextAddFiles (/**/*/) covers the contextAddFile a/b/c", entry{
				configDockerfileContextAddFilesGlob: "/**/*/",
				contextAddFile:                      "a/b/c",
			}),
			Entry("config.dockerfile.allowContextAddFiles (/*/) does not cover the contextAddFile /a/b/c", entry{
				configDockerfileContextAddFilesGlob: "/*/",
				contextAddFile:                      "a/b/c",
				expectedErrSubstring:                "the configuration with external dependency found in the werf config: contextAddFile 'a/b/c' not allowed",
			}),
			Entry("config.dockerfile.allowContextAddFiles (a/b/c/) does not cover the contextAddFile a/b/c inside context d", entry{
				configDockerfileContextAddFilesGlob: "a/b/c",
				context:                             "d",
				contextAddFile:                      "a/b/c",
				expectedErrSubstring:                "the configuration with external dependency found in the werf config: contextAddFile 'd/a/b/c' not allowed",
			}),
			Entry("config.dockerfile.allowContextAddFiles (d/a/b/c/) covers the contextAddFile a/b/c inside context d", entry{
				configDockerfileContextAddFilesGlob: "d/a/b/c",
				context:                             "d",
				contextAddFile:                      "a/b/c",
			}),
		)
	})
})
