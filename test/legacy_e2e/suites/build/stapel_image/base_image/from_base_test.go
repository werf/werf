package base_image_test

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/werf/werf/v2/test/pkg/utils"
	utilsDocker "github.com/werf/werf/v2/test/pkg/utils/docker"
)

var _ = XDescribe("from and fromLatest", func() {
	var fromBaseRepoImageState1ID, fromBaseRepoImageState2ID string
	var fromImage string

	fromBaseRepoImageState1IDFunc := func() string { return fromBaseRepoImageState1ID }
	fromBaseRepoImageState2IDFunc := func() string { return fromBaseRepoImageState2ID }

	registryProjectRepositoryLatestAs := func(ctx context.Context, imageName string) {
		if !utilsDocker.IsImageExist(ctx, imageName) {
			Expect(utilsDocker.Pull(ctx, imageName)).Should(Succeed(), "docker pull")
		}
		Expect(utilsDocker.CliTag(ctx, imageName, os.Getenv("WERF_REPO"))).Should(Succeed(), "docker tag")
		Expect(utilsDocker.CliPush(ctx, os.Getenv("WERF_REPO"))).Should(Succeed(), "docker push")
		Expect(utilsDocker.CliRmi(ctx, os.Getenv("WERF_REPO"))).Should(Succeed(), "docker rmi")
	}

	BeforeEach(func(ctx SpecContext) {
		SuiteData.TestDirPath = utils.FixturePath("from_and_from_latest")

		fromBaseRepoImageState1ID = utilsDocker.ImageID(ctx, suiteImage1)
		fromBaseRepoImageState2ID = utilsDocker.ImageID(ctx, suiteImage2)

		fromImage = os.Getenv("WERF_REPO")

		SuiteData.Stubs.SetEnv("FROM_IMAGE", fromImage)
		SuiteData.Stubs.SetEnv("FROM_LATEST", "false")
	})

	type entry struct {
		fromLatest              bool
		expectedOutputMatchers  []types.GomegaMatcher
		expectedFromStageParent func() string
		expectedErr             bool
	}

	entryItBody := func(ctx SpecContext, e entry) {
		SuiteData.Stubs.SetEnv("FROM_LATEST", strconv.FormatBool(e.fromLatest))

		res, err := utils.RunCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")

		if e.expectedErr {
			Expect(err).Should(HaveOccurred())
		} else {
			Expect(err).ShouldNot(HaveOccurred())
		}

		for _, matcher := range e.expectedOutputMatchers {
			Expect(string(res)).Should(matcher)
		}

		if err == nil {
			resultImageName := utils.SucceedCommandOutputString(
				ctx,
				SuiteData.TestDirPath,
				SuiteData.WerfBinPath,
				"stage", "image",
			)

			Expect(utilsDocker.ImageParent(ctx, strings.TrimSpace(resultImageName))).Should(Equal(e.expectedFromStageParent()))
		}
	}

	Context("when from stage is not built", func() {
		Context("when registry from image does not exist", func() {
			Context("when local from image does not exist", func() {
				It("should fail during pulling (fromLatest: false)", func(ctx SpecContext) {
					entryItBody(ctx, entry{
						fromLatest: false,
						expectedOutputMatchers: []types.GomegaMatcher{
							Not(ContainSubstring("Trying to get from base image id from registry")),
							ContainSubstring("Pulling base image"),
							Not(ContainSubstring("Building stage ~/from")),
							ContainSubstring(fmt.Sprintf("Error: phase build on image ~ stage from handler failed: unable to fetch base image %s", fromImage)),
						},
						expectedErr: true,
					})
				})

				It("should fail during getting actual id (fromLatest: true)", func(ctx SpecContext) {
					entryItBody(ctx, entry{
						fromLatest: true,
						expectedOutputMatchers: []types.GomegaMatcher{
							ContainSubstring("Trying to get from base image id from registry"),
							Not(ContainSubstring("Pulling base image")),
							Not(ContainSubstring("Building stage ~/from")),
							ContainSubstring(fmt.Sprintf("Error: can not get base image id from registry (%s)", fromImage)),
						},
						expectedErr: true,
					})
				})
			})

			Context("when local from image exist", func() {
				BeforeEach(func(ctx SpecContext) {
					Expect(utilsDocker.CliTag(ctx, suiteImage1, fromImage)).Should(Succeed(), "docker tag")
				})

				AfterEach(func(ctx SpecContext) {
					utilsDocker.ImageRemoveIfExists(ctx, fromImage)
				})

				It("should be built with local image and warnings (fromLatest: false)", func(ctx SpecContext) {
					entryItBody(ctx, entry{
						fromLatest: false,
						expectedOutputMatchers: []types.GomegaMatcher{
							ContainSubstring("Trying to get from base image id from registry"),
							ContainSubstring("cannot get base image id"),
							ContainSubstring(fmt.Sprintf("using existing image %s without pull", fromImage)),
							ContainSubstring("Trying to get from base image id from registry"),
							Not(ContainSubstring("Pulling base image")),
							ContainSubstring("Building stage ~/from"),
						},
						expectedFromStageParent: fromBaseRepoImageState1IDFunc,
					})
				})

				It("should fail during getting actual id (fromLatest: true)", func(ctx SpecContext) {
					entryItBody(ctx, entry{
						fromLatest: true,
						expectedOutputMatchers: []types.GomegaMatcher{
							ContainSubstring("Trying to get from base image id from registry"),
							Not(ContainSubstring("Pulling base image")),
							Not(ContainSubstring("Building stage ~/from")),
							ContainSubstring(fmt.Sprintf("Error: can not get base image id from registry (%s)", fromImage)),
						},
						expectedErr: true,
					})
				})
			})
		})

		Context("when registry from image exists", func() {
			BeforeEach(func(ctx SpecContext) {
				registryProjectRepositoryLatestAs(ctx, suiteImage2)
			})

			AfterEach(func(ctx SpecContext) {
				utilsDocker.ImageRemoveIfExists(ctx, os.Getenv("WERF_REPO"))
			})

			Context("when local from image is actual", func() {
				BeforeEach(func(ctx SpecContext) {
					Expect(utilsDocker.Pull(ctx, os.Getenv("WERF_REPO"))).Should(Succeed(), "docker pull")
				})

				DescribeTable("checking from stage logic",
					entryItBody,
					Entry("should be built with local image (fromLatest: false)", entry{
						fromLatest: false,
						expectedOutputMatchers: []types.GomegaMatcher{
							ContainSubstring("Trying to get from base image id from registry"),
							Not(ContainSubstring("Pulling base image")),
							ContainSubstring("Building stage ~/from"),
						},
						expectedFromStageParent: fromBaseRepoImageState2IDFunc,
					}),
					Entry("should be built with local image (fromLatest: true)", entry{
						fromLatest: true,
						expectedOutputMatchers: []types.GomegaMatcher{
							ContainSubstring("Trying to get from base image id from registry"),
							Not(ContainSubstring("Pulling base image")),
							ContainSubstring("Building stage ~/from"),
						},
						expectedFromStageParent: fromBaseRepoImageState2IDFunc,
					}),
				)
			})

			Context("when local from image is not actual", func() {
				Context("when from image does not exist locally", func() {
					DescribeTable("checking from stage logic",
						entryItBody,
						Entry("should be built with actual image (fromLatest: false)", entry{
							fromLatest: false,
							expectedOutputMatchers: []types.GomegaMatcher{
								Not(ContainSubstring("Trying to get from base image id from registry")),
								ContainSubstring("Pulling base image"),
								ContainSubstring("Building stage ~/from"),
							},
							expectedFromStageParent: fromBaseRepoImageState2IDFunc,
						}),
						Entry("should be built with actual image (fromLatest: true)", entry{
							fromLatest: true,
							expectedOutputMatchers: []types.GomegaMatcher{
								ContainSubstring("Trying to get from base image id from registry"),
								ContainSubstring("Pulling base image"),
								ContainSubstring("Building stage ~/from"),
							},
							expectedFromStageParent: fromBaseRepoImageState2IDFunc,
						}),
					)
				})

				Context("when from image exists locally", func() {
					BeforeEach(func(ctx SpecContext) {
						Expect(utilsDocker.CliTag(ctx, suiteImage1, os.Getenv("WERF_REPO"))).Should(Succeed(), "docker tag")
					})

					DescribeTable("checking from stage logic",
						entryItBody,
						Entry("should be built with actual image (fromLatest: false)", entry{
							fromLatest: false,
							expectedOutputMatchers: []types.GomegaMatcher{
								ContainSubstring("Trying to get from base image id from registry"),
								ContainSubstring("Pulling base image"),
								ContainSubstring("Building stage ~/from"),
							},
							expectedFromStageParent: fromBaseRepoImageState2IDFunc,
						}),
						Entry("should be built with actual image (fromLatest: true)", entry{
							fromLatest: true,
							expectedOutputMatchers: []types.GomegaMatcher{
								ContainSubstring("Trying to get from base image id from registry"),
								ContainSubstring("Pulling base image"),
								ContainSubstring("Building stage ~/from"),
							},
							expectedFromStageParent: fromBaseRepoImageState2IDFunc,
						}),
					)
				})
			})
		})
	})

	Context("when from stage is built", func() {
		BeforeEach(func(ctx SpecContext) {
			registryProjectRepositoryLatestAs(ctx, suiteImage2)
		})

		AfterEach(func(ctx SpecContext) {
			utilsDocker.ImageRemoveIfExists(ctx, os.Getenv("WERF_REPO"))
		})

		type entryWithPreBuild struct {
			entry
			afterFirstBuildHook func(context.Context)
		}

		entryWithPreBuildItBody := func(ctx SpecContext, e entryWithPreBuild) {
			SuiteData.Stubs.SetEnv("FROM_LATEST", strconv.FormatBool(e.fromLatest))

			utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")

			if e.afterFirstBuildHook != nil {
				e.afterFirstBuildHook(ctx)
			}

			entryItBody(ctx, e.entry)
		}

		Context("when from stage image is actual", func() {
			DescribeTable("checking from stage logic",
				entryWithPreBuildItBody,
				Entry("should not be rebuilt (fromLatest: false)", entryWithPreBuild{
					afterFirstBuildHook: func(ctx context.Context) {
						Expect(utilsDocker.Pull(ctx, os.Getenv("WERF_REPO"))).Should(Succeed(), "docker pull")
					},
					entry: entry{
						fromLatest: false,
						expectedOutputMatchers: []types.GomegaMatcher{
							Not(ContainSubstring("Trying to get from base image id from registry")),
							Not(ContainSubstring("Pulling base image")),
							Not(ContainSubstring("Building stage ~/from")),
						},
						expectedFromStageParent: fromBaseRepoImageState2IDFunc,
					},
				}),
				Entry("should not be rebuilt (fromLatest: true)", entryWithPreBuild{
					afterFirstBuildHook: func(ctx context.Context) {
						Expect(utilsDocker.Pull(ctx, os.Getenv("WERF_REPO"))).Should(Succeed(), "docker pull")
					},
					entry: entry{
						fromLatest: true,
						expectedOutputMatchers: []types.GomegaMatcher{
							ContainSubstring("Trying to get from base image id from registry"),
							Not(ContainSubstring("Pulling base image")),
							Not(ContainSubstring("Building stage ~/from")),
						},
						expectedFromStageParent: fromBaseRepoImageState2IDFunc,
					},
				}),
			)
		})

		Context("when from stage image is not actual", func() {
			DescribeTable("checking from stage logic",
				entryWithPreBuildItBody,
				Entry("should not be rebuilt (fromLatest: false)", entryWithPreBuild{
					afterFirstBuildHook: func(ctx context.Context) {
						Expect(utilsDocker.CliRmi(ctx, os.Getenv("WERF_REPO"))).Should(Succeed(), "docker rmi")
						registryProjectRepositoryLatestAs(ctx, suiteImage1)
					},
					entry: entry{
						fromLatest: false,
						expectedOutputMatchers: []types.GomegaMatcher{
							Not(ContainSubstring("Building stage ~/from")),
							Not(ContainSubstring("Trying to get from base image id from registry")),
							Not(ContainSubstring("Pulling base image")),
						},
						expectedFromStageParent: fromBaseRepoImageState2IDFunc,
					},
				}),
				Entry("should be rebuilt with actual image (fromLatest: true)", entryWithPreBuild{
					afterFirstBuildHook: func(ctx context.Context) {
						Expect(utilsDocker.CliRmi(ctx, os.Getenv("WERF_REPO"))).Should(Succeed(), "docker rmi")
						registryProjectRepositoryLatestAs(ctx, suiteImage1)
					},
					entry: entry{
						fromLatest: true,
						expectedOutputMatchers: []types.GomegaMatcher{
							ContainSubstring("Building stage ~/from"),
							ContainSubstring("Trying to get from base image id from registry"),
							ContainSubstring("Pulling base image"),
						},
						expectedFromStageParent: fromBaseRepoImageState1IDFunc,
					},
				}),
			)
		})
	})
})

var _ = XDescribe("fromCacheVersion", func() {
	BeforeEach(func() {
		SuiteData.TestDirPath = utils.FixturePath("from_cache_version")
	})

	It("should be rebuilt", func(ctx SpecContext) {
		specStep := func(fromCacheVersion string) {
			By(fmt.Sprintf("fromCacheVersion: %s", fromCacheVersion))
			SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", fromCacheVersion)

			output := utils.SucceedCommandOutputString(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")

			Expect(output).Should(ContainSubstring("Building stage ~/from"))
		}

		specStep("0")
		specStep("1")
	})
})
