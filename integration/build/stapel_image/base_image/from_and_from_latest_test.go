// +build integration

package base_image

import (
	"os"
	"strconv"
	"strings"

	"github.com/onsi/gomega/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/flant/werf/integration/utils"
	utilsDocker "github.com/flant/werf/integration/utils/docker"
)

var _ = Describe("from and from latest", func() {
	var testDirPath string
	var fromBaseRepoImageState1ID, fromBaseRepoImageState2ID string

	fromBaseRepoImageState1IDFunc := func() string { return fromBaseRepoImageState1ID }
	fromBaseRepoImageState2IDFunc := func() string { return fromBaseRepoImageState2ID }

	state1Image := "hello-world"
	state2Image := "alpine"

	registryProjectRepositoryLatestAs := func(imageName string) {
		Ω(utilsDocker.CliPull(imageName)).Should(Succeed(), "docker pull")
		Ω(utilsDocker.CliTag(imageName, registryProjectRepository)).Should(Succeed(), "docker tag")
		Ω(utilsDocker.CliPush(registryProjectRepository)).Should(Succeed(), "docker push")
		Ω(utilsDocker.CliRmi(registryProjectRepository)).Should(Succeed(), "docker rmi")
	}

	BeforeEach(func() {
		testDirPath = tmpPath()

		utils.CopyIn(fixturePath("from_and_from_latest"), testDirPath)

		Ω(utilsDocker.CliPull(state1Image)).Should(Succeed(), "docker pull")
		Ω(utilsDocker.CliPull(state2Image)).Should(Succeed(), "docker pull")

		fromBaseRepoImageState1ID = utilsDocker.ImageID(state1Image)
		fromBaseRepoImageState2ID = utilsDocker.ImageID(state2Image)

		registryProjectRepositoryLatestAs(state2Image)

		Ω(os.Setenv("WERF_STAGES_STORAGE", ":local")).Should(Succeed())
		Ω(os.Setenv("FROM_IMAGE", registryProjectRepository))
		Ω(os.Setenv("FROM_LATEST", "false"))
	})

	AfterEach(func() {
		utilsDocker.ImageRemoveIfExists(registryProjectRepository)

		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"stages", "purge", "-s", ":local", "--force",
		)
	})

	type entry struct {
		fromLatest              bool
		expectedOutputMatchers  []types.GomegaMatcher
		expectedFromStageParent func() string
	}

	entryItBody := func(e entry) {
		Ω(os.Setenv("FROM_LATEST", strconv.FormatBool(e.fromLatest)))

		out := utils.SucceedCommandOutput(
			testDirPath,
			werfBinPath,
			"build",
		)

		for _, matcher := range e.expectedOutputMatchers {
			Ω(out).Should(matcher)
		}

		resultImageName := utils.SucceedCommandOutput(
			testDirPath,
			werfBinPath,
			"stage", "image",
		)

		Ω(utilsDocker.ImageParent(strings.TrimSpace(resultImageName))).Should(Equal(e.expectedFromStageParent()))
	}

	Context("when from stage is not built", func() {
		Context("when local from image is actual", func() {
			BeforeEach(func() {
				Ω(utilsDocker.CliPull(registryProjectRepository)).Should(Succeed(), "docker pull")
			})

			DescribeTable("checking from stage logic",
				entryItBody,
				Entry("should be built without pulling (fromLatest: false)", entry{
					fromLatest: false,
					expectedOutputMatchers: []types.GomegaMatcher{
						ContainSubstring("Trying to get from base image id from registry"),
						Not(ContainSubstring("Pulling base image")),
						ContainSubstring("Building stage from"),
					},
					expectedFromStageParent: fromBaseRepoImageState2IDFunc,
				}),
				Entry("should be built without pulling (fromLatest: true)", entry{
					fromLatest: true,
					expectedOutputMatchers: []types.GomegaMatcher{
						ContainSubstring("Trying to get from base image id from registry"),
						Not(ContainSubstring("Pulling base image")),
						ContainSubstring("Building stage from"),
					},
					expectedFromStageParent: fromBaseRepoImageState2IDFunc,
				}),
			)
		})

		Context("when local from image is not actual", func() {
			Context("when from image does not exist locally", func() {
				DescribeTable("checking from stage logic",
					entryItBody,
					Entry("should be built with pulling, without checking registry (fromLatest: false)", entry{
						fromLatest: false,
						expectedOutputMatchers: []types.GomegaMatcher{
							Not(ContainSubstring("Trying to get from base image id from registry")),
							ContainSubstring("Pulling base image"),
							ContainSubstring("Building stage from"),
						},
						expectedFromStageParent: fromBaseRepoImageState2IDFunc,
					}),
					Entry("should be built with getting id and pulling (fromLatest: true)", entry{
						fromLatest: true,
						expectedOutputMatchers: []types.GomegaMatcher{
							ContainSubstring("Trying to get from base image id from registry"),
							ContainSubstring("Pulling base image"),
							ContainSubstring("Building stage from"),
						},
						expectedFromStageParent: fromBaseRepoImageState2IDFunc,
					}),
				)
			})

			Context("when from image exists locally", func() {
				BeforeEach(func() {
					Ω(utilsDocker.CliPull(state1Image)).Should(Succeed(), "docker pull")
					Ω(utilsDocker.CliTag(state1Image, registryProjectRepository)).Should(Succeed(), "docker tag")
				})

				DescribeTable("checking from stage logic",
					entryItBody,
					Entry("should be built with pulling and checking registry (fromLatest: false)", entry{
						fromLatest: false,
						expectedOutputMatchers: []types.GomegaMatcher{
							ContainSubstring("Trying to get from base image id from registry"),
							ContainSubstring("Pulling base image"),
							ContainSubstring("Building stage from"),
						},
						expectedFromStageParent: fromBaseRepoImageState2IDFunc,
					}),
					Entry("should be built with getting id and pulling (fromLatest: true)", entry{
						fromLatest: true,
						expectedOutputMatchers: []types.GomegaMatcher{
							ContainSubstring("Trying to get from base image id from registry"),
							ContainSubstring("Pulling base image"),
							ContainSubstring("Building stage from"),
						},
						expectedFromStageParent: fromBaseRepoImageState2IDFunc,
					}),
				)
			})
		})
	})

	Context("when from stage is built", func() {
		type entryWithPreBuild struct {
			entry
			afterFirstBuildHook func()
		}

		entryWithPreBuildItBody := func(e entryWithPreBuild) {
			Ω(os.Setenv("FROM_LATEST", strconv.FormatBool(e.fromLatest)))

			utils.RunSucceedCommand(
				testDirPath,
				werfBinPath,
				"build",
			)

			if e.afterFirstBuildHook != nil {
				e.afterFirstBuildHook()
			}

			entryItBody(e.entry)
		}

		Context("when from stage image is actual", func() {
			DescribeTable("checking from stage logic",
				entryWithPreBuildItBody,
				Entry("should not be rebuilt, without checking registry (fromLatest: false)", entryWithPreBuild{
					afterFirstBuildHook: func() {
						Ω(utilsDocker.CliPull(registryProjectRepository)).Should(Succeed(), "docker pull")
					},
					entry: entry{
						fromLatest: false,
						expectedOutputMatchers: []types.GomegaMatcher{
							Not(ContainSubstring("Trying to get from base image id from registry")),
							Not(ContainSubstring("Pulling base image")),
							Not(ContainSubstring("Building stage from")),
						},
						expectedFromStageParent: fromBaseRepoImageState2IDFunc,
					},
				}),
				Entry("should not be rebuilt, with checking registry (fromLatest: true)", entryWithPreBuild{
					afterFirstBuildHook: func() {
						Ω(utilsDocker.CliPull(registryProjectRepository)).Should(Succeed(), "docker pull")
					},
					entry: entry{
						fromLatest: true,
						expectedOutputMatchers: []types.GomegaMatcher{
							ContainSubstring("Trying to get from base image id from registry"),
							Not(ContainSubstring("Pulling base image")),
							Not(ContainSubstring("Building stage from")),
						},
						expectedFromStageParent: fromBaseRepoImageState2IDFunc,
					},
				}),
			)
		})

		Context("when from stage image is not actual", func() {
			DescribeTable("checking from stage logic",
				entryWithPreBuildItBody,
				Entry("should not be rebuilt, without checking registry (fromLatest: false)", entryWithPreBuild{
					afterFirstBuildHook: func() {
						Ω(utilsDocker.CliRmi(registryProjectRepository)).Should(Succeed(), "docker rmi")
						registryProjectRepositoryLatestAs(state1Image)
					},
					entry: entry{
						fromLatest: false,
						expectedOutputMatchers: []types.GomegaMatcher{
							Not(ContainSubstring("Building stage from")),
							Not(ContainSubstring("Trying to get from base image id from registry")),
							Not(ContainSubstring("Pulling base image")),
						},
						expectedFromStageParent: fromBaseRepoImageState2IDFunc,
					},
				}),
				Entry("should be rebuilt with getting id and pulling (fromLatest: true)", entryWithPreBuild{
					afterFirstBuildHook: func() {
						Ω(utilsDocker.CliRmi(registryProjectRepository)).Should(Succeed(), "docker rmi")
						registryProjectRepositoryLatestAs(state1Image)
					},
					entry: entry{
						fromLatest: true,
						expectedOutputMatchers: []types.GomegaMatcher{
							ContainSubstring("Building stage from"),
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
