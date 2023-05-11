package git_test

import (
	"fmt"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/werf/werf/test/pkg/utils"
	"github.com/werf/werf/test/pkg/utils/docker"
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
			commonBeforeEach(utils.FixturePath(fixturesPathParts...))
		})

		It("gitArchive stage should be built", func() {
			specSteps = append(specSteps, toBuildGitArchiveStageStep)
			runStagesSpecSteps(specSteps)
		})

		When("gitArchive stage is built", func() {
			toBuildGitCacheStageStep := stagesSpecStep{
				byText: "Diff between gitArchive commit and current commit >=1MB: gitCache stage should be built",
				beforeBuildHookFunc: func() {
					createAndCommitFile(SuiteData.TestDirPath, "file_1MB", gitCacheSizeStep)
				},
				checkResultedFilesChecksum: true,
				expectedOutputMatchers: []types.GomegaMatcher{
					Not(ContainSubstring("stage image/gitLatestPatch")),
					ContainSubstring("Use previously built image for image/gitArchive"),
					ContainSubstring("Building stage image/gitCache"),
				},
			}

			toBuildGitLatestPatchStageStep := stagesSpecStep{
				byText: "Diff between gitArchive commit and current commit <1MB: gitLatestPatch stage should be built",
				beforeBuildHookFunc: func() {
					createAndCommitFile(SuiteData.TestDirPath, "file_1023KiB", gitCacheSizeStep-1024)
				},
				checkResultedFilesChecksum: true,
				expectedOutputMatchers: []types.GomegaMatcher{
					Not(ContainSubstring("stage image/gitCache")),
					ContainSubstring("Use previously built image for image/gitArchive"),
					ContainSubstring("Building stage image/gitLatestPatch"),
				},
			}

			BeforeEach(func() {
				specSteps = append(specSteps, toBuildGitArchiveStageStep)
			})

			It("gitCache stage should be built (diff between gitArchive commit and current commit >=1MB)", func() {
				specSteps = append(specSteps, toBuildGitCacheStageStep)
				runStagesSpecSteps(specSteps)
			})

			It("gitLatestPatch stage should be built (diff between gitArchive commit and current commit <1MB)", func() {
				specSteps = append(specSteps, toBuildGitLatestPatchStageStep)
				runStagesSpecSteps(specSteps)
			})

			When("gitCache stage is built", func() {
				toRepeatedlyBuildGitCacheStageStep := stagesSpecStep{
					byText: "Diff between gitArchive commit and current commit >=1MB: gitCache stage should be built",
					beforeBuildHookFunc: func() {
						createAndCommitFile(SuiteData.TestDirPath, "file2_1MB", gitCacheSizeStep)
					},
					checkResultedFilesChecksum: true,
					expectedOutputMatchers: []types.GomegaMatcher{
						Not(ContainSubstring("stage image/gitLatestPatch")),
						ContainSubstring("Use previously built image for image/gitArchive"),
						ContainSubstring("Building stage image/gitCache"),
					},
				}

				toBuildGitLatestPatchStageStep := stagesSpecStep{
					byText: "Diff between gitArchive commit and current commit <1MB: gitLatestPatch stage should be built",
					beforeBuildHookFunc: func() {
						createAndCommitFile(SuiteData.TestDirPath, "file_1023KiB", gitCacheSizeStep-1024)
					},
					checkResultedFilesChecksum: true,
					expectedOutputMatchers: []types.GomegaMatcher{
						ContainSubstring("Use previously built image for image/gitCache"),
						ContainSubstring("Use previously built image for image/gitArchive"),
						ContainSubstring("Building stage image/gitLatestPatch"),
					},
				}

				BeforeEach(func() {
					specSteps = append(specSteps, toBuildGitCacheStageStep)
				})

				It("gitCache stage should be built (diff between gitCache commit and current commit >=1MB)", func() {
					specSteps = append(specSteps, toRepeatedlyBuildGitCacheStageStep)
					runStagesSpecSteps(specSteps)
				})

				It("gitLatestPatch stage should be built (diff between gitCache commit and current commit <1MB)", func() {
					specSteps = append(specSteps, toBuildGitLatestPatchStageStep)
					runStagesSpecSteps(specSteps)
				})
			})

			When("gitLatestPatch stage is built", func() {
				BeforeEach(func() {
					specSteps = append(specSteps, toBuildGitLatestPatchStageStep)
				})

				It("gitCache stage should be built (diff between gitArchive commit and current commit >=1MB)", func() {
					specSteps = append(specSteps, toBuildGitCacheStageStep)
					runStagesSpecSteps(specSteps)
				})

				It("gitLatestPatch stage should be built (diff between gitCache commit and current commit <1MB)", func() {
					specSteps = append(specSteps, toBuildGitLatestPatchStageStep)
					runStagesSpecSteps(specSteps)
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
				createAndCommitFile(SuiteData.TestDirPath, "file", gitCacheSizeStep)
			},
			checkResultedFilesChecksum: false,
			expectedOutputMatchers: []types.GomegaMatcher{
				Not(ContainSubstring("Building stage")),
				ContainSubstring("Use previously built image for artifact/gitArchive"),
			},
		}

		BeforeEach(func() {
			fixturesPathParts = append(fixturesPathParts, "artifact")
			commonBeforeEach(utils.FixturePath(fixturesPathParts...))
		})

		It("gitArchive stage should be built", func() {
			specSteps = append(specSteps, toBuildGitArchiveStageStep)
			runStagesSpecSteps(specSteps)
		})

		When("gitArchive stage is built", func() {
			BeforeEach(func() {
				specSteps = append(specSteps, toBuildGitArchiveStageStep)
			})

			It("nothing should be built", func() {
				specSteps = append(specSteps, toBuildNothingStep)
				runStagesSpecSteps(specSteps)
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
				createAndCommitFile(SuiteData.TestDirPath, "file_1MB", gitCacheSizeStep)
			},
			checkResultedFilesChecksum: true,
			expectedOutputMatchers: []types.GomegaMatcher{
				Not(ContainSubstring("stage image/gitLatestPatch")),
				ContainSubstring("Use previously built image for image/gitArchive"),
				ContainSubstring("Building stage image/gitCache"),
			},
		}

		toBuildGitLatestPatchStageStep := stagesSpecStep{
			byText: "Diff between gitArchive commit and current commit <1MB: gitLatestPatch stage should be built",
			beforeBuildHookFunc: func() {
				createAndCommitFile(SuiteData.TestDirPath, "file_1023KiB", gitCacheSizeStep-1024)
			},
			checkResultedFilesChecksum: true,
			expectedOutputMatchers: []types.GomegaMatcher{
				Not(ContainSubstring("stage image/gitCache")),
				ContainSubstring("Use previously built image for image/gitArchive"),
				ContainSubstring("Building stage image/gitLatestPatch"),
			},
		}

		BeforeEach(func() {
			fixturesPathParts = append(fixturesPathParts, "image")
		})

		When("stageDependencies are not defined", func() {
			BeforeEach(func() {
				fixturesPathParts = append(fixturesPathParts, "without_stage_dependencies")
				commonBeforeEach(utils.FixturePath(fixturesPathParts...))
			})

			When("gitArchive stage is built", func() {
				userStagesSpecSetFunc := func() {
					It("gitArchive stage should be built (beforeInstall)", func() {
						specSteps = append(specSteps, stagesSpecStep{
							byText: "beforeInstallCacheVersion changed: beforeInstall stage should be built",
							beforeBuildHookFunc: func() {
								SuiteData.Stubs.SetEnv("WERF_CONFIG", "werf_beforeInstallCacheVersion.yaml")
							},
							checkResultedFilesChecksum: true,
							expectedOutputMatchers: []types.GomegaMatcher{
								Not(ContainSubstring("stage image/gitCache")),
								Not(ContainSubstring("stage image/gitLatestPatch")),
								ContainSubstring("Building stage image/gitArchive"),
							},
						})
						runStagesSpecSteps(specSteps)
					})

					for _, userStage := range []string{"install", "beforeSetup", "setup"} {
						boundedUserStage := userStage

						itMsg := fmt.Sprintf("%s stage should be built", boundedUserStage)

						It(itMsg, func() {
							specSteps = append(specSteps, stagesSpecStep{
								byText: fmt.Sprintf("%[1]sCacheVersion changed: %[1]s stage should be built", boundedUserStage),
								beforeBuildHookFunc: func() {
									SuiteData.Stubs.SetEnv("WERF_CONFIG", fmt.Sprintf("werf_%sCacheVersion.yaml", boundedUserStage))
								},
								checkResultedFilesChecksum: true,
								expectedOutputMatchers: []types.GomegaMatcher{
									Not(ContainSubstring("stage image/gitCache")),
									Not(ContainSubstring("stage image/gitLatestPatch")),
									ContainSubstring(fmt.Sprintf("Building stage image/%s", boundedUserStage)),
								},
							})
							runStagesSpecSteps(specSteps)
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
				commonBeforeEach(utils.FixturePath(fixturesPathParts...))
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
									createAndCommitFile(SuiteData.TestDirPath, boundedUserStage, 10)
								},
								checkResultedFilesChecksum: true,
								expectedOutputMatchers: []types.GomegaMatcher{
									ContainSubstring(fmt.Sprintf("Building stage image/%s", boundedUserStage)),
								},
							})
							runStagesSpecSteps(specSteps)
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
				createAndCommitFile(SuiteData.TestDirPath, "file", gitCacheSizeStep)
			},
			checkResultedFilesChecksum: false,
			expectedOutputMatchers: []types.GomegaMatcher{
				Not(ContainSubstring("Building stage")),
				ContainSubstring("Use previously built image for artifact/gitArchive"),
			},
		}

		BeforeEach(func() {
			fixturesPathParts = append(fixturesPathParts, "artifact")
		})

		When("stageDependencies are not defined", func() {
			BeforeEach(func() {
				fixturesPathParts = append(fixturesPathParts, "without_stage_dependencies")
				commonBeforeEach(utils.FixturePath(fixturesPathParts...))
			})

			When("gitArchive stage is built", func() {
				toBuildBeforeInstallStageStep := stagesSpecStep{
					byText: fmt.Sprintf("beforeInstallCacheVersion changed: beforeInstall stage should be built"),
					beforeBuildHookFunc: func() {
						SuiteData.Stubs.SetEnv("WERF_CONFIG", "werf_beforeInstallCacheVersion.yaml")
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
					runStagesSpecSteps(specSteps)
				})

				for _, userStage := range []string{"install", "beforeSetup", "setup"} {
					boundedUserStage := userStage

					itMsg := fmt.Sprintf("%s stage should be built", boundedUserStage)

					It(itMsg, func() {
						specSteps = append(specSteps, stagesSpecStep{
							byText: fmt.Sprintf("%[1]sCacheVersion changed: %[1]s stage should be built", boundedUserStage),
							beforeBuildHookFunc: func() {
								SuiteData.Stubs.SetEnv("WERF_CONFIG", fmt.Sprintf("werf_%sCacheVersion.yaml", boundedUserStage))
							},
							checkResultedFilesChecksum: true,
							expectedOutputMatchers: []types.GomegaMatcher{
								ContainSubstring(fmt.Sprintf("Building stage artifact/%s", boundedUserStage)),
							},
						})
						runStagesSpecSteps(specSteps)
					})
				}

				It("nothing should be built", func() {
					specSteps = append(specSteps, toBuildNothingStep)
					runStagesSpecSteps(specSteps)
				})
			})
		})

		When("stageDependencies are defined", func() {
			BeforeEach(func() {
				fixturesPathParts = append(fixturesPathParts, "with_stage_dependencies")
				commonBeforeEach(utils.FixturePath(fixturesPathParts...))
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
								createAndCommitFile(SuiteData.TestDirPath, boundedUserStage, 10)
							},
							checkResultedFilesChecksum: true,
							expectedOutputMatchers: []types.GomegaMatcher{
								ContainSubstring(fmt.Sprintf("Building stage artifact/%s", boundedUserStage)),
							},
						})
						runStagesSpecSteps(specSteps)
					})
				}

				It("nothing should be built", func() {
					specSteps = append(specSteps, toBuildNothingStep)
					runStagesSpecSteps(specSteps)
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

func runStagesSpecSteps(steps []stagesSpecStep) {
	for _, step := range steps {
		By(step.byText)

		if step.beforeBuildHookFunc != nil {
			step.beforeBuildHookFunc()
		}

		out := utils.SucceedCommandOutputString(
			SuiteData.TestDirPath,
			SuiteData.WerfBinPath,
			"build",
		)

		if step.checkResultedFilesChecksum {
			checkResultedFilesChecksum()
		}

		for _, matcher := range step.expectedOutputMatchers {
			Ω(out).Should(matcher)
		}

		out = utils.SucceedCommandOutputString(
			SuiteData.TestDirPath,
			SuiteData.WerfBinPath,
			"build",
		)
		Ω(out).ShouldNot(ContainSubstring("Building stage"))
	}
}

func checkResultedFilesChecksum() {
	containerTestDirPath := "/source"

	expectedFilesChecksum := filesChecksumCommand(containerTestDirPath)
	resultFilesChecksum := filesChecksumCommand("/app")
	diffCommand := fmt.Sprintf("diff <(%s) <(%s)", resultFilesChecksum, expectedFilesChecksum)

	docker.RunSucceedContainerCommandWithStapel(
		SuiteData.WerfBinPath,
		SuiteData.TestDirPath,
		[]string{fmt.Sprintf("-v %s:%s", SuiteData.TestDirPath, containerTestDirPath)},
		[]string{diffCommand},
	)
}

func createAndCommitFile(dirPath, filename string, contentSize int) {
	newFilePath := filepath.Join(dirPath, filename)
	newFileData := []byte(utils.GetRandomString(contentSize))
	utils.WriteFile(newFilePath, newFileData)

	addAndCommitFile(dirPath, filename, "Add file "+filename)
}

func addFile(dirPath, filename string) {
	utils.RunSucceedCommand(
		dirPath,
		"git",
		"add", filename,
	)
}

func addAndCommitFile(dirPath, filename, commitMsg string) {
	addFile(dirPath, filename)

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
