package e2e_build_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/report"
	"github.com/werf/werf/v2/test/pkg/suite_init"
	"github.com/werf/werf/v2/test/pkg/utils"
	"github.com/werf/werf/v2/test/pkg/werf"
)

const sbomProcessingPrefix = "SBOM processing"

var _ = Describe("Simple build", Label("e2e", "build", "sbom", "simple"), func() {
	DescribeTable("should succeed with registry-only SBOM",
		func(ctx SpecContext, testOpts simpleTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)

			By("state0: preparing test repo")
			repoDirname := "repo0"
			fixtureRelPath := "sbom/state0"
			buildReportName := "report0.json"
			SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

			By("state0: building images")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
			reportProject := report.NewProjectWithReport(werfProject)
			buildOut, _ := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath(buildReportName), nil)
			Expect(buildOut).To(ContainSubstring("Building stage"))
			Expect(buildOut).To(ContainSubstring(sbomProcessingPrefix))
		},
		Entry("with local repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("with local repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with rootless isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-rootless",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with chroot isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-chroot",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}, FlakeAttempts(5)),
	)

	DescribeTable("should succeed with registry-only SBOM when base image SBOM is scratch",
		func(ctx SpecContext, testOpts simpleTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)

			By("preparing test repo")
			repoDirname := "repo_base_sbom"
			fixtureRelPath := "sbom/state1"
			SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

			By("building images")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
			reportProject := report.NewProjectWithReport(werfProject)
			buildOut, _ := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath("report_base_sbom.json"), nil)
			Expect(buildOut).To(ContainSubstring(sbomProcessingPrefix))
		},
		Entry("with local repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("with local repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with chroot isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-chroot",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with rootless isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-rootless",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
	)

	DescribeTable("should fail when base image has no SBOM and is not a trusted builder image",
		func(ctx SpecContext, testOpts simpleTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)

			By("preparing test repo")
			repoDirname := "repo_base_sbom"
			fixtureRelPath := "sbom/state2"
			SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

			By("building images expecting failure")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
			buildOut := werfProject.Build(ctx, &werf.BuildOptions{CommonOptions: werf.CommonOptions{ShouldFail: true}})
			Expect(buildOut).To(ContainSubstring("unable to get base image sbom"))
		},
		Entry("with local repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("with local repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with chroot isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-chroot",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with rootless isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-rootless",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
	)

	DescribeTable("should succeed with registry-only SBOM for import stapel",
		func(ctx SpecContext, testOpts simpleTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)

			By("preparing test repo")
			repoDirName := "repo_import_sbom"
			fixtureRelPath := "sbom/import_stapel/state0"
			SuiteData.InitTestRepo(ctx, repoDirName, fixtureRelPath)

			By("building images")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirName))
			reportProject := report.NewProjectWithReport(werfProject)
			buildOut, _ := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath("report_import_sbom.json"), nil)
			Expect(buildOut).To(ContainSubstring(sbomProcessingPrefix))
		},
		Entry("with local repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("with local repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with chroot isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-chroot",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with rootless isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-rootless",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
	)

	DescribeTable("should fail when external import image has no SBOM and is not a trusted builder image",
		func(ctx SpecContext, testOpts simpleTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)

			By("preparing test repo")
			repoDirName := "repo_import_sbom"
			fixtureRelPath := "sbom/import_stapel/state1"
			SuiteData.InitTestRepo(ctx, repoDirName, fixtureRelPath)

			By("building images expecting failure")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirName))
			buildOut := werfProject.Build(ctx, &werf.BuildOptions{CommonOptions: werf.CommonOptions{ShouldFail: true}})
			Expect(buildOut).To(ContainSubstring("unable to get import image sbom"))
		},
		Entry("with local repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("with local repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with chroot isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-chroot",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with rootless isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-rootless",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
	)
})

var _ = Describe("SBOM merge", Label("e2e", "build", "sbom", "merge", "simple"), func() {
	DescribeTable("should succeed with registry-only SBOM for base+fragment merge",
		func(ctx SpecContext, testOpts simpleTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)

			By("preparing test repo")
			repoDirname := "repo_merge_base_fragment"
			fixtureRelPath := "sbom/merge_base_fragment"
			SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

			By("building images")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
			reportProject := report.NewProjectWithReport(werfProject)
			buildOut, _ := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath("report_merge_base_fragment.json"), nil)
			Expect(buildOut).To(ContainSubstring(sbomProcessingPrefix))
		},
		Entry("with local repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("with local repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with chroot isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-chroot",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with rootless isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-rootless",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
	)

	DescribeTable("should succeed with registry-only SBOM for derived+base merge",
		func(ctx SpecContext, testOpts simpleTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)

			By("preparing test repo")
			repoDirname := "repo_merge_derived_with_base"
			fixtureRelPath := "sbom/merge_derived_with_base_fragment"
			SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

			By("building images")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
			reportProject := report.NewProjectWithReport(werfProject)
			buildOut, _ := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath("report_merge_derived_with_base.json"), nil)
			Expect(buildOut).To(ContainSubstring(sbomProcessingPrefix))
		},
		Entry("with local repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("with local repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with chroot isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-chroot",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with rootless isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-rootless",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
	)

	DescribeTable("should succeed with registry-only SBOM for full merge (base+import+fragment)",
		func(ctx SpecContext, testOpts simpleTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)

			By("preparing test repo")
			repoDirname := "repo_merge_full"
			fixtureRelPath := "sbom/merge_full"
			SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

			By("building images")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
			reportProject := report.NewProjectWithReport(werfProject)
			buildOut, _ := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath("report_merge_full.json"), nil)
			Expect(buildOut).To(ContainSubstring(sbomProcessingPrefix))
		},
		Entry("with local repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("with local repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with chroot isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-chroot",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with rootless isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-rootless",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
	)
})

var _ = Describe("SBOM cross-project merge", Label("e2e", "build", "sbom", "merge", "simple"), func() {
	DescribeTable("should succeed with registry-only SBOM for cross-project merge",
		func(ctx SpecContext, testOpts simpleTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)

			By("Step 1: building base project")
			baseRepoDirname := "repo_cross_project_base"
			baseFixtureRelPath := "sbom/cross_project/base"
			SuiteData.InitTestRepo(ctx, baseRepoDirname, baseFixtureRelPath)

			baseWerfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(baseRepoDirname))
			baseReportProject := report.NewProjectWithReport(baseWerfProject)
			baseBuildOut, baseBuildReport := baseReportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath("report_cross_project_base.json"), nil)
			Expect(baseBuildOut).To(ContainSubstring(sbomProcessingPrefix))

			baseReportRecord, ok := baseBuildReport.Images["base-level-0"]
			Expect(ok).To(BeTrue(), "base-level-0 should be in build report")

			By("Step 2: building derived project with BASE_IMAGE_REF env")
			derivedRepoDirname := "repo_cross_project_derived"
			derivedFixtureRelPath := "sbom/cross_project/derived"
			SuiteData.InitTestRepo(ctx, derivedRepoDirname, derivedFixtureRelPath)

			derivedWerfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(derivedRepoDirname))
			derivedReportProject := report.NewProjectWithReport(derivedWerfProject)

			derivedBuildOpts := &werf.WithReportOptions{
				CommonOptions: werf.CommonOptions{
					Envs: []string{
						fmt.Sprintf("BASE_IMAGE_REF=%s", baseReportRecord.DockerImageName),
					},
				},
			}
			derivedBuildOut, _ := derivedReportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath("report_cross_project_derived.json"), derivedBuildOpts)
			Expect(derivedBuildOut).To(ContainSubstring(sbomProcessingPrefix))
		},
		Entry("with local repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("with local repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with chroot isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-chroot",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		XEntry("with local repo using Native Buildah with rootless isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-rootless",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
	)
})

var _ = Describe("GOST SBOM fields", Label("e2e", "build", "sbom", "gost", "simple"), func() {
	DescribeTable("should succeed with registry-only SBOM for GOST fields",
		func(ctx SpecContext, testOpts simpleTestOptions, fixtureRelPath, repoDirname string) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)

			By("preparing test repo")
			SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

			By("building image")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
			reportProject := report.NewProjectWithReport(werfProject)
			buildOut, _ := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath(repoDirname+".json"), nil)
			Expect(buildOut).To(ContainSubstring("Building stage"))
			Expect(buildOut).To(ContainSubstring(sbomProcessingPrefix))
		},
		Entry("default values using Vanilla Docker",
			simpleTestOptions{setupEnvOptions{
				ContainerBackendMode: "vanilla-docker",
				WithLocalRepo:        true,
			}},
			"sbom/gost_defaults",
			"gost-defaults",
		),
		Entry("default values using BuildKit Docker",
			simpleTestOptions{setupEnvOptions{
				ContainerBackendMode: "buildkit-docker",
				WithLocalRepo:        true,
			}},
			"sbom/gost_defaults",
			"gost-defaults",
		),
		Entry("image override meta using Vanilla Docker",
			simpleTestOptions{setupEnvOptions{
				ContainerBackendMode: "vanilla-docker",
				WithLocalRepo:        true,
			}},
			"sbom/gost_meta_image",
			"gost-meta-image",
		),
		Entry("image override meta using BuildKit Docker",
			simpleTestOptions{setupEnvOptions{
				ContainerBackendMode: "buildkit-docker",
				WithLocalRepo:        true,
			}},
			"sbom/gost_meta_image",
			"gost-meta-image",
		),
	)
})

var _ = Describe("SBOM go-replace", Label("e2e", "build", "sbom", "go-replace", "simple"), func() {
	DescribeTable("should succeed with registry-only SBOM for go-replace",
		func(ctx SpecContext, testOpts simpleTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)

			By("preparing test repo")
			repoDirname := "repo_sbom_go_replace"
			fixtureRelPath := "sbom/go_replace"
			SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

			testRepoPath := SuiteData.GetTestRepoPath(repoDirname)
			utils.RunSucceedCommand(ctx, testRepoPath, "git", "tag", "v1.0.0")

			By("building and pushing builder-base image to local registry")
			builderBaseRef := fmt.Sprintf("%s/golang-builder:test", suite_init.TestRegistry())
			utils.RunSucceedCommand(ctx, testRepoPath, "docker", "build", "-t", builderBaseRef, "-f", "Dockerfile.builder-base", ".")
			utils.RunSucceedCommand(ctx, testRepoPath, "docker", "push", builderBaseRef)

			By("building images")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
			reportProject := report.NewProjectWithReport(werfProject)
			buildOpts := &werf.WithReportOptions{
				CommonOptions: werf.CommonOptions{
					Envs: []string{
						fmt.Sprintf("BUILDER_BASE_IMAGE=%s", builderBaseRef),
					},
				},
			}
			buildOut, _ := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath("report_sbom_go_replace.json"), buildOpts)
			Expect(buildOut).To(ContainSubstring("Building stage"))
			Expect(buildOut).To(ContainSubstring(sbomProcessingPrefix))
		},
		Entry("with local repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("with local repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
	)
})
