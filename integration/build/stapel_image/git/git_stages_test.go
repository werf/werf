// +build integration

package git

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/flant/werf/integration/utils"
)

var _ = Describe("git stages", func() {
	var testDirPath string
	var fixturesPathParts []string
	var specSteps []stagesSpecStep

	BeforeEach(func() {
		testDirPath = tmpPath()
		fixturesPathParts = []string{"git_stages"}
		specSteps = []stagesSpecStep{}
	})

	AfterEach(func() {
		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"stages", "purge", "-s", ":local", "--force",
		)
	})

	Context("when using image", func() {
		toBuildGitArchiveStageStep := stagesSpecStep{
			byText:                     "First build: gitArchive stage should be built",
			beforeBuildHookFunc:        nil,
			checkResultedFilesChecksum: true,
			expectedOutputMatchers: []types.GomegaMatcher{
				ContainSubstring("gitCache:               <empty>"),
				ContainSubstring("gitLatestPatch:         <empty>"),
				ContainSubstring("Git files will be actualized on stage gitArchive"),
				ContainSubstring("Building stage gitArchive"),
			},
		}

		toResetGitArchiveStageStep := stagesSpecStep{
			byText: "Commit with specific reset message ([werf reset]|[reset werf]): gitArchive stage should be rebuilt",
			beforeBuildHookFunc: func() {
				utils.RunSucceedCommand(
					testDirPath,
					"git",
					"commit", "--allow-empty", "-m", "[werf reset] Reset gitArchive stage",
				)
			},
			checkResultedFilesChecksum: true,
			expectedOutputMatchers: []types.GomegaMatcher{
				ContainSubstring("gitCache:               <empty>"),
				ContainSubstring("gitLatestPatch:         <empty>"),
				ContainSubstring("Git files will be actualized on stage gitArchive"),
				ContainSubstring("Building stage gitArchive"),
			},
		}

		BeforeEach(func() {
			fixturesPathParts = append(fixturesPathParts, "image")
			commonBeforeEach(testDirPath, fixturePath(fixturesPathParts...))
		})

		It("gitArchive stage should be built", func() {
			specSteps = append(specSteps, toBuildGitArchiveStageStep)
			runStagesSpecSteps(testDirPath, specSteps)
		})

		Context("when gitArchive stage is built", func() {
			toBuildGitCacheStageStep := stagesSpecStep{
				byText: "Diff between gitArchive commit and current commit >=1MB: gitCache stage should be built",
				beforeBuildHookFunc: func() {
					createAndCommitFile(testDirPath, "file_1MB", gitCacheSizeStep)
				},
				checkResultedFilesChecksum: true,
				expectedOutputMatchers: []types.GomegaMatcher{
					ContainSubstring("gitLatestPatch:         <empty>"),
					ContainSubstring("Git files will be actualized on stage gitCache"),
					ContainSubstring("Use cache image for stage gitArchive"),
					ContainSubstring("Building stage gitCache"),
				},
			}

			toBuildGitLatestPatchStageStep := stagesSpecStep{
				byText: "Diff between gitArchive commit and current commit <1MB: gitLatestPatch stage should be built",
				beforeBuildHookFunc: func() {
					createAndCommitFile(testDirPath, "file_1023KiB", gitCacheSizeStep-1024)
				},
				checkResultedFilesChecksum: true,
				expectedOutputMatchers: []types.GomegaMatcher{
					ContainSubstring("gitCache:               <empty>"),
					ContainSubstring("Git files will be actualized on stage gitLatestPatch"),
					ContainSubstring("Use cache image for stage gitArchive"),
					ContainSubstring("Building stage gitLatestPatch"),
				},
			}

			BeforeEach(func() {
				specSteps = append(specSteps, toBuildGitArchiveStageStep)
			})

			It("gitCache stage should be built (diff between gitArchive commit and current commit >=1MB)", func() {
				specSteps = append(specSteps, toBuildGitCacheStageStep)
				runStagesSpecSteps(testDirPath, specSteps)
			})

			It("gitLatestPatch stage should be built (diff between gitArchive commit and current commit <1MB)", func() {
				specSteps = append(specSteps, toBuildGitLatestPatchStageStep)
				runStagesSpecSteps(testDirPath, specSteps)
			})

			Context("when gitCache stage is built", func() {
				toRepeatedlyBuildGitCacheStageStep := stagesSpecStep{
					byText: "Diff between gitArchive commit and current commit >=1MB: gitCache stage should be built",
					beforeBuildHookFunc: func() {
						createAndCommitFile(testDirPath, "file2_1MB", gitCacheSizeStep)
					},
					checkResultedFilesChecksum: true,
					expectedOutputMatchers: []types.GomegaMatcher{
						ContainSubstring("gitLatestPatch:         <empty>"),
						ContainSubstring("Git files will be actualized on stage gitCache"),
						ContainSubstring("Use cache image for stage gitArchive"),
						ContainSubstring("Building stage gitCache"),
					},
				}

				toBuildGitLatestPatchStageStep := stagesSpecStep{
					byText: "Diff between gitArchive commit and current commit <1MB: gitLatestPatch stage should be built",
					beforeBuildHookFunc: func() {
						createAndCommitFile(testDirPath, "file_1023KiB", gitCacheSizeStep-1024)
					},
					checkResultedFilesChecksum: true,
					expectedOutputMatchers: []types.GomegaMatcher{
						ContainSubstring("Git files will be actualized on stage gitLatestPatch"),
						ContainSubstring("Use cache image for stage gitCache"),
						ContainSubstring("Use cache image for stage gitArchive"),
						ContainSubstring("Building stage gitLatestPatch"),
					},
				}

				BeforeEach(func() {
					specSteps = append(specSteps, toBuildGitCacheStageStep)
				})

				It("gitArchive stage should be built (commit with specific reset message ([werf reset]|[reset werf]))", func() {
					specSteps = append(specSteps, toResetGitArchiveStageStep)
					runStagesSpecSteps(testDirPath, specSteps)
				})

				It("gitCache stage should be built (diff between gitCache commit and current commit >=1MB)", func() {
					specSteps = append(specSteps, toRepeatedlyBuildGitCacheStageStep)
					runStagesSpecSteps(testDirPath, specSteps)
				})

				It("gitLatestPatch stage should be built (diff between gitCache commit and current commit <1MB)", func() {
					specSteps = append(specSteps, toBuildGitLatestPatchStageStep)
					runStagesSpecSteps(testDirPath, specSteps)
				})
			})

			Context("when gitLatestPatch stage is built", func() {
				BeforeEach(func() {
					specSteps = append(specSteps, toBuildGitLatestPatchStageStep)
				})

				It("gitArchive stage should be built (commit with specific reset message ([werf reset]|[reset werf]))", func() {
					specSteps = append(specSteps, toResetGitArchiveStageStep)
					runStagesSpecSteps(testDirPath, specSteps)
				})

				It("gitCache stage should be built (diff between gitArchive commit and current commit >=1MB)", func() {
					specSteps = append(specSteps, toBuildGitCacheStageStep)
					runStagesSpecSteps(testDirPath, specSteps)
				})

				It("gitLatestPatch stage should be built (diff between gitCache commit and current commit <1MB)", func() {
					specSteps = append(specSteps, toBuildGitLatestPatchStageStep)
					runStagesSpecSteps(testDirPath, specSteps)
				})
			})
		})
	})

	Context("when using artifact", func() {
		toBuildGitArchiveStageStep := stagesSpecStep{
			byText:                     "First build: gitArchive stage should be built",
			beforeBuildHookFunc:        nil,
			checkResultedFilesChecksum: true,
			expectedOutputMatchers: []types.GomegaMatcher{
				ContainSubstring("Git files will be actualized on stage gitArchive"),
				ContainSubstring("Building stage gitArchive"),
			},
		}

		toResetGitArchiveStageStep := stagesSpecStep{
			byText: "Commit with specific reset message ([werf reset]|[reset werf]): gitArchive stage should be rebuilt",
			beforeBuildHookFunc: func() {
				utils.RunSucceedCommand(
					testDirPath,
					"git",
					"commit", "--allow-empty", "-m", "[werf reset] Reset gitArchive stage",
				)
			},
			checkResultedFilesChecksum: true,
			expectedOutputMatchers: []types.GomegaMatcher{
				ContainSubstring("Git files will be actualized on stage gitArchive"),
				ContainSubstring("Building stage gitArchive"),
			},
		}

		toBuildNothingStep := stagesSpecStep{
			byText: "Any changes: nothing should be built",
			beforeBuildHookFunc: func() {
				createAndCommitFile(testDirPath, "file", gitCacheSizeStep)
			},
			checkResultedFilesChecksum: false,
			expectedOutputMatchers: []types.GomegaMatcher{
				Not(ContainSubstring("Building stage ")),
				Not(ContainSubstring("Git files will be actualized on stage ")),
				ContainSubstring("Use cache image for stage gitArchive"),
			},
		}

		BeforeEach(func() {
			fixturesPathParts = append(fixturesPathParts, "artifact")
			commonBeforeEach(testDirPath, fixturePath(fixturesPathParts...))
		})

		It("gitArchive stage should be built", func() {
			specSteps = append(specSteps, toBuildGitArchiveStageStep)
			runStagesSpecSteps(testDirPath, specSteps)
		})

		Context("when gitArchive stage is built", func() {
			BeforeEach(func() {
				specSteps = append(specSteps, toBuildGitArchiveStageStep)
			})

			It("gitArchive stage should be built (commit with specific reset message ([werf reset]|[reset werf]))", func() {
				specSteps = append(specSteps, toResetGitArchiveStageStep)
				runStagesSpecSteps(testDirPath, specSteps)
			})

			It("nothing should be built", func() {
				specSteps = append(specSteps, toBuildNothingStep)
				runStagesSpecSteps(testDirPath, specSteps)
			})
		})
	})
})

type stagesSpecStep struct {
	byText                     string
	beforeBuildHookFunc        func()
	checkResultedFilesChecksum bool
	expectedOutputMatchers     []types.GomegaMatcher
}

func runStagesSpecSteps(testDirPath string, steps []stagesSpecStep) {
	for _, step := range steps {
		By(step.byText)

		if step.beforeBuildHookFunc != nil {
			step.beforeBuildHookFunc()
		}

		out := utils.SucceedCommandOutput(
			testDirPath,
			werfBinPath,
			"build",
		)

		if step.checkResultedFilesChecksum {
			checkResultedFilesChecksum(testDirPath)
		}

		for _, matcher := range step.expectedOutputMatchers {
			Ω(out).Should(matcher)
		}

		out = utils.SucceedCommandOutput(
			testDirPath,
			werfBinPath,
			"build",
		)
		Ω(out).ShouldNot(ContainSubstring("Building stage "))
	}
}

func checkResultedFilesChecksum(testDirPath string) {
	containerTestDirPath := "/source"

	expectedFilesChecksum := filesChecksumCommand(containerTestDirPath)
	resultFilesChecksum := filesChecksumCommand("/app")
	diffCommand := fmt.Sprintf("diff <(%s) <(%s)", resultFilesChecksum, expectedFilesChecksum)

	werfRunArgs := []string{
		"run",
		"--docker-options", fmt.Sprintf("--rm -v %s:%s", testDirPath, containerTestDirPath),
		"--",
		"bash", "-ec", utils.ShelloutPack(diffCommand),
	}
	utils.RunSucceedCommand(
		testDirPath,
		werfBinPath,
		werfRunArgs...,
	)
}

func createAndCommitFile(dirPath string, filename string, contentSize int) {
	newFilePath := filepath.Join(dirPath, filename)
	newFileData := []byte(utils.GetRandomString(contentSize))
	utils.CreateFile(newFilePath, newFileData)

	addAndCommitFile(dirPath, filename, "Add file "+filename)
}

func addAndCommitFile(dirPath string, filename string, commitMsg string) {
	utils.RunSucceedCommand(
		dirPath,
		"git",
		"add", filename,
	)

	utils.RunSucceedCommand(
		dirPath,
		"git",
		"commit", "-m", commitMsg,
	)
}

func filesChecksumCommand(path string) string {
	return fmt.Sprintf(
		"[[ -d %[1]s ]] && find %[1]s -xtype f -not -path '**/.git' -not -path '**/.git/*' -exec bash -c 'printf \"%%s\\n\" \"${@@Q}\"' sh {} + | xargs md5sum | awk '{ print $1 }' | sort | md5sum | awk '{ print $1 }'",
		path,
	)
}
