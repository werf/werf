package ci_env_test

import (
	"path/filepath"
	"slices"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/werf"
)

var _ = Describe("Simple ci-env", Label("e2e", "ci-env", "simple"), func() {
	DescribeTable("host cleanup after ci-env",
		func(ctx SpecContext, ciEnvArgs []string, expectDockerDir, expectOutputFile bool) {
			tmpDir := GinkgoT().TempDir()

			werfProject := werf.NewProject(SuiteData.WerfBinPath, tmpDir)

			outputCiEnv := werfProject.CiEnv(ctx, &werf.CiEnvOptions{
				CommonOptions: werf.CommonOptions{
					ExtraArgs: slices.Concat(ciEnvArgs, []string{
						"--tmp-dir", tmpDir,
					}),
				},
			})

			outputHostCleanup := werfProject.HostCleanup(ctx, &werf.HostCleanupOptions{
				CommonOptions: werf.CommonOptions{
					ExtraArgs: []string{
						"--tmp-dir", tmpDir,
					},
				},
			})

			dockerDirPattern := filepath.Join(tmpDir, "werf-docker-config-*")
			dockerDirMatches, err := filepath.Glob(dockerDirPattern)
			Expect(err).To(Succeed())

			if expectDockerDir {
				Expect(dockerDirMatches).To(HaveLen(1))
				Expect(outputHostCleanup).NotTo(ContainSubstring(dockerDirMatches[0]))
			}

			if expectOutputFile {
				Expect(strings.TrimSpace(outputCiEnv)).To(BeARegularFile())
			}
		},
		Entry(
			"should keep copy of ~/.docker dir and outputted env file",
			[]string{
				"gitlab",
				"--as-env-file",
			},
			true,
			true,
		),
		Entry(
			"should keep copy of ~/.docker dir and outputted script file",
			[]string{
				"github",
				"--as-file",
			},
			true,
			true,
		),
	)
})
