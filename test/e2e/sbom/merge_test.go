package e2e_build_test

import (
	"encoding/json"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/report"
	"github.com/werf/werf/v2/test/pkg/werf"
)

var _ = Describe("Sbom merge", Label("e2e", "sbom", "merge", "simple"), func() {
	Describe("happy path", Label("simple"), func() {
		DescribeTable("should succeed with registry-only SBOM generation",
			func(ctx SpecContext, testOpts simpleTestOptions) {
				By("initializing")
				setupEnv(testOpts.setupEnvOptions)

				repoDirname := "repo1"
				buildReportPath := filepath.Join(SuiteData.TmpDir, "merge-build-report.json")

				By("preparing test repo")
				SuiteData.InitTestRepo(ctx, repoDirname, "state1")

				By("building images with SBOM")
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
				reportProject := report.NewProjectWithReport(werfProject)
				buildOut, _ := reportProject.BuildWithReport(ctx, buildReportPath, nil)

				Expect(buildOut).To(ContainSubstring(sbomProcessingPrefix))
			},
			Entry("with Vanilla Docker", simpleTestOptions{setupEnvOptions{
				ContainerBackendMode:        "vanilla-docker",
				WithLocalRepo:               true,
				WithStagedDockerfileBuilder: false,
			}}),
		)
	})

	Describe("negative cases", Label("negative"), func() {
		DescribeTable("should fail with invalid input",
			func(ctx SpecContext, extraArgs []string) {
				setupEnv(setupEnvOptions{
					ContainerBackendMode: "vanilla-docker",
					WithLocalRepo:        true,
				})
				repoDirname := "repo1-neg"
				SuiteData.InitTestRepo(ctx, repoDirname, "state1")
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

				werfProject.SbomMerge(ctx, &werf.SbomMergeOptions{
					CommonOptions: werf.CommonOptions{
						ShouldFail: true,
						ExtraArgs:  extraArgs,
					},
				})
			},
			Entry("no flags at all", []string{}),
			Entry("invalid ispras-format",
				[]string{
					"--input", "/dev/null",
					"--repo", "localhost:5000/test",
					"--ispras-format", "invalid-format",
					"--app-name", "app",
					"--app-version", "1.0.0",
					"--manufacturer", "Corp",
				},
			),
			Entry("non-existent mapping file",
				[]string{
					"--input", "/does/not/exist/mapping.json",
					"--repo", "localhost:5000/test",
					"--ispras-format", "oss",
					"--app-name", "app",
					"--app-version", "1.0.0",
					"--manufacturer", "Corp",
				},
			),
		)

		It("fails on empty mapping JSON", func(ctx SpecContext) {
			setupEnv(setupEnvOptions{ContainerBackendMode: "vanilla-docker", WithLocalRepo: true})
			repoDirname := "repo1-neg-empty"
			SuiteData.InitTestRepo(ctx, repoDirname, "state1")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

			mappingPath := filepath.Join(GinkgoT().TempDir(), "empty-mapping.json")
			writeMergeJSON(mappingPath, map[string]string{})

			werfProject.SbomMerge(ctx, &werf.SbomMergeOptions{
				CommonOptions: werf.CommonOptions{
					ShouldFail: true,
					ExtraArgs: []string{
						"--input", mappingPath,
						"--repo", os.Getenv("WERF_REPO"),
						"--ispras-format", "oss",
						"--app-name", "app",
						"--app-version", "1.0.0",
						"--manufacturer", "Corp",
					},
				},
			})
		})

		It("fails on invalid digest in mapping", func(ctx SpecContext) {
			setupEnv(setupEnvOptions{ContainerBackendMode: "vanilla-docker", WithLocalRepo: true})
			repoDirname := "repo1-neg-digest"
			SuiteData.InitTestRepo(ctx, repoDirname, "state1")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

			mappingPath := filepath.Join(GinkgoT().TempDir(), "bad-digest-mapping.json")
			writeMergeJSON(mappingPath, map[string]string{"frontend": "not-a-valid-digest"})

			werfProject.SbomMerge(ctx, &werf.SbomMergeOptions{
				CommonOptions: werf.CommonOptions{
					ShouldFail: true,
					ExtraArgs: []string{
						"--input", mappingPath,
						"--repo", os.Getenv("WERF_REPO"),
						"--ispras-format", "oss",
						"--app-name", "app",
						"--app-version", "1.0.0",
						"--manufacturer", "Corp",
					},
				},
			})
		})
	})
})

func writeMergeJSON(path string, v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	Expect(err).NotTo(HaveOccurred())
	Expect(os.WriteFile(path, data, 0o644)).To(Succeed())
}
