package giterminism_test

import (
	"fmt"
	"path/filepath"

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

	Context("Dockerfile", func() {
		type entry struct {
			configDockerfileAllowUncommittedGlob string
			context                              string
			addDockerfile                        bool
			commitDockerfile                     bool
			changeDockerfileAfterCommit          bool
			expectedErrSubstring                 string
		}

		DescribeTable("config.dockerfile.allowUncommitted",
			func(e entry) {
				fileCreateOrAppend("werf.yaml", fmt.Sprintf(`
image: test
dockerfile: Dockerfile
context: %s
`, e.context))
				gitAddAndCommit("werf.yaml")

				if e.configDockerfileAllowUncommittedGlob != "" {
					contentToAppend := fmt.Sprintf(`
config:
  dockerfile:
    allowUncommitted: [%s]`, e.configDockerfileAllowUncommittedGlob)
					fileCreateOrAppend("werf-giterminism.yaml", contentToAppend)
					gitAddAndCommit("werf-giterminism.yaml")
				}

				dockerfileRelPath := filepath.Join(e.context, "Dockerfile")
				if e.addDockerfile {
					fileCreateOrAppend(dockerfileRelPath, fmt.Sprintf(`
FROM alpine
`))
				}

				if e.commitDockerfile {
					gitAddAndCommit(dockerfileRelPath)
				}

				if e.changeDockerfileAfterCommit {
					fileCreateOrAppend(dockerfileRelPath, fmt.Sprintf("\n"))
				}

				output, err := utils.RunCommand(
					SuiteData.TestDirPath,
					SuiteData.WerfBinPath,
					"run", "--skip-build",
				)

				Ω(err).Should(HaveOccurred())
				if e.expectedErrSubstring != "" {
					Ω(string(output)).Should(ContainSubstring(e.expectedErrSubstring))
				} else {
					Ω(string(output)).Should(ContainSubstring("stages required"))
				}
			},
			Entry("the dockerfile not found", entry{
				expectedErrSubstring: "the dockerfile 'Dockerfile' not found in the project git repository",
			}),
			Entry("the dockerfile not committed", entry{
				addDockerfile:        true,
				expectedErrSubstring: "the uncommitted configuration found in the project directory: the dockerfile 'Dockerfile' must be committed",
			}),
			Entry("the dockerfile committed", entry{
				addDockerfile:    true,
				commitDockerfile: true,
			}),
			Entry("the dockerfile committed, the dockerfile changed", entry{
				addDockerfile:               true,
				commitDockerfile:            true,
				changeDockerfileAfterCommit: true,
				expectedErrSubstring:        "the uncommitted configuration found in the project directory: the dockerfile 'Dockerfile' changes must be committed",
			}),
			Entry("config.dockerfile.allowUncommitted (Dockerfile) covers the uncommitted dockerfile 'Dockerfile'", entry{
				configDockerfileAllowUncommittedGlob: "Dockerfile",
				addDockerfile:                        true,
			}),
			Entry("config.dockerfile.allowUncommitted (Dockerfile) covers the committed dockerfile 'Dockerfile'", entry{
				configDockerfileAllowUncommittedGlob: "Dockerfile",
				addDockerfile:                        true,
				commitDockerfile:                     true,
			}),
			Entry("config.dockerfile.allowUncommitted (/*/) covers the dockerfile 'Dockerfile'", entry{
				configDockerfileAllowUncommittedGlob: "/*/",
				addDockerfile:                        true,
			}),
			Entry("config.dockerfile.allowUncommitted (/docker*/) does not cover the dockerfile 'Dockerfile'", entry{
				configDockerfileAllowUncommittedGlob: "/docker*/",
				addDockerfile:                        true,
				expectedErrSubstring:                 "the uncommitted configuration found in the project directory: the dockerfile 'Dockerfile' must be committed",
			}),
			Entry("config.dockerfile.allowContextAddFiles (Dockerfile) does not cover the dockerfile 'Dockerfile' inside context 'context'", entry{
				configDockerfileAllowUncommittedGlob: "Dockerfile",
				context:                              "context",
				addDockerfile:                        true,
				expectedErrSubstring:                 "the uncommitted configuration found in the project directory: the dockerfile 'context/Dockerfile' must be committed",
			}),
			Entry("config.dockerfile.allowContextAddFiles (context/Dockerfile) covers the dockerfile 'Dockerfile' inside context 'context'", entry{
				configDockerfileAllowUncommittedGlob: "context/Dockerfile",
				context:                              "context",
				addDockerfile:                        true,
			}),
		)
	})

	Context(".dockerignore", func() {
		type entry struct {
			configDockerfileAllowUncommittedDockerignoreFilesGlob string
			context                                               string
			addDockerignore                                       bool
			commitDockerignore                                    bool
			changeDockerignoreAfterCommit                         bool
			expectedErrSubstring                                  string
		}

		DescribeTable("config.dockerfile.allowUncommittedDockerignoreFiles",
			func(e entry) {
				fileCreateOrAppend("werf.yaml", fmt.Sprintf(`
image: test
dockerfile: Dockerfile
context: %s
`, e.context))
				gitAddAndCommit("werf.yaml")

				dockerfileRelPath := filepath.Join(e.context, "Dockerfile")
				fileCreateOrAppend(dockerfileRelPath, fmt.Sprintf(`
FROM alpine
`))
				gitAddAndCommit(dockerfileRelPath)

				if e.configDockerfileAllowUncommittedDockerignoreFilesGlob != "" {
					contentToAppend := fmt.Sprintf(`
config:
  dockerfile:
    allowUncommittedDockerignoreFiles: [%s]`, e.configDockerfileAllowUncommittedDockerignoreFilesGlob)
					fileCreateOrAppend("werf-giterminism.yaml", contentToAppend)
					gitAddAndCommit("werf-giterminism.yaml")
				}

				dockerignoreRelPath := filepath.Join(e.context, ".dockerignore")
				if e.addDockerignore {
					fileCreateOrAppend(dockerignoreRelPath, fmt.Sprintf(`
**/*
`))
				}

				if e.commitDockerignore {
					gitAddAndCommit(dockerignoreRelPath)
				}

				if e.changeDockerignoreAfterCommit {
					fileCreateOrAppend(dockerignoreRelPath, fmt.Sprintf("\n"))
				}

				output, err := utils.RunCommand(
					SuiteData.TestDirPath,
					SuiteData.WerfBinPath,
					"run", "--skip-build",
				)

				Ω(err).Should(HaveOccurred())
				if e.expectedErrSubstring != "" {
					Ω(string(output)).Should(ContainSubstring(e.expectedErrSubstring))
				} else {
					Ω(string(output)).Should(ContainSubstring("stages required"))
				}
			},
			Entry("the dockerignore not found", entry{}),
			Entry("the dockerignore not committed", entry{
				addDockerignore:      true,
				expectedErrSubstring: "the uncommitted configuration found in the project directory: the dockerignore file '.dockerignore' must be committed",
			}),
			Entry("the dockerignore committed", entry{
				addDockerignore:    true,
				commitDockerignore: true,
			}),
			Entry("the dockerignore committed, the dockerignore changed", entry{
				addDockerignore:               true,
				commitDockerignore:            true,
				changeDockerignoreAfterCommit: true,
				expectedErrSubstring:          "the uncommitted configuration found in the project directory: the dockerignore file '.dockerignore' changes must be committed",
			}),
			Entry("config.dockerfile.allowUncommittedDockerignoreFiles (.dockerignore) covers the uncommitted dockerignore '.dockerignore'", entry{
				configDockerfileAllowUncommittedDockerignoreFilesGlob: ".dockerignore",
				addDockerignore: true,
			}),
			Entry("config.dockerfile.allowUncommittedDockerignoreFiles (.dockerignore) covers the committed dockerignore '.dockerignore'", entry{
				configDockerfileAllowUncommittedDockerignoreFilesGlob: ".dockerignore",
				addDockerignore:    true,
				commitDockerignore: true,
			}),
			Entry("config.dockerfile.allowUncommittedDockerignoreFiles (/*/) covers the dockerignore '.dockerignore'", entry{
				configDockerfileAllowUncommittedDockerignoreFilesGlob: "/*/",
				addDockerignore: true,
			}),
			Entry("config.dockerfile.allowUncommittedDockerignoreFiles (/docker*/) does not cover the dockerignore '.dockerignore'", entry{
				configDockerfileAllowUncommittedDockerignoreFilesGlob: "/docker*/",
				addDockerignore:      true,
				expectedErrSubstring: "the uncommitted configuration found in the project directory: the dockerignore file '.dockerignore' must be committed",
			}),
			Entry("config.dockerignore.allowContextAddFiles (.dockerignore) does not cover the dockerignore '.dockerignore' inside context 'context'", entry{
				configDockerfileAllowUncommittedDockerignoreFilesGlob: ".dockerignore",
				context:              "context",
				addDockerignore:      true,
				expectedErrSubstring: "the uncommitted configuration found in the project directory: the dockerignore file 'context/.dockerignore' must be committed",
			}),
			Entry("config.dockerignore.allowContextAddFiles (context/.dockerignore) covers the dockerignore '.dockerignore' inside context 'context'", entry{
				configDockerfileAllowUncommittedDockerignoreFilesGlob: "context/.dockerignore",
				context:         "context",
				addDockerignore: true,
			}),
		)
	})
})
