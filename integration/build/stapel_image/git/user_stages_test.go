// +build integration

package git

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/flant/werf/integration/utils"
)

var _ = Describe("user stages", func() {
	var testDirPath string
	var fixturesPathParts []string
	var specSteps []stagesSpecStep

	BeforeEach(func() {
		testDirPath = tmpPath()
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
				commonBeforeEach(testDirPath, fixturePath(fixturesPathParts...))

				Ω(os.Setenv("BEFORE_INSTALL_CACHE_VERSION", "0")).Should(Succeed())
				Ω(os.Setenv("INSTALL_CACHE_VERSION", "0")).Should(Succeed())
				Ω(os.Setenv("BEFORE_SETUP_CACHE_VERSION", "0")).Should(Succeed())
				Ω(os.Setenv("SETUP_CACHE_VERSION", "0")).Should(Succeed())
			})

			Context("when gitArchive stage is built", func() {
				userStagesSpecSetFunc := func() {
					It("gitArchive stage should be built (beforeInstall)", func() {
						specSteps = append(specSteps, stagesSpecStep{
							byText: "BEFORE_INSTALL_CACHE_VERSION changed: beforeInstall stage should be built",
							beforeBuildHookFunc: func() {
								Ω(os.Setenv("BEFORE_INSTALL_CACHE_VERSION", "1")).Should(Succeed())
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
									Ω(os.Setenv(envName, "2")).Should(Succeed())
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
				commonBeforeEach(testDirPath, fixturePath(fixturesPathParts...))
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
				commonBeforeEach(testDirPath, fixturePath(fixturesPathParts...))

				Ω(os.Setenv("BEFORE_INSTALL_CACHE_VERSION", "0")).Should(Succeed())
				Ω(os.Setenv("INSTALL_CACHE_VERSION", "0")).Should(Succeed())
				Ω(os.Setenv("BEFORE_SETUP_CACHE_VERSION", "0")).Should(Succeed())
				Ω(os.Setenv("SETUP_CACHE_VERSION", "0")).Should(Succeed())
			})

			Context("when gitArchive stage is built", func() {
				toBuildBeforeInstallStageStep := stagesSpecStep{
					byText: "BEFORE_INSTALL_CACHE_VERSION changed: beforeInstall stage should be built",
					beforeBuildHookFunc: func() {
						Ω(os.Setenv("BEFORE_INSTALL_CACHE_VERSION", "1")).Should(Succeed())
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
								Ω(os.Setenv(envName, "2")).Should(Succeed())
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
				commonBeforeEach(testDirPath, fixturePath(fixturesPathParts...))
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
