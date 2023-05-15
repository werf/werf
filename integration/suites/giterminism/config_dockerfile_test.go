package giterminism_test

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/utils"
)

var _ = Describe("config dockerfile", func() {
	BeforeEach(CommonBeforeEach)

	const minimalDockerfile = `
FROM alpine
`

	Context("contextAddFiles", func() {
		type entry struct {
			configDockerfileContextAddFilesGlob string
			context                             string
			contextAddFiles                     string
			expectedErrSubstring                string
		}

		DescribeTable("config.dockerfile.allowContextAddFiles",
			func(e entry) {
				fileCreateOrAppend("werf.yaml", fmt.Sprintf(`
image: test
dockerfile: Dockerfile
context: %s
contextAddFiles: [%s]
`, e.context, e.contextAddFiles))
				gitAddAndCommit("werf.yaml")

				if e.configDockerfileContextAddFilesGlob != "" {
					contentToAppend := fmt.Sprintf(`
config:
  dockerfile:
    allowContextAddFiles: ["%s"]`, e.configDockerfileContextAddFilesGlob)
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
				contextAddFiles:      "a/b/c",
				expectedErrSubstring: `the configuration with potential external dependency found in the werf config: contextAddFile "a/b/c" not allowed by giterminism`,
			}),
			Entry("config.dockerfile.allowContextAddFiles (a/b/c) covers the contextAddFile a/b/c", entry{
				configDockerfileContextAddFilesGlob: "a/b/c",
				contextAddFiles:                     "a/b/c",
			}),
			Entry("config.dockerfile.allowContextAddFiles (**/*) covers the contextAddFile a/b/c", entry{
				configDockerfileContextAddFilesGlob: "**/*",
				contextAddFiles:                     "a/b/c",
			}),
			Entry("config.dockerfile.allowContextAddFiles (a/b/c/) does not cover the contextAddFile a/b/c inside context d", entry{
				configDockerfileContextAddFilesGlob: "a/b/c",
				context:                             "d",
				contextAddFiles:                     "a/b/c",
				expectedErrSubstring:                `the configuration with potential external dependency found in the werf config: contextAddFile "d/a/b/c" not allowed by giterminism`,
			}),
			Entry("config.dockerfile.allowContextAddFiles (d/a/b/c/) covers the contextAddFile a/b/c inside context d", entry{
				configDockerfileContextAddFilesGlob: "d/a/b/c",
				context:                             "d",
				contextAddFiles:                     "a/b/c",
			}),
		)
	})

	Context("Dockerfile", func() {
		Context("regular files", func() {
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
    allowUncommitted: ["%s"]`, e.configDockerfileAllowUncommittedGlob)
						fileCreateOrAppend("werf-giterminism.yaml", contentToAppend)
						gitAddAndCommit("werf-giterminism.yaml")
					}

					dockerfileRelPath := filepath.Join(e.context, "Dockerfile")
					if e.addDockerfile {
						fileCreateOrAppend(dockerfileRelPath, fmt.Sprintf(minimalDockerfile))
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
						"run", "--require-built-images",
					)

					Ω(err).Should(HaveOccurred())
					if e.expectedErrSubstring != "" {
						Ω(string(output)).Should(ContainSubstring(e.expectedErrSubstring))
					} else {
						Ω(string(output)).Should(ContainSubstring("stages required"))
					}
				},
				Entry("the dockerfile not found", entry{
					expectedErrSubstring: `unable to read dockerfile "Dockerfile": the file "Dockerfile" not found in the project git repository`,
				}),
				Entry("the dockerfile not tracked", entry{
					addDockerfile:        true,
					expectedErrSubstring: `unable to read dockerfile "Dockerfile": the untracked file "Dockerfile" must be committed`,
				}),
				Entry("the dockerfile committed", entry{
					addDockerfile:    true,
					commitDockerfile: true,
				}),
				Entry("the dockerfile committed, the dockerfile changed", entry{
					addDockerfile:               true,
					commitDockerfile:            true,
					changeDockerfileAfterCommit: true,
					expectedErrSubstring:        `unable to read dockerfile "Dockerfile": the file "Dockerfile" must be committed`,
				}),
				Entry(`config.dockerfile.allowUncommitted (Dockerfile) covers the uncommitted dockerfile "Dockerfile"`, entry{
					configDockerfileAllowUncommittedGlob: "Dockerfile",
					addDockerfile:                        true,
				}),
				Entry(`config.dockerfile.allowUncommitted (Dockerfile) covers the committed dockerfile "Dockerfile"`, entry{
					configDockerfileAllowUncommittedGlob: "Dockerfile",
					addDockerfile:                        true,
					commitDockerfile:                     true,
				}),
				Entry(`config.dockerfile.allowUncommitted (*) covers the dockerfile "Dockerfile"`, entry{
					configDockerfileAllowUncommittedGlob: "*",
					addDockerfile:                        true,
				}),
				Entry(`config.dockerfile.allowUncommitted (docker*) does not cover the untracked dockerfile "Dockerfile"`, entry{
					configDockerfileAllowUncommittedGlob: "docker*",
					addDockerfile:                        true,
					expectedErrSubstring:                 `unable to read dockerfile "Dockerfile": the untracked file "Dockerfile" must be committed`,
				}),
				Entry(`config.dockerfile.allowUncommitted (Dockerfile) does not cover the untracked dockerfile "Dockerfile" inside context "context"`, entry{
					configDockerfileAllowUncommittedGlob: "Dockerfile",
					context:                              "context",
					addDockerfile:                        true,
					expectedErrSubstring:                 `unable to read dockerfile "context/Dockerfile": the untracked file "context/Dockerfile" must be committed`,
				}),
				XEntry(`config.dockerfile.allowUncommitted (context/Dockerfile) covers the untracked dockerfile "Dockerfile" inside context "context"`, entry{
					configDockerfileAllowUncommittedGlob: "context/Dockerfile",
					context:                              "context",
					addDockerfile:                        true,
				}),
			)
		})

		Context("symlinks", func() {
			const dockerfilePath = "dir/Dockerfile"

			type entry struct {
				allowUncommittedGlobs     []string
				addDockerfile             bool
				commitDockerfile          bool
				addSymlinks               map[string]string
				addAndCommitSymlinks      map[string]string
				changeSymlinksAfterCommit map[string]string
				expectedErrSubstring      string
				skipOnWindows             bool
			}

			DescribeTable("config.dockerfile.allowUncommitted",
				func(e entry) {
					if e.skipOnWindows && runtime.GOOS == "windows" {
						Skip("skip on windows")
					}

					{ // werf.yaml
						fileCreateOrAppend("werf.yaml", fmt.Sprintf(`
image: test
dockerfile: Dockerfile
`))
						gitAddAndCommit("werf.yaml")
					}

					{ // werf-giterminism.yaml
						if len(e.allowUncommittedGlobs) != 0 {
							contentToAppend := fmt.Sprintf(`
config:
  dockerfile:
    allowUncommitted: ["%s"]
`, strings.Join(e.allowUncommittedGlobs, `", "`))
							fileCreateOrAppend("werf-giterminism.yaml", contentToAppend)
							gitAddAndCommit("werf-giterminism.yaml")
						}
					}

					{ // Dockerfile
						if e.addDockerfile {
							fileCreateOrAppend(dockerfilePath, minimalDockerfile)
						}

						if e.commitDockerfile {
							gitAddAndCommit(dockerfilePath)
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
					}

					output, err := utils.RunCommand(
						SuiteData.TestDirPath,
						SuiteData.WerfBinPath,
						"run", "--require-built-images",
					)

					Ω(err).Should(HaveOccurred())
					if e.expectedErrSubstring != "" {
						Ω(string(output)).Should(ContainSubstring(e.expectedErrSubstring))
					} else {
						Ω(string(output)).Should(ContainSubstring("stages required"))
					}
				},
				Entry("the dockerfile committed: Dockerfile -> a -> dir/Dockerfile", entry{
					commitDockerfile: true,
					addDockerfile:    true,
					addAndCommitSymlinks: map[string]string{
						"Dockerfile": "a",
						"a":          dockerfilePath,
					},
				}),
				Entry("the dockerfile not tracked: Dockerfile -> a -> dir/Dockerfile (not tracked)", entry{
					skipOnWindows: true,
					addDockerfile: true,
					addAndCommitSymlinks: map[string]string{
						"Dockerfile": "a",
						"a":          dockerfilePath,
					},
					expectedErrSubstring: `unable to read dockerfile "Dockerfile": symlink "Dockerfile" check failed: the untracked file "dir/Dockerfile" must be committed`,
				}),
				Entry("the symlink to the config file changed after commit: Dockerfile (changed) -> a -> dir/Dockerfile", entry{
					commitDockerfile: true,
					addDockerfile:    true,
					addAndCommitSymlinks: map[string]string{
						"Dockerfile": "a",
						"a":          dockerfilePath,
					},
					changeSymlinksAfterCommit: map[string]string{
						"Dockerfile": dockerfilePath,
					},
					expectedErrSubstring: `unable to read dockerfile "Dockerfile": the untracked file "Dockerfile" must be committed`,
				}),
				Entry("config.allowUncommitted (Dockerfile) does not cover uncommitted files", entry{
					skipOnWindows:         true,
					allowUncommittedGlobs: []string{"Dockerfile"},
					addDockerfile:         true,
					addSymlinks: map[string]string{
						"Dockerfile": "a",
						"a":          dockerfilePath,
					},
					expectedErrSubstring: `unable to read dockerfile "Dockerfile": accepted file "Dockerfile" check failed: the link target "a" should be also accepted by giterminism config`,
				}),
				Entry("config.allowUncommitted (Dockerfile, a) does not cover uncommitted files", entry{
					skipOnWindows:         true,
					allowUncommittedGlobs: []string{"Dockerfile", "a"},
					addDockerfile:         true,
					addSymlinks: map[string]string{
						"Dockerfile": "a",
						"a":          dockerfilePath,
					},
					expectedErrSubstring: `unable to read dockerfile "Dockerfile": accepted file "Dockerfile" check failed: the link target "dir/Dockerfile" should be also accepted by giterminism config`,
				}),
				Entry("config.allowUncommitted (Dockerfile, a, dir/Dockerfile) covers uncommitted files", entry{
					skipOnWindows:         true,
					allowUncommittedGlobs: []string{"Dockerfile", "a", dockerfilePath},
					addDockerfile:         true,
					addSymlinks: map[string]string{
						"Dockerfile": "a",
						"a":          dockerfilePath,
					},
				}),
			)
		})
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
				fileCreateOrAppend(dockerfileRelPath, fmt.Sprintf(minimalDockerfile))
				gitAddAndCommit(dockerfileRelPath)

				if e.configDockerfileAllowUncommittedDockerignoreFilesGlob != "" {
					contentToAppend := fmt.Sprintf(`
config:
  dockerfile:
    allowUncommittedDockerignoreFiles: ["%s"]`, e.configDockerfileAllowUncommittedDockerignoreFilesGlob)
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
					"run", "--require-built-images",
				)

				Ω(err).Should(HaveOccurred())
				if e.expectedErrSubstring != "" {
					Ω(string(output)).Should(ContainSubstring(e.expectedErrSubstring))
				} else {
					Ω(string(output)).Should(ContainSubstring("stages required"))
				}
			},
			Entry("the dockerignore not found", entry{}),
			Entry("the dockerignore not tracked", entry{
				addDockerignore:      true,
				expectedErrSubstring: `unable to read dockerignore file ".dockerignore": the untracked file ".dockerignore" must be committed`,
			}),
			Entry("the dockerignore committed", entry{
				addDockerignore:    true,
				commitDockerignore: true,
			}),
			Entry("the dockerignore committed, the dockerignore changed", entry{
				addDockerignore:               true,
				commitDockerignore:            true,
				changeDockerignoreAfterCommit: true,
				expectedErrSubstring:          `unable to read dockerignore file ".dockerignore": the file ".dockerignore" must be committed`,
			}),
			Entry(`config.dockerfile.allowUncommittedDockerignoreFiles (.dockerignore) covers the uncommitted dockerignore ".dockerignore"`, entry{
				configDockerfileAllowUncommittedDockerignoreFilesGlob: ".dockerignore",
				addDockerignore: true,
			}),
			Entry(`config.dockerfile.allowUncommittedDockerignoreFiles (.dockerignore) covers the committed dockerignore ".dockerignore"`, entry{
				configDockerfileAllowUncommittedDockerignoreFilesGlob: ".dockerignore",
				addDockerignore:    true,
				commitDockerignore: true,
			}),
			Entry(`config.dockerfile.allowUncommittedDockerignoreFiles (*) covers the untracked dockerignore ".dockerignore"`, entry{
				configDockerfileAllowUncommittedDockerignoreFilesGlob: "*",
				addDockerignore: true,
			}),
			Entry(`config.dockerfile.allowUncommittedDockerignoreFiles (docker*) does not cover untracked dockerignore ".dockerignore"`, entry{
				configDockerfileAllowUncommittedDockerignoreFilesGlob: "docker*",
				addDockerignore:      true,
				expectedErrSubstring: `unable to read dockerignore file ".dockerignore": the untracked file ".dockerignore" must be committed`,
			}),
			Entry(`config.dockerignore.allowContextAddFiles (.dockerignore) does not cover untracked dockerignore ".dockerignore" inside context "context"`, entry{
				configDockerfileAllowUncommittedDockerignoreFilesGlob: ".dockerignore",
				context:              "context",
				addDockerignore:      true,
				expectedErrSubstring: `unable to read dockerignore file "context/.dockerignore": the untracked file "context/.dockerignore" must be committed`,
			}),
			Entry(`config.dockerignore.allowContextAddFiles (context/.dockerignore) covers untracked dockerignore ".dockerignore" inside context "context"`, entry{
				configDockerfileAllowUncommittedDockerignoreFilesGlob: "context/.dockerignore",
				context:         "context",
				addDockerignore: true,
			}),
		)
	})
})
