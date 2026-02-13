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

	Describe("--init-tmp-docker-config flag", Label("init-tmp-docker-config"), func() {
		It("should create empty docker config directory without copying host config", func(ctx SpecContext) {
			tmpDir := GinkgoT().TempDir()

			SuiteData.InitTestRepo(ctx, tmpDir, "some-repo")

			// Create a fake host docker config with some content
			fakeHostDockerConfig := filepath.Join(tmpDir, "fake-docker")
			err := os.MkdirAll(fakeHostDockerConfig, 0o755)
			Expect(err).To(Succeed())

			// Create a config.json file in the fake docker config with credentials
			configContent := `{"auths":{"fake.registry.io":{"auth":"dGVzdDp0ZXN0"}}}`
			err = os.WriteFile(filepath.Join(fakeHostDockerConfig, "config.json"), []byte(configContent), 0o644)
			Expect(err).To(Succeed())

			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(tmpDir))

			outputCiEnv := werfProject.CiEnv(ctx, &werf.CiEnvOptions{
				CommonOptions: werf.CommonOptions{
					ExtraArgs: []string{
						"gitlab",
						"--as-env-file",
						"--tmp-dir", tmpDir,
						"--init-tmp-docker-config",
						"--login-to-registry=false",             // Disable login to ensure config stays empty
						"--docker-config", fakeHostDockerConfig, // Should be ignored when --init-tmp-docker-config is set
					},
				},
			})

			Expect(strings.TrimSpace(outputCiEnv)).To(BeARegularFile())

			// Find the created docker config directory
			dockerDirPattern := filepath.Join(tmpDir, "werf-*-docker-config-*")
			dockerDirMatches, err := filepath.Glob(dockerDirPattern)
			Expect(err).To(Succeed())
			Expect(dockerDirMatches).To(HaveLen(1))

			createdDockerConfig := dockerDirMatches[0]

			// The created docker config should NOT contain the fake host credentials
			configJsonPath := filepath.Join(createdDockerConfig, "config.json")
			configData, err := os.ReadFile(configJsonPath)
			Expect(err).To(Succeed())
			Expect(string(configData)).NotTo(ContainSubstring("fake.registry.io"), "config.json should not contain fake host credentials when --init-tmp-docker-config is used")
		})

		It("should export WERF_DOCKER_CONFIG and DOCKER_CONFIG pointing to tmp dir with --init-tmp-docker-config", func(ctx SpecContext) {
			tmpDir := GinkgoT().TempDir()

			SuiteData.InitTestRepo(ctx, tmpDir, "some-repo")

			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(tmpDir))

			outputCiEnv := werfProject.CiEnv(ctx, &werf.CiEnvOptions{
				CommonOptions: werf.CommonOptions{
					ExtraArgs: []string{
						"github",
						"--as-env-file",
						"--tmp-dir", tmpDir,
						"--init-tmp-docker-config",
					},
				},
			})

			envFilePath := strings.TrimSpace(outputCiEnv)
			Expect(envFilePath).To(BeARegularFile())

			// Read the env file and check that both WERF_DOCKER_CONFIG and DOCKER_CONFIG are set
			envContent, err := os.ReadFile(envFilePath)
			Expect(err).To(Succeed())

			// The env file should contain WERF_DOCKER_CONFIG (priority for werf)
			Expect(string(envContent)).To(ContainSubstring("WERF_DOCKER_CONFIG="))
			// The env file should contain DOCKER_CONFIG (for docker and other tools)
			Expect(string(envContent)).To(ContainSubstring("DOCKER_CONFIG="))
			// Both should point to our tmp dir
			Expect(string(envContent)).To(ContainSubstring(tmpDir))
		})
	})
})
