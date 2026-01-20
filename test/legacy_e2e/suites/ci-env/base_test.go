package ci_env_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/test/pkg/utils"
)

var _ = Describe("base", func() {
	BeforeEach(func(ctx SpecContext) {
		Expect(werf.Init("", "")).Should(Succeed())
		SuiteData.CommitProjectWorktree(ctx, SuiteData.ProjectName, utils.FixturePath("base"), "initial commit")
	})

	ciSystems := []string{
		"gitlab",
		"github",
	}

	for i := range ciSystems {
		ciSystem := ciSystems[i]

		Context(ciSystem, func() {
			It("should print only script path", func(ctx SpecContext) {
				output := utils.SucceedCommandOutputString(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "ci-env", ciSystem, "--as-file")

				expectedPathGlob := filepath.Join(
					werf.GetServiceDir(),
					"tmp",
					"ci_env",
					"source_*_*",
				)

				resultPath := strings.TrimSuffix(output, "\n")
				matched, err := doublestar.PathMatch(expectedPathGlob, resultPath)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(matched).Should(BeTrue(), output)
				Expect(resultPath).Should(BeARegularFile())
			})

			It("should print only shell script", func(ctx SpecContext) {
				output := utils.SucceedCommandOutputString(
					ctx,
					SuiteData.GetProjectWorktree(SuiteData.ProjectName),
					SuiteData.WerfBinPath,
					"ci-env", ciSystem,
				)

				useAsFileOutput := utils.SucceedCommandOutputString(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "ci-env", ciSystem, "--as-file")

				scriptPath := strings.TrimSpace(useAsFileOutput)
				scriptDataByte, err := os.ReadFile(scriptPath)
				Expect(err).ShouldNot(HaveOccurred())

				re := regexp.MustCompile("(.*/tmp/werf-.*?-docker-config-)[0-9]+(.*)")
				scriptData := re.ReplaceAllString(string(scriptDataByte), "${1}${2}")
				output = re.ReplaceAllString(output, "${1}${2}")

				Expect(len(scriptData)).Should(Equal(len(output)))
			})
		})
	}
})
