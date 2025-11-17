package ci_env_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/werf"
)

var _ = Describe("Simple ci-env", Label("e2e", "ci-env", "simple"), func() {
	It("should keep copy of ~/.docker dir after host cleanup", func(ctx SpecContext) {
		tmpDir := GinkgoT().TempDir()

		werfProject := werf.NewProject(SuiteData.WerfBinPath, tmpDir)

		outputCiEnv := werfProject.CiEnv(ctx, &werf.CiEnvOptions{
			CommonOptions: werf.CommonOptions{
				ExtraArgs: []string{
					"gitlab",
					"--tmp-dir", tmpDir,
				},
			},
		})
		Expect(outputCiEnv).To(ContainSubstring("export DOCKER_CONFIG="))

		globPattern := filepath.Join(tmpDir, "werf-docker-config-*")
		Expect(glob(globPattern)).To(HaveLen(1))

		outputHostCleanup := werfProject.HostCleanup(ctx, &werf.HostCleanupOptions{
			CommonOptions: werf.CommonOptions{
				ExtraArgs: []string{
					"--tmp-dir", tmpDir,
				},
			},
		})
		globMatches := glob(globPattern)
		Expect(globMatches).To(HaveLen(1))
		Expect(outputHostCleanup).NotTo(ContainSubstring(globMatches[0]))
	})
})

func glob(pattern string) []string {
	matches, err := filepath.Glob(pattern)
	Expect(err).To(Succeed())
	return matches
}
