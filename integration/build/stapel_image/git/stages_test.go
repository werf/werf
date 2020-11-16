package git_test

import (
	"fmt"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/werf/werf/pkg/testing/utils"
	"github.com/werf/werf/pkg/testing/utils/docker"
)

var _ = Describe("git stages", func() {
	var fixturesPathParts []string
	var specSteps []stagesSpecStep

	BeforeEach(func() {
		if runtime.GOOS == "windows" {
			Skip("skip on windows")
		}

		fixturesPathParts = []string{"git_stages"}
		specSteps = []stagesSpecStep{}
	})

	Context("image", func() {
		toBuildGitArchiveStageStep := stagesSpecStep{
			byText:                     "First build: gitArchive stage should be built",
			beforeBuildHookFunc:        nil,
			checkResultedFilesChecksum: true,
			expectedOutputMatchers: []types.GomegaMatcher{
				Not(ContainSubstring("stage image/gitCache")),
				Not(ContainSubstring("stage image/gitLatestPatch")),
				ContainSubstring("Building stage image/gitArchive"),
			},
		}

		BeforeEach(func() {
			fixturesPathParts = append(fixturesPathParts, "image")
			commonBeforeEach(testDirPath, utils.FixturePath(fixturesPathParts...))
		})

		It("gitArchive stage should be built", func() {
			specSteps = append(specSteps, toBuildGitArchiveStageStep)
			runStagesSpecSteps(testDirPath, specSteps)
		})

		When("gitArchive stage is built", func() {
			toBuildGitCacheStageStep := stagesSpecStep{
				byText: "Diff between gitArchive commit and current commit >=1MB: gitCache stage should be built",
				beforeBuildHookFunc: func() {
					createAndCommitFile(testDirPath, "file_1MB", gitCacheSizeStep)
				},
				checkResultedFilesChecksum: true,
				expectedOutputMatchers: []types.GomegaMatcher{
					Not(ContainSubstring("stage image/gitLatestPatch")),
					ContainSubstring("Use cache image for image/gitArchive"),
					ContainSubstring("Building stage image/gitCache"),
				},
			}

			toBuildGitLatestPatchStageStep := stagesSpecStep{
				byText: "Diff between gitArchive commit and current commit <1MB: gitLatestPatch stage should be built",
				beforeBuildHookFunc: func() {
					createAndCommitFile(testDirPath, "file_1023KiB", gitCacheSizeStep-1024)
				},
				checkResultedFilesChecksum: true,
				expectedOutputMatchers: []types.GomegaMatcher{
					Not(ContainSubstring("stage image/gitCache")),
					ContainSubstring("Use cache image for image/gitArchive"),
					ContainSubstring("Building stage image/gitLatestPatch"),
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

			When("gitCache stage is built", func() {
				toRepeatedlyBuildGitCacheStageStep := stagesSpecStep{
					byText: "Diff between gitArchive commit and current commit >=1MB: gitCache stage should be built",
					beforeBuildHookFunc: func() {
						createAndCommitFile(testDirPath, "file2_1MB", gitCacheSizeStep)
					},
					checkResultedFilesChecksum: true,
					expectedOutputMatchers: []types.GomegaMatcher{
						Not(ContainSubstring("stage image/gitLatestPatch")),
						ContainSubstring("Use cache image for image/gitArchive"),
						ContainSubstring("Building stage image/gitCache"),
					},
				}

				toBuildGitLatestPatchStageStep := stagesSpecStep{
					byText: "Diff between gitArchive commit and current commit <1MB: gitLatestPatch stage should be built",
					beforeBuildHookFunc: func() {
						createAndCommitFile(testDirPath, "file_1023KiB", gitCacheSizeStep-1024)
					},
					checkResultedFilesChecksum: true,
					expectedOutputMatchers: []types.GomegaMatcher{
						ContainSubstring("Use cache image for image/gitCache"),
						ContainSubstring("Use cache image for image/gitArchive"),
						ContainSubstring("Building stage image/gitLatestPatch"),
					},
				}

				BeforeEach(func() {
					specSteps = append(specSteps, toBuildGitCacheStageStep)
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

			When("gitLatestPatch stage is built", func() {
				BeforeEach(func() {
					specSteps = append(specSteps, toBuildGitLatestPatchStageStep)
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

	Context("artifact", func() {
		toBuildGitArchiveStageStep := stagesSpecStep{
			byText:                     "First build: gitArchive stage should be built",
			beforeBuildHookFunc:        nil,
			checkResultedFilesChecksum: true,
			expectedOutputMatchers: []types.GomegaMatcher{
				ContainSubstring("Building stage artifact/gitArchive"),
			},
		}

		toBuildNothingStep := stagesSpecStep{
			byText: "Any changes: nothing should be built",
			beforeBuildHookFunc: func() {
				createAndCommitFile(testDirPath, "file", gitCacheSizeStep)
			},
			checkResultedFilesChecksum: false,
			expectedOutputMatchers: []types.GomegaMatcher{
				Not(ContainSubstring("Building stage")),
				ContainSubstring("Use cache image for artifact/gitArchive"),
			},
		}

		BeforeEach(func() {
			fixturesPathParts = append(fixturesPathParts, "artifact")
			commonBeforeEach(testDirPath, utils.FixturePath(fixturesPathParts...))
		})

		It("gitArchive stage should be built", func() {
			specSteps = append(specSteps, toBuildGitArchiveStageStep)
			runStagesSpecSteps(testDirPath, specSteps)
		})

		When("gitArchive stage is built", func() {
			BeforeEach(func() {
				specSteps = append(specSteps, toBuildGitArchiveStageStep)
			})

			It("nothing should be built", func() {
				specSteps = append(specSteps, toBuildNothingStep)
				runStagesSpecSteps(testDirPath, specSteps)
			})
		})
	})
})

var _ = Describe("user stages", func() {
	var fixturesPathParts []string
	var specSteps []stagesSpecStep

	BeforeEach(func() {
		if runtime.GOOS == "windows" {
			Skip("skip on windows")
		}

		fixturesPathParts = []string{"user_stages"}
		specSteps = []stagesSpecStep{}
	})

	Context("image", func() {
		toBuildGitArchiveStageStep := stagesSpecStep{
			byText:                     "First build: gitArchive stage should be built",
			beforeBuildHookFunc:        nil,
			checkResultedFilesChecksum: true,
			expectedOutputMatchers: []types.GomegaMatcher{
				Not(ContainSubstring("stage image/gitCache")),
				Not(ContainSubstring("stage image/gitLatestPatch")),
				ContainSubstring("Building stage image/gitArchive"),
			},
		}

		toBuildGitCacheStageStep := stagesSpecStep{
			byText: "Diff between gitArchive commit and current commit >=1MB: gitCache stage should be built",
			beforeBuildHookFunc: func() {
				createAndCommitFile(testDirPath, "file_1MB", gitCacheSizeStep)
			},
			checkResultedFilesChecksum: true,
			expectedOutputMatchers: []types.GomegaMatcher{
				Not(ContainSubstring("stage image/gitLatestPatch")),
				ContainSubstring("Use cache image for image/gitArchive"),
				ContainSubstring("Building stage image/gitCache"),
			},
		}

		toBuildGitLatestPatchStageStep := stagesSpecStep{
			byText: "Diff between gitArchive commit and current commit <1MB: gitLatestPatch stage should be built",
			beforeBuildHookFunc: func() {
				createAndCommitFile(testDirPath, "file_1023KiB", gitCacheSizeStep-1024)
			},
			checkResultedFilesChecksum: true,
			expectedOutputMatchers: []types.GomegaMatcher{
				Not(ContainSubstring("stage image/gitCache")),
				ContainSubstring("Use cache image for image/gitArchive"),
				ContainSubstring("Building stage image/gitLatestPatch"),
			},
		}

		BeforeEach(func() {
			fixturesPathParts = append(fixturesPathParts, "image")
		})

		When("stageDependencies are not defined", func() {
			BeforeEach(func() {
				fixturesPathParts = append(fixturesPathParts, "without_stage_dependencies")
				commonBeforeEach(testDirPath, utils.FixturePath(fixturesPathParts...))
			})

			When("gitArchive stage is built", func() {
				userStagesSpecSetFunc := func() {
					It("gitArchive stage should be built (beforeInstall)", func() {
						specSteps = append(specSteps, stagesSpecStep{
							byText: "beforeInstallCacheVersion changed: beforeInstall stage should be built",
							beforeBuildHookFunc: func() {
								stubs.SetEnv("WERF_CONFIG", "werf_beforeInstallCacheVersion.yaml")
							},
							checkResultedFilesChecksum: true,
							expectedOutputMatchers: []types.GomegaMatcher{
								Not(ContainSubstring("stage image/gitCache")),
								Not(ContainSubstring("stage image/gitLatestPatch")),
								ContainSubstring("Building stage image/gitArchive"),
							},
						})
						runStagesSpecSteps(testDirPath, specSteps)
					})

					for _, userStage := range []string{"install", "beforeSetup", "setup"} {
						boundedUserStage := userStage

						itMsg := fmt.Sprintf("%s stage should be built", boundedUserStage)

						It(itMsg, func() {
							specSteps = append(specSteps, stagesSpecStep{
								byText: fmt.Sprintf("%[1]sCacheVersion changed: %[1]s stage should be built", boundedUserStage),
								beforeBuildHookFunc: func() {
									stubs.SetEnv("WERF_CONFIG", fmt.Sprintf("werf_%sCacheVersion.yaml", boundedUserStage))
								},
								checkResultedFilesChecksum: true,
								expectedOutputMatchers: []types.GomegaMatcher{
									Not(ContainSubstring("stage image/gitCache")),
									Not(ContainSubstring("stage image/gitLatestPatch")),
									ContainSubstring(fmt.Sprintf("Building stage image/%s", boundedUserStage)),
								},
							})
							runStagesSpecSteps(testDirPath, specSteps)
						})
					}
				}

				BeforeEach(func() {
					specSteps = append(specSteps, toBuildGitArchiveStageStep)
				})

				userStagesSpecSetFunc()

				When("gitCache stage is built", func() {
					BeforeEach(func() {
						specSteps = append(specSteps, toBuildGitCacheStageStep)
					})

					userStagesSpecSetFunc()
				})

				When("gitLatestPatch stage is built", func() {
					BeforeEach(func() {
						specSteps = append(specSteps, toBuildGitLatestPatchStageStep)
					})

					userStagesSpecSetFunc()
				})
			})
		})

		When("stageDependencies are defined", func() {
			BeforeEach(func() {
				fixturesPathParts = append(fixturesPathParts, "with_stage_dependencies")
				commonBeforeEach(testDirPath, utils.FixturePath(fixturesPathParts...))
			})

			When("gitArchive stage is built", func() {
				userStagesSpecSetFunc := func() {
					for _, userStage := range []string{"install", "beforeSetup", "setup"} {
						boundedUserStage := userStage

						itMsg := fmt.Sprintf("%s stage should be built", boundedUserStage)

						It(itMsg, func() {
							specSteps = append(specSteps, stagesSpecStep{
								byText: fmt.Sprintf("Dependent file changed: %s stage should be built", boundedUserStage),
								beforeBuildHookFunc: func() {
									createAndCommitFile(testDirPath, boundedUserStage, 10)
								},
								checkResultedFilesChecksum: true,
								expectedOutputMatchers: []types.GomegaMatcher{
									ContainSubstring(fmt.Sprintf("Building stage image/%s", boundedUserStage)),
								},
							})
							runStagesSpecSteps(testDirPath, specSteps)
						})
					}
				}

				BeforeEach(func() {
					specSteps = append(specSteps, toBuildGitArchiveStageStep)
				})

				userStagesSpecSetFunc()

				When("gitCache stage is built", func() {
					BeforeEach(func() {
						specSteps = append(specSteps, toBuildGitCacheStageStep)
					})

					userStagesSpecSetFunc()
				})

				When("gitLatestPatch stage is built", func() {
					BeforeEach(func() {
						specSteps = append(specSteps, toBuildGitLatestPatchStageStep)
					})

					userStagesSpecSetFunc()
				})
			})
		})
	})

	When("artifact", func() {
		toBuildGitArchiveStageStep := stagesSpecStep{
			byText:                     "First build: gitArchive stage should be built",
			beforeBuildHookFunc:        nil,
			checkResultedFilesChecksum: true,
			expectedOutputMatchers: []types.GomegaMatcher{
				ContainSubstring("Building stage artifact/gitArchive"),
			},
		}

		toBuildNothingStep := stagesSpecStep{
			byText: "Any changes: nothing should be built",
			beforeBuildHookFunc: func() {
				createAndCommitFile(testDirPath, "file", gitCacheSizeStep)
			},
			checkResultedFilesChecksum: false,
			expectedOutputMatchers: []types.GomegaMatcher{
				Not(ContainSubstring("Building stage")),
				ContainSubstring("Use cache image for artifact/gitArchive"),
			},
		}

		BeforeEach(func() {
			fixturesPathParts = append(fixturesPathParts, "artifact")
		})

		When("stageDependencies are not defined", func() {
			BeforeEach(func() {
				fixturesPathParts = append(fixturesPathParts, "without_stage_dependencies")
				commonBeforeEach(testDirPath, utils.FixturePath(fixturesPathParts...))
			})

			When("gitArchive stage is built", func() {
				toBuildBeforeInstallStageStep := stagesSpecStep{
					byText: fmt.Sprintf("beforeInstallCacheVersion changed: beforeInstall stage should be built"),
					beforeBuildHookFunc: func() {
						stubs.SetEnv("WERF_CONFIG", "werf_beforeInstallCacheVersion.yaml")
					},
					checkResultedFilesChecksum: true,
					expectedOutputMatchers: []types.GomegaMatcher{
						ContainSubstring("Building stage artifact/gitArchive"),
					},
				}

				BeforeEach(func() {
					specSteps = append(specSteps, toBuildGitArchiveStageStep)
				})

				It("gitArchive stage should be built (beforeInstall)", func() {
					specSteps = append(specSteps, toBuildBeforeInstallStageStep)
					runStagesSpecSteps(testDirPath, specSteps)
				})

				for _, userStage := range []string{"install", "beforeSetup", "setup"} {
					boundedUserStage := userStage

					itMsg := fmt.Sprintf("%s stage should be built", boundedUserStage)

					It(itMsg, func() {
						specSteps = append(specSteps, stagesSpecStep{
							byText: fmt.Sprintf("%[1]sCacheVersion changed: %[1]s stage should be built", boundedUserStage),
							beforeBuildHookFunc: func() {
								stubs.SetEnv("WERF_CONFIG", fmt.Sprintf("werf_%sCacheVersion.yaml", boundedUserStage))
							},
							checkResultedFilesChecksum: true,
							expectedOutputMatchers: []types.GomegaMatcher{
								ContainSubstring(fmt.Sprintf("Building stage artifact/%s", boundedUserStage)),
							},
						})
						runStagesSpecSteps(testDirPath, specSteps)
					})
				}

				It("nothing should be built", func() {
					specSteps = append(specSteps, toBuildNothingStep)
					runStagesSpecSteps(testDirPath, specSteps)
				})
			})
		})

		When("stageDependencies are defined", func() {
			BeforeEach(func() {
				fixturesPathParts = append(fixturesPathParts, "with_stage_dependencies")
				commonBeforeEach(testDirPath, utils.FixturePath(fixturesPathParts...))
			})

			When("gitArchive stage is built", func() {
				BeforeEach(func() {
					specSteps = append(specSteps, toBuildGitArchiveStageStep)
				})

				for _, userStage := range []string{"install", "beforeSetup", "setup"} {
					boundedUserStage := userStage

					itMsg := fmt.Sprintf("%s stage should be built", boundedUserStage)

					It(itMsg, func() {
						specSteps = append(specSteps, stagesSpecStep{
							byText: fmt.Sprintf("Dependent file changed: %s stage should be built", boundedUserStage),
							beforeBuildHookFunc: func() {
								createAndCommitFile(testDirPath, boundedUserStage, 10)
							},
							checkResultedFilesChecksum: true,
							expectedOutputMatchers: []types.GomegaMatcher{
								ContainSubstring(fmt.Sprintf("Building stage artifact/%s", boundedUserStage)),
							},
						})
						runStagesSpecSteps(testDirPath, specSteps)
					})
				}

				It("nothing should be built", func() {
					specSteps = append(specSteps, toBuildNothingStep)
					runStagesSpecSteps(testDirPath, specSteps)
				})
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

		out := utils.SucceedCommandOutputString(
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

		out = utils.SucceedCommandOutputString(
			testDirPath,
			werfBinPath,
			"build",
		)
		Ω(out).ShouldNot(ContainSubstring("Building stage"))
	}
}

func checkResultedFilesChecksum(testDirPath string) {
	containerTestDirPath := "/source"

	expectedFilesChecksum := filesChecksumCommand(containerTestDirPath)
	resultFilesChecksum := filesChecksumCommand("/app")
	diffCommand := fmt.Sprintf("diff <(%s) <(%s)", resultFilesChecksum, expectedFilesChecksum)

	docker.RunSucceedContainerCommandWithStapel(
		werfBinPath,
		testDirPath,
		[]string{fmt.Sprintf("-v %s:%s", testDirPath, containerTestDirPath)},
		[]string{diffCommand},
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
