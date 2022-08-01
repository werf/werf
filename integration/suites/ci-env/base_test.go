package ci_env_test

import (
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bmatcuk/doublestar"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/werf"
	"github.com/werf/werf/test/pkg/utils"
)

var _ = Describe("base", func() {
	BeforeEach(func() {
		Ω(werf.Init("", "")).Should(Succeed())
		SuiteData.CommitProjectWorktree(SuiteData.ProjectName, utils.FixturePath("base"), "initial commit")
	})

	ciSystems := []string{
		"gitlab",
		"github",
	}

	for i := range ciSystems {
		ciSystem := ciSystems[i]

		Context(ciSystem, func() {
			It("should print only script path", func() {
				output := utils.SucceedCommandOutputString(
					SuiteData.GetProjectWorktree(SuiteData.ProjectName),
					SuiteData.WerfBinPath,
					utils.WerfBinArgs("ci-env", ciSystem, "--as-file")...,
				)

				expectedPathGlob := filepath.Join(
					werf.GetServiceDir(),
					"tmp",
					"ci_env",
					"source_*_*",
				)

				resultPath := strings.TrimSuffix(output, "\n")
				matched, err := doublestar.PathMatch(expectedPathGlob, resultPath)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(matched).Should(BeTrue(), output)
				Ω(resultPath).Should(BeARegularFile())
			})

			It("should print only shell script", func() {
				output := utils.SucceedCommandOutputString(
					SuiteData.GetProjectWorktree(SuiteData.ProjectName),
					SuiteData.WerfBinPath,
					utils.WerfBinArgs("ci-env", ciSystem)...,
				)

				useAsFileOutput := utils.SucceedCommandOutputString(
					SuiteData.GetProjectWorktree(SuiteData.ProjectName),
					SuiteData.WerfBinPath,
					utils.WerfBinArgs("ci-env", ciSystem, "--as-file")...,
				)

				scriptPath := strings.TrimSpace(useAsFileOutput)
				scriptDataByte, err := ioutil.ReadFile(scriptPath)
				Ω(err).ShouldNot(HaveOccurred())

				re := regexp.MustCompile("(.*/tmp/werf-docker-config-)[0-9]+(.*)")
				scriptData := re.ReplaceAllString(string(scriptDataByte), "${1}${2}")
				output = re.ReplaceAllString(output, "${1}${2}")

				Ω(len(scriptData)).Should(Equal(len(output)))
			})
		})
	}
})
