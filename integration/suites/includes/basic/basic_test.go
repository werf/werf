package basic_test

import (
	"fmt"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"

	"github.com/werf/werf/v2/pkg/includes"
	"github.com/werf/werf/v2/test/pkg/contback"
	"github.com/werf/werf/v2/test/pkg/utils"
	"github.com/werf/werf/v2/test/pkg/werf"
)

type simpleTestOptions struct {
	setupEnvOptions
}

type Config struct {
	Includes []includeConf `yaml:"includes"`
}

type includeConf struct {
	Name         string   `yaml:"name"`
	Git          string   `yaml:"git"`
	Branch       string   `yaml:"branch"`
	Tag          string   `yaml:"tag"`
	Commit       string   `yaml:"commit"`
	Add          string   `yaml:"add,omitempty"`
	To           string   `yaml:"to,omitempty"`
	IncludePaths []string `yaml:"includePaths"`
	ExcludePaths []string `yaml:"excludePaths"`
}

var _ = Describe("build and mutate image spec", Label("integration", "build", "mutate spec config"), func() {
	DescribeTable("should succeed and produce expected image",
		func(ctx SpecContext, testOpts simpleTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)
			_, err := contback.NewContainerBackend(testOpts.ContainerBackendMode)
			if err == contback.ErrRuntimeUnavailable {
				Skip(err.Error())
			} else if err != nil {
				Fail(err.Error())
			}

			By("prepearing repos")
			{
				mainRepoDirName := "main_repo"
				fixtureRelPath := "basic"
				SuiteData.InitTestRepo(ctx, mainRepoDirName, filepath.Join(fixtureRelPath, mainRepoDirName))

				remoteRepoDirName1 := "remote_repo1"
				SuiteData.InitTestRepo(ctx, remoteRepoDirName1, filepath.Join(fixtureRelPath, remoteRepoDirName1))
				branch1 := getBranchName(ctx, SuiteData.GetTestRepoPath(remoteRepoDirName1))

				remoteRepoDirName2 := "remote_repo2"
				SuiteData.InitTestRepo(ctx, remoteRepoDirName2, filepath.Join(fixtureRelPath, remoteRepoDirName2))
				branch2 := getBranchName(ctx, SuiteData.GetTestRepoPath(remoteRepoDirName2))

				SuiteData.WerfRepo = SuiteData.GetTestRepoPath(mainRepoDirName)
				includesConfig := &Config{
					Includes: []includeConf{
						{
							Name:   "remote_repo1",
							Git:    SuiteData.GetTestRepoPath(remoteRepoDirName1),
							Branch: branch1,
							Add:    "/",
							To:     "/",
						},
						{
							Name:   "remote_repo2",
							Git:    SuiteData.GetTestRepoPath(remoteRepoDirName2),
							Branch: branch2,
							Add:    "/",
							To:     "/",
						},
					},
				}

				out, err := yaml.Marshal(includesConfig)
				Expect(err).ToNot(HaveOccurred())
				Expect(out).ToNot(BeEmpty())

				werfIncludesPath := includes.GetWerfIncludesConfigRelPath("", SuiteData.WerfRepo)
				By(fmt.Sprintf("writing includes config to %s", werfIncludesPath))
				{
					utils.WriteFile(werfIncludesPath, out)
				}
				By("generate includes lock file")
				{
					utils.RunSucceedCommand(ctx,
						SuiteData.GetTestRepoPath(mainRepoDirName),
						SuiteData.WerfBinPath,
						"includes", "update", "--dev",
					)
				}
				By("committing includes config")
				{
					utils.RunSucceedCommand(ctx, SuiteData.GetTestRepoPath(mainRepoDirName), "git", "add", "-A")
					utils.RunSucceedCommand(ctx, SuiteData.GetTestRepoPath(mainRepoDirName), "git", "commit", "-m", "add includes config")
				}
			}
			By(fmt.Sprintf("starting"))
			{
				repoDirname := "main_repo"

				buildReportName := "build_report.json"
				By(fmt.Sprintf("building images"))
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
				buildOut, _ := werfProject.BuildWithReport(ctx,
					SuiteData.GetBuildReportPath(buildReportName),
					nil,
				)
				Expect(buildOut).To(ContainSubstring("Building stage"))

				By(fmt.Sprintf("render chart"))
				output := utils.SucceedCommandOutputString(
					ctx,
					SuiteData.GetTestRepoPath(repoDirname),
					SuiteData.WerfBinPath,
					"render",
				)
				for _, substrFormat := range []string{
					"# Source: %s/templates/backend.yaml",
				} {
					Expect(output).Should(ContainSubstring(fmt.Sprintf(substrFormat, utils.ProjectName())))
				}
			}
		},
		Entry("without local repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode: "buildkit-docker",
		}}),
	)
})

func getBranchName(ctx SpecContext, repoDir string) string {
	out, err := utils.RunCommand(ctx, repoDir, "git", "rev-parse", "--abbrev-ref", "HEAD")
	Expect(err).ToNot(HaveOccurred())
	return strings.TrimSpace(string(out))
}
