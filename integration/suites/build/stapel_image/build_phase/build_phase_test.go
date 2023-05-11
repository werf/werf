package build_phase_test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/gookit/color"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/otiai10/copy"

	"github.com/werf/werf/test/pkg/utils"
	"github.com/werf/werf/test/pkg/utils/liveexec"
)

type StageInfo struct {
	ImageID               string
	Repository            string
	Tag                   string
	Digest                string
	UniqueID              string
	CreatedAtUnixMillisec int64
}

func ExtractStageInfoFromOutputLine(stageInfo *StageInfo, line string) *StageInfo {
	if stageInfo == nil {
		stageInfo = &StageInfo{}
	}

	fields := strings.Fields(line)
	if strings.Contains(line, "id: ") {
		stageInfo.ImageID = fields[len(fields)-1]
	}
	if strings.Contains(line, "name: ") {
		nameParts := strings.Split(fields[len(fields)-1], ":")
		stageInfo.Repository = color.ClearCode(nameParts[0])
		stageInfo.Tag = color.ClearCode(nameParts[1])

		sigAndID := strings.SplitN(stageInfo.Tag, "-", 2)
		stageInfo.Digest = sigAndID[0]
		stageInfo.UniqueID = sigAndID[1]

		ts, err := strconv.ParseInt(stageInfo.UniqueID, 10, 64)
		Expect(err).To(BeNil())

		stageInfo.CreatedAtUnixMillisec = ts
	}

	return stageInfo
}

var _ = Describe("Build phase", func() {
	Context("when building the same stage for two commits at the same time", func() {
		BeforeEach(func() {
			SuiteData.Stubs.SetEnv("WERF_VERBOSE", "1")
		})

		AfterEach(func() {
			werfHostPurge("build_phase-001", liveexec.ExecCommandOptions{}, "--force")

			os.RemoveAll("build_phase_repo1")
			os.RemoveAll("build_phase_repo2")
			os.RemoveAll("build_phase-001/.git")
			os.RemoveAll("build_phase-002/.git")
		})

		It("should build install stage twice (because of ancestry check) and use the oldest stage by time of saving into stages storage", func() {
			Expect(utils.SetGitRepoState("build_phase-001", "build_phase_repo1", "one")).To(Succeed())
			Expect(copy.Copy("build_phase_repo1", "build_phase_repo2")).To(Succeed())
			Expect(utils.SetGitRepoState("build_phase-002", "build_phase_repo2", "two")).To(Succeed())

			SuiteData.Stubs.SetEnv("WERF_CONFIG", "werf_1.yaml")
			Expect(werfBuild("build_phase-001", liveexec.ExecCommandOptions{})).To(Succeed())

			var wg sync.WaitGroup
			startFirst := make(chan struct{})
			startSecond := make(chan struct{})
			wg.Add(2)

			var firstCommitInstallStage, secondCommitInstallStage, secondCommitInstallStageOnRetry *StageInfo

			go func() {
				defer wg.Done()
				defer GinkgoRecover()

				<-startFirst
				buildingInstall := false
				stageParserState := ""

				Expect(werfBuild("build_phase-001", liveexec.ExecCommandOptions{
					Env: map[string]string{
						"WERF_TEST_ATOMIC_STAGE_BUILD__SLEEP_SECONDS_BEFORE_STAGE_SAVE": "9",
						"WERF_CONFIG": "werf_2.yaml",
					},
					OutputLineHandler: func(line string) {
						if strings.Contains(line, "Building stage ~/install") {
							buildingInstall = true
							stageParserState = "buildingInstall"
						}

						Expect(strings.Contains(line, "Discarding newly built image for stage")).To(BeFalse(), fmt.Sprintf("should not discard stages, got: %v", line))

						switch stageParserState {
						case "buildingInstall":
							firstCommitInstallStage = ExtractStageInfoFromOutputLine(firstCommitInstallStage, line)
							if firstCommitInstallStage.ImageID != "" {
								stageParserState = ""
							}
						}
					},
				})).To(Succeed())

				Expect(buildingInstall).To(BeTrue(), "should build install stage")
			}()

			go func() {
				defer wg.Done()
				defer GinkgoRecover()

				<-startSecond
				buildingInstall := false
				stageParserState := ""

				Expect(werfBuild("build_phase-002", liveexec.ExecCommandOptions{
					Env: map[string]string{
						"WERF_TEST_ATOMIC_STAGE_BUILD__SLEEP_SECONDS_BEFORE_STAGE_BUILD": "1", // make sure this stage docker-image is created after build_phase-001 install stage docker-image, and despite this fact in the end of the test exactly this stage should be used as a cache
						"WERF_TEST_ATOMIC_STAGE_BUILD__SLEEP_SECONDS_BEFORE_STAGE_SAVE":  "3",
						"WERF_CONFIG": "werf_2.yaml",
					},
					OutputLineHandler: func(line string) {
						if strings.Contains(line, "Building stage ~/install") {
							buildingInstall = true
							stageParserState = "buildingInstall"
						}

						Expect(strings.Contains(line, "Discarding newly built image for stage")).To(BeFalse(), fmt.Sprintf("should not discard stages, got: %v", line))

						switch stageParserState {
						case "buildingInstall":
							secondCommitInstallStage = ExtractStageInfoFromOutputLine(secondCommitInstallStage, line)
							if secondCommitInstallStage.ImageID != "" {
								stageParserState = ""
							}
						}
					},
				})).To(Succeed())

				Expect(buildingInstall).To(BeTrue(), "should build install stage")
			}()

			startFirst <- struct{}{}
			startSecond <- struct{}{}
			wg.Wait()

			Expect(firstCommitInstallStage.ImageID).NotTo(Equal(secondCommitInstallStage.ImageID))
			Expect(firstCommitInstallStage.Repository).To(Equal(secondCommitInstallStage.Repository))
			Expect(firstCommitInstallStage.Digest).To(Equal(secondCommitInstallStage.Digest))
			Expect(firstCommitInstallStage.UniqueID).NotTo(Equal(secondCommitInstallStage.UniqueID))
			Expect(firstCommitInstallStage.CreatedAtUnixMillisec > secondCommitInstallStage.CreatedAtUnixMillisec).To(BeTrue(), "second stage should be saved into stages-storage earlier than first")

			By("first ~/install stage saved into the stages storage should be")

			useCachedInstall := false
			stageParserState := ""
			Expect(werfBuild("build_phase-002", liveexec.ExecCommandOptions{
				Env: map[string]string{
					"WERF_CONFIG": "werf_2.yaml",
				},
				OutputLineHandler: func(line string) {
					if strings.Contains(line, "Use previously built image for ~/install") {
						useCachedInstall = true
						stageParserState = "usingCachedInstall"
					}

					Expect(strings.Contains(line, "Building stage")).To(BeFalse(), fmt.Sprintf("should not build stages, got: %v", line))

					switch stageParserState {
					case "usingCachedInstall":
						secondCommitInstallStageOnRetry = ExtractStageInfoFromOutputLine(secondCommitInstallStageOnRetry, line)
						if secondCommitInstallStageOnRetry.ImageID != "" {
							stageParserState = ""
						}
					}
				},
			})).To(Succeed())

			Expect(useCachedInstall).To(BeTrue(), "should used cached install stage")
			Expect(secondCommitInstallStageOnRetry.ImageID).To(Equal(secondCommitInstallStage.ImageID))
			Expect(secondCommitInstallStageOnRetry.Repository).To(Equal(secondCommitInstallStage.Repository))
			Expect(secondCommitInstallStageOnRetry.Digest).To(Equal(secondCommitInstallStage.Digest))
			Expect(secondCommitInstallStageOnRetry.UniqueID).To(Equal(secondCommitInstallStage.UniqueID))
		})
	})
})
