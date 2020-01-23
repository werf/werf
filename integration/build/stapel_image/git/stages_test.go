package git_test

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/flant/werf/pkg/testing/utils"
	"github.com/flant/werf/pkg/testing/utils/docker"
)

var _ = Describe("git stages", func() {
	var fixturesPathParts []string
	var specSteps []stagesSpecStep

	BeforeEach(func() {
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
			commonBeforeEach(testDirPath, utils.FixturePath(fixturesPathParts...))
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
			commonBeforeEach(testDirPath, utils.FixturePath(fixturesPathParts...))
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

var _ = Describe("user stages", func() {
	var fixturesPathParts []string
	var specSteps []stagesSpecStep

	BeforeEach(func() {
		fixturesPathParts = []string{"user_stages"}
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
			fixturesPathParts = append(fixturesPathParts, "image")
		})

		Context("when stageDependencies are not defined", func() {
			BeforeEach(func() {
				fixturesPathParts = append(fixturesPathParts, "without_stage_dependencies")
				commonBeforeEach(testDirPath, utils.FixturePath(fixturesPathParts...))
			})

			Context("when gitArchive stage is built", func() {
				userStagesSpecSetFunc := func() {
					It("gitArchive stage should be built (beforeInstall)", func() {
						specSteps = append(specSteps, stagesSpecStep{
							byText: "BEFORE_INSTALL_CACHE_VERSION changed: beforeInstall stage should be built",
							beforeBuildHookFunc: func() {
								stubs.SetEnv("BEFORE_INSTALL_CACHE_VERSION", "1")
							},
							checkResultedFilesChecksum: true,
							expectedOutputMatchers: []types.GomegaMatcher{
								ContainSubstring("gitCache:               <empty>"),
								ContainSubstring("gitLatestPatch:         <empty>"),
								ContainSubstring("Git files will be actualized on stage gitArchive"),
								ContainSubstring("Building stage gitArchive"),
							},
						})
						runStagesSpecSteps(testDirPath, specSteps)
					})

					for _, userStage := range []string{"install", "beforeSetup", "setup"} {
						boundedUserStage := userStage

						itMsg := fmt.Sprintf("%s stage should be built", boundedUserStage)

						It(itMsg, func() {
							var envPrefixName string
							switch boundedUserStage {
							case "install":
								envPrefixName = "INSTALL"
							case "beforeSetup":
								envPrefixName = "BEFORE_SETUP"
							case "setup":
								envPrefixName = "SETUP"
							}

							envName := envPrefixName + "_CACHE_VERSION"

							specSteps = append(specSteps, stagesSpecStep{
								byText: fmt.Sprintf("%s changed: %s stage should be built", envName, boundedUserStage),
								beforeBuildHookFunc: func() {
									stubs.SetEnv(envName, "2")
								},
								checkResultedFilesChecksum: true,
								expectedOutputMatchers: []types.GomegaMatcher{
									ContainSubstring("gitCache:               <empty>"),
									ContainSubstring("gitLatestPatch:         <empty>"),
									ContainSubstring(fmt.Sprintf("Git files will be actualized on stage %s", boundedUserStage)),
									ContainSubstring(fmt.Sprintf("Building stage %s", boundedUserStage)),
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

				Context("when gitCache stage is built", func() {
					BeforeEach(func() {
						specSteps = append(specSteps, toBuildGitCacheStageStep)
					})

					userStagesSpecSetFunc()
				})

				Context("when gitLatestPatch stage is built", func() {
					BeforeEach(func() {
						specSteps = append(specSteps, toBuildGitLatestPatchStageStep)
					})

					userStagesSpecSetFunc()
				})
			})
		})

		Context("when stageDependencies are defined", func() {
			BeforeEach(func() {
				fixturesPathParts = append(fixturesPathParts, "with_stage_dependencies")
				commonBeforeEach(testDirPath, utils.FixturePath(fixturesPathParts...))
			})

			Context("when gitArchive stage is built", func() {
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
									ContainSubstring(fmt.Sprintf("Git files will be actualized on stage %s", boundedUserStage)),
									ContainSubstring(fmt.Sprintf("Building stage %s", boundedUserStage)),
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

				Context("when gitCache stage is built", func() {
					BeforeEach(func() {
						specSteps = append(specSteps, toBuildGitCacheStageStep)
					})

					userStagesSpecSetFunc()
				})

				Context("when gitLatestPatch stage is built", func() {
					BeforeEach(func() {
						specSteps = append(specSteps, toBuildGitLatestPatchStageStep)
					})

					userStagesSpecSetFunc()
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
		})

		Context("when stageDependencies are not defined", func() {
			BeforeEach(func() {
				fixturesPathParts = append(fixturesPathParts, "without_stage_dependencies")
				commonBeforeEach(testDirPath, utils.FixturePath(fixturesPathParts...))
			})

			Context("when gitArchive stage is built", func() {
				toBuildBeforeInstallStageStep := stagesSpecStep{
					byText: "BEFORE_INSTALL_CACHE_VERSION changed: beforeInstall stage should be built",
					beforeBuildHookFunc: func() {
						stubs.SetEnv("BEFORE_INSTALL_CACHE_VERSION", "1")
					},
					checkResultedFilesChecksum: true,
					expectedOutputMatchers: []types.GomegaMatcher{
						ContainSubstring("Git files will be actualized on stage gitArchive"),
						ContainSubstring("Building stage gitArchive"),
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
						var envPrefixName string
						switch boundedUserStage {
						case "install":
							envPrefixName = "INSTALL"
						case "beforeSetup":
							envPrefixName = "BEFORE_SETUP"
						case "setup":
							envPrefixName = "SETUP"
						}

						envName := envPrefixName + "_CACHE_VERSION"

						specSteps = append(specSteps, stagesSpecStep{
							byText: fmt.Sprintf("%s changed: %s stage should be built", envName, boundedUserStage),
							beforeBuildHookFunc: func() {
								stubs.SetEnv(envName, "2")
							},
							checkResultedFilesChecksum: true,
							expectedOutputMatchers: []types.GomegaMatcher{
								ContainSubstring(fmt.Sprintf("Git files will be actualized on stage %s", boundedUserStage)),
								ContainSubstring(fmt.Sprintf("Building stage %s", boundedUserStage)),
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

		Context("when stageDependencies are defined", func() {
			BeforeEach(func() {
				fixturesPathParts = append(fixturesPathParts, "with_stage_dependencies")
				commonBeforeEach(testDirPath, utils.FixturePath(fixturesPathParts...))
			})

			Context("when gitArchive stage is built", func() {
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
								ContainSubstring(fmt.Sprintf("Git files will be actualized on stage %s", boundedUserStage)),
								ContainSubstring(fmt.Sprintf("Building stage %s", boundedUserStage)),
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
		Ω(out).ShouldNot(ContainSubstring("Building stage "))
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
