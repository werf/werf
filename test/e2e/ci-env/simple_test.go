package ci_env_test

import (
	"os"
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

			SuiteData.InitTestRepo(ctx, tmpDir, "some-repo")

			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(tmpDir))

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

			dockerDirPattern := filepath.Join(tmpDir, "werf-*-docker-config-*")
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

	DescribeTable("should export WERF_DOCKER_CONFIG alongside DOCKER_CONFIG",
		func(ctx SpecContext, ciEnvArgs []string, parseEnvLine func(string) (string, string, bool)) {
			tmpDir := GinkgoT().TempDir()

			SuiteData.InitTestRepo(ctx, tmpDir, "some-repo")

			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(tmpDir))

			outputCiEnv := werfProject.CiEnv(ctx, &werf.CiEnvOptions{
				CommonOptions: werf.CommonOptions{
					ExtraArgs: slices.Concat(ciEnvArgs, []string{
						"--tmp-dir", tmpDir,
					}),
				},
			})

			envFilePath := strings.TrimSpace(outputCiEnv)
			Expect(envFilePath).To(BeARegularFile())

			content, err := os.ReadFile(envFilePath)
			Expect(err).To(Succeed())

			envs := map[string]string{}
			for _, line := range strings.Split(string(content), "\n") {
				if key, val, ok := parseEnvLine(line); ok {
					envs[key] = val
				}
			}

			Expect(envs).To(HaveKey("DOCKER_CONFIG"))
			Expect(envs).To(HaveKey("WERF_DOCKER_CONFIG"))
			Expect(envs["WERF_DOCKER_CONFIG"]).To(Equal(envs["DOCKER_CONFIG"]))
		},
		Entry(
			"gitlab --as-env-file",
			[]string{"gitlab", "--as-env-file"},
			func(line string) (string, string, bool) {
				k, v, ok := strings.Cut(line, "=")
				return k, v, ok && !strings.HasPrefix(line, "#")
			},
		),
		Entry(
			"github --as-env-file",
			[]string{"github", "--as-env-file"},
			func(line string) (string, string, bool) {
				k, v, ok := strings.Cut(line, "=")
				return k, v, ok && !strings.HasPrefix(line, "#")
			},
		),
	)
})
