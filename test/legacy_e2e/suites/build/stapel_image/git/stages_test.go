package git_test

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/werf/werf/v2/test/pkg/utils"
	"github.com/werf/werf/v2/test/pkg/utils/docker"
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

		BeforeEach(func(ctx SpecContext) {
			fixturesPathParts = append(fixturesPathParts, "image")
			commonBeforeEach(ctx, utils.FixturePath(fixturesPathParts...))
		})

		It("gitArchive stage should be built", func(ctx SpecContext) {
			specSteps = append(specSteps, toBuildGitArchiveStageStep)
			runStagesSpecSteps(ctx, specSteps)
		})

		When("gitArchive stage is built", func() {
			toBuildGitCacheStageStep := stagesSpecStep{
				byText: "Diff between gitArchive commit and current commit >=1MB: gitCache stage should be built",
				beforeBuildHookFunc: func(ctx context.Context) {
					createAndCommitFile(ctx, SuiteData.TestDirPath, "file_1MB", gitCacheSizeStep)
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
				beforeBuildHookFunc: func(ctx context.Context) {
					createAndCommitFile(ctx, SuiteData.TestDirPath, "file_1023KiB", gitCacheSizeStep-1024)
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

			It("gitCache stage should be built (diff between gitArchive commit and current commit >=1MB)", func(ctx SpecContext) {
				specSteps = append(specSteps, toBuildGitCacheStageStep)
				runStagesSpecSteps(ctx, specSteps)
			})

			It("gitLatestPatch stage should be built (diff between gitArchive commit and current commit <1MB)", func(ctx SpecContext) {
				specSteps = append(specSteps, toBuildGitLatestPatchStageStep)
				runStagesSpecSteps(ctx, specSteps)
			})

			When("gitCache stage is built", func() {
				toRepeatedlyBuildGitCacheStageStep := stagesSpecStep{
					byText: "Diff between gitArchive commit and current commit >=1MB: gitCache stage should be built",
					beforeBuildHookFunc: func(ctx context.Context) {
						createAndCommitFile(ctx, SuiteData.TestDirPath, "file2_1MB", gitCacheSizeStep)
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
					beforeBuildHookFunc: func(ctx context.Context) {
						createAndCommitFile(ctx, SuiteData.TestDirPath, "file_1023KiB", gitCacheSizeStep-1024)
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

				It("gitCache stage should be built (diff between gitCache commit and current commit >=1MB)", func(ctx SpecContext) {
					specSteps = append(specSteps, toRepeatedlyBuildGitCacheStageStep)
					runStagesSpecSteps(ctx, specSteps)
				})

				It("gitLatestPatch stage should be built (diff between gitCache commit and current commit <1MB)", func(ctx SpecContext) {
					specSteps = append(specSteps, toBuildGitLatestPatchStageStep)
					runStagesSpecSteps(ctx, specSteps)
				})
			})

			When("gitLatestPatch stage is built", func() {
				BeforeEach(func() {
					specSteps = append(specSteps, toBuildGitLatestPatchStageStep)
				})

				It("gitCache stage should be built (diff between gitArchive commit and current commit >=1MB)", func(ctx SpecContext) {
					specSteps = append(specSteps, toBuildGitCacheStageStep)
					runStagesSpecSteps(ctx, specSteps)
				})

				It("gitLatestPatch stage should be built (diff between gitCache commit and current commit <1MB)", func(ctx SpecContext) {
					specSteps = append(specSteps, toBuildGitLatestPatchStageStep)
					runStagesSpecSteps(ctx, specSteps)
				})
			})
		})
	})

	Context("disableGitAfterPatch", func() {
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
			beforeBuildHookFunc: func(ctx context.Context) {
				createAndCommitFile(ctx, SuiteData.TestDirPath, "file", gitCacheSizeStep)
			},
			checkResultedFilesChecksum: false,
			expectedOutputMatchers: []types.GomegaMatcher{
				Not(ContainSubstring("Building stage")),
				ContainSubstring("Use previously built image for artifact/gitArchive"),
			},
		}

		BeforeEach(func(ctx SpecContext) {
			fixturesPathParts = append(fixturesPathParts, "artifact")
			commonBeforeEach(ctx, utils.FixturePath(fixturesPathParts...))
		})

		It("gitArchive stage should be built", func(ctx SpecContext) {
			specSteps = append(specSteps, toBuildGitArchiveStageStep)
			runStagesSpecSteps(ctx, specSteps)
		})

		When("gitArchive stage is built", func() {
			BeforeEach(func() {
				specSteps = append(specSteps, toBuildGitArchiveStageStep)
			})

			It("nothing should be built", func(ctx SpecContext) {
				specSteps = append(specSteps, toBuildNothingStep)
				runStagesSpecSteps(ctx, specSteps)
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
			beforeBuildHookFunc: func(ctx context.Context) {
				createAndCommitFile(ctx, SuiteData.TestDirPath, "file_1MB", gitCacheSizeStep)
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
			beforeBuildHookFunc: func(ctx context.Context) {
				createAndCommitFile(ctx, SuiteData.TestDirPath, "file_1023KiB", gitCacheSizeStep-1024)
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
			BeforeEach(func(ctx SpecContext) {
				fixturesPathParts = append(fixturesPathParts, "without_stage_dependencies")
				commonBeforeEach(ctx, utils.FixturePath(fixturesPathParts...))
			})

			When("gitArchive stage is built", func() {
				userStagesSpecSetFunc := func() {
					It("gitArchive stage should be built (beforeInstall)", func(ctx SpecContext) {
						specSteps = append(specSteps, stagesSpecStep{
							byText: "beforeInstallCacheVersion changed: beforeInstall stage should be built",
							beforeBuildHookFunc: func(_ context.Context) {
								SuiteData.Stubs.SetEnv("WERF_CONFIG", "werf_beforeInstallCacheVersion.yaml")
							},
							checkResultedFilesChecksum: true,
							expectedOutputMatchers: []types.GomegaMatcher{
								Not(ContainSubstring("stage image/gitCache")),
								Not(ContainSubstring("stage image/gitLatestPatch")),
								ContainSubstring("Building stage image/gitArchive"),
							},
						})
						runStagesSpecSteps(ctx, specSteps)
					})

					for _, userStage := range []string{"install", "beforeSetup", "setup"} {
						boundedUserStage := userStage

						itMsg := fmt.Sprintf("%s stage should be built", boundedUserStage)

						It(itMsg, func(ctx SpecContext) {
							specSteps = append(specSteps, stagesSpecStep{
								byText: fmt.Sprintf("%[1]sCacheVersion changed: %[1]s stage should be built", boundedUserStage),
								beforeBuildHookFunc: func(_ context.Context) {
									SuiteData.Stubs.SetEnv("WERF_CONFIG", fmt.Sprintf("werf_%sCacheVersion.yaml", boundedUserStage))
								},
								checkResultedFilesChecksum: true,
								expectedOutputMatchers: []types.GomegaMatcher{
									Not(ContainSubstring("stage image/gitCache")),
									Not(ContainSubstring("stage image/gitLatestPatch")),
									ContainSubstring(fmt.Sprintf("Building stage image/%s", boundedUserStage)),
								},
							})
							runStagesSpecSteps(ctx, specSteps)
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
			BeforeEach(func(ctx SpecContext) {
				fixturesPathParts = append(fixturesPathParts, "with_stage_dependencies")
				commonBeforeEach(ctx, utils.FixturePath(fixturesPathParts...))
			})

			When("gitArchive stage is built", func() {
				userStagesSpecSetFunc := func() {
					for _, userStage := range []string{"install", "beforeSetup", "setup"} {
						boundedUserStage := userStage

						itMsg := fmt.Sprintf("%s stage should be built", boundedUserStage)

						It(itMsg, func(ctx SpecContext) {
							specSteps = append(specSteps, stagesSpecStep{
								byText: fmt.Sprintf("Dependent file changed: %s stage should be built", boundedUserStage),
								beforeBuildHookFunc: func(ctx context.Context) {
									createAndCommitFile(ctx, SuiteData.TestDirPath, boundedUserStage, 10)
								},
								checkResultedFilesChecksum: true,
								expectedOutputMatchers: []types.GomegaMatcher{
									ContainSubstring(fmt.Sprintf("Building stage image/%s", boundedUserStage)),
								},
							})
							runStagesSpecSteps(ctx, specSteps)
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

	When("disableGitAfterPatch", func() {
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
			beforeBuildHookFunc: func(ctx context.Context) {
				createAndCommitFile(ctx, SuiteData.TestDirPath, "file", gitCacheSizeStep)
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
			BeforeEach(func(ctx SpecContext) {
				fixturesPathParts = append(fixturesPathParts, "without_stage_dependencies")
				commonBeforeEach(ctx, utils.FixturePath(fixturesPathParts...))
			})

			When("gitArchive stage is built", func() {
				toBuildBeforeInstallStageStep := stagesSpecStep{
					byText: fmt.Sprintf("beforeInstallCacheVersion changed: beforeInstall stage should be built"),
					beforeBuildHookFunc: func(_ context.Context) {
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

				It("gitArchive stage should be built (beforeInstall)", func(ctx SpecContext) {
					specSteps = append(specSteps, toBuildBeforeInstallStageStep)
					runStagesSpecSteps(ctx, specSteps)
				})

				for _, userStage := range []string{"install", "beforeSetup", "setup"} {
					boundedUserStage := userStage

					itMsg := fmt.Sprintf("%s stage should be built", boundedUserStage)

					It(itMsg, func(ctx SpecContext) {
						specSteps = append(specSteps, stagesSpecStep{
							byText: fmt.Sprintf("%[1]sCacheVersion changed: %[1]s stage should be built", boundedUserStage),
							beforeBuildHookFunc: func(_ context.Context) {
								SuiteData.Stubs.SetEnv("WERF_CONFIG", fmt.Sprintf("werf_%sCacheVersion.yaml", boundedUserStage))
							},
							checkResultedFilesChecksum: true,
							expectedOutputMatchers: []types.GomegaMatcher{
								ContainSubstring(fmt.Sprintf("Building stage artifact/%s", boundedUserStage)),
							},
						})
						runStagesSpecSteps(ctx, specSteps)
					})
				}

				It("nothing should be built", func(ctx SpecContext) {
					specSteps = append(specSteps, toBuildNothingStep)
					runStagesSpecSteps(ctx, specSteps)
				})
			})
		})

		When("stageDependencies are defined", func() {
			BeforeEach(func(ctx SpecContext) {
				fixturesPathParts = append(fixturesPathParts, "with_stage_dependencies")
				commonBeforeEach(ctx, utils.FixturePath(fixturesPathParts...))
			})

			When("gitArchive stage is built", func() {
				BeforeEach(func() {
					specSteps = append(specSteps, toBuildGitArchiveStageStep)
				})

				for _, userStage := range []string{"install", "beforeSetup", "setup"} {
					boundedUserStage := userStage

					itMsg := fmt.Sprintf("%s stage should be built", boundedUserStage)

					It(itMsg, func(ctx SpecContext) {
						specSteps = append(specSteps, stagesSpecStep{
							byText: fmt.Sprintf("Dependent file changed: %s stage should be built", boundedUserStage),
							beforeBuildHookFunc: func(ctx context.Context) {
								createAndCommitFile(ctx, SuiteData.TestDirPath, boundedUserStage, 10)
							},
							checkResultedFilesChecksum: true,
							expectedOutputMatchers: []types.GomegaMatcher{
								ContainSubstring(fmt.Sprintf("Building stage artifact/%s", boundedUserStage)),
							},
						})
						runStagesSpecSteps(ctx, specSteps)
					})
				}

				It("nothing should be built", func(ctx SpecContext) {
					specSteps = append(specSteps, toBuildNothingStep)
					runStagesSpecSteps(ctx, specSteps)
				})
			})
		})
	})
})

type stagesSpecStep struct {
	byText                     string
	beforeBuildHookFunc        func(ctx context.Context)
	checkResultedFilesChecksum bool
	expectedOutputMatchers     []types.GomegaMatcher
}

func runStagesSpecSteps(ctx context.Context, steps []stagesSpecStep) {
	for _, step := range steps {
		By(step.byText)

		if step.beforeBuildHookFunc != nil {
			step.beforeBuildHookFunc(ctx)
		}

		out := utils.SucceedCommandOutputString(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")

		if step.checkResultedFilesChecksum {
			checkResultedFilesChecksum(ctx)
		}

		for _, matcher := range step.expectedOutputMatchers {
			Expect(out).Should(matcher)
		}

		out = utils.SucceedCommandOutputString(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")
		Expect(out).ShouldNot(ContainSubstring("Building stage"))
	}
}

func checkResultedFilesChecksum(ctx context.Context) {
	containerTestDirPath := "/source"

	expectedFilesChecksum := filesChecksumCommand(containerTestDirPath)
	resultFilesChecksum := filesChecksumCommand("/app")
	diffCommand := fmt.Sprintf("diff <(%s) <(%s)", resultFilesChecksum, expectedFilesChecksum)

	docker.RunSucceedContainerCommandWithStapel(ctx, SuiteData.WerfBinPath, SuiteData.TestDirPath, []string{fmt.Sprintf("-v %s:%s", SuiteData.TestDirPath, containerTestDirPath)}, []string{diffCommand})
}

func createAndCommitFile(ctx context.Context, dirPath, filename string, contentSize int) {
	newFilePath := filepath.Join(dirPath, filename)
	newFileData := []byte(utils.GetRandomString(contentSize))
	utils.WriteFile(newFilePath, newFileData)

	addAndCommitFile(ctx, dirPath, filename, "Add file "+filename)
}

func addFile(ctx context.Context, dirPath, filename string) {
	utils.RunSucceedCommand(
		ctx,
		dirPath,
		"git",
		"add", filename,
	)
}

func addAndCommitFile(ctx context.Context, dirPath, filename, commitMsg string) {
	addFile(ctx, dirPath, filename)

	utils.RunSucceedCommand(ctx, dirPath, "git", "commit", "-m", commitMsg)
}

func filesChecksumCommand(path string) string {
	return fmt.Sprintf(
		"[[ -d %[1]s ]] && find %[1]s -xtype f -not -path '**/.git' -not -path '**/.git/*' -exec bash -c 'printf \"%%s\\n\" \"${@@Q}\"' sh {} + | xargs md5sum | awk '{ print $1 }' | sort | md5sum | awk '{ print $1 }'",
		path,
	)
}
