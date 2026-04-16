package e2e_build_test

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/report"
	"github.com/werf/werf/v2/test/pkg/werf"
)

var _ = Describe("Sbom get", Label("e2e", "sbom", "get", "simple"), func() {
	Describe("default", func() {
		DescribeTable("should generate and store SBOM as an image",
			func(ctx SpecContext, testOpts simpleTestOptions) {
				By("initializing")
				setupEnv(testOpts.setupEnvOptions)

				By("state0: case", func() {
					repoDirname := "repo0"
					fixtureRelPath := "state0"

					By("state0: preparing test repo")
					SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

					By("state0: building images")
					werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

					output := werfProject.SbomGet(ctx, &werf.SbomGetOptions{
						CommonOptions: werf.CommonOptions{
							ExtraArgs: []string{"stapel"},
						},
					})

					switch testOpts.ContainerBackendMode {
					case "vanilla-docker", "buildkit-docker":
						// TODO: remove workaround for Docker backend after fixing
						// Note: Generation of SBOM returns something like
						// `sha256:bee01feb22b978b11472e8bc86065fd88ee370c9782288536ddb58e9904aa584`
						// in the first line of output. So, we need to omit this noize.
						output = output[(71 + 1):]
					}

					Expect(output).To(ContainSubstring(`{"$schema":"http://cyclonedx.org/schema/bom-1.6.schema.json"`))
				})
			},
			Entry("without repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
				ContainerBackendMode:        "vanilla-docker",
				WithLocalRepo:               false,
				WithStagedDockerfileBuilder: false,
			}}),
			Entry("with local repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
				ContainerBackendMode:        "vanilla-docker",
				WithLocalRepo:               true,
				WithStagedDockerfileBuilder: false,
			}}),
			Entry("without repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
				ContainerBackendMode:        "buildkit-docker",
				WithLocalRepo:               false,
				WithStagedDockerfileBuilder: false,
			}}),
			Entry("with local repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
				ContainerBackendMode:        "buildkit-docker",
				WithLocalRepo:               true,
				WithStagedDockerfileBuilder: false,
			}}),
			// TODO (zaytsev): it does not work currently
			// https://github.com/werf/werf/actions/runs/15076648086/job/42385521980?pr=6860#step:11:150
			XEntry("with local repo using Native Buildah with rootless isolation", simpleTestOptions{setupEnvOptions{
				ContainerBackendMode:        "native-rootless",
				WithLocalRepo:               true,
				WithStagedDockerfileBuilder: false,
			}}),
			XEntry("with local repo using Native Buildah with chroot isolation", simpleTestOptions{setupEnvOptions{
				ContainerBackendMode:        "native-chroot",
				WithLocalRepo:               true,
				WithStagedDockerfileBuilder: false,
			}}),
		)
	})

	Describe("lightweight", Label("tag"), func() {
		DescribeTable("should get SBOM by content-based tag",
			func(ctx SpecContext, testOpts simpleTestOptions) {
				By("initializing")
				setupEnv(testOpts.setupEnvOptions)

				repoDirname := "repo0-tag"
				buildReportPath := filepath.Join(SuiteData.TmpDir, "get-tag-build-report.json")

				By("preparing test repo")
				SuiteData.InitTestRepo(ctx, repoDirname, "state0")

				By("building images with report")
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
				reportProject := report.NewProjectWithReport(werfProject)
				_, buildReport := reportProject.BuildWithReport(ctx, buildReportPath, nil)

				record, ok := buildReport.Images["stapel"]
				Expect(ok).To(BeTrue(), "build report must contain 'stapel' image")
				Expect(record.DockerTag).NotTo(BeEmpty(), "DockerTag must not be empty")

				By("getting SBOM by --tag")
				output := werfProject.SbomGet(ctx, &werf.SbomGetOptions{
					CommonOptions: werf.CommonOptions{
						ExtraArgs: []string{
							"--tag", record.DockerTag,
							"--repo", os.Getenv("WERF_REPO"),
						},
					},
				})

				Expect(output).To(ContainSubstring(`"$schema":"http://cyclonedx.org/schema/bom-1.6.schema.json"`))
			},
			Entry("without repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
				ContainerBackendMode:        "vanilla-docker",
				WithLocalRepo:               false,
				WithStagedDockerfileBuilder: false,
			}}),
			Entry("with local repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
				ContainerBackendMode:        "vanilla-docker",
				WithLocalRepo:               true,
				WithStagedDockerfileBuilder: false,
			}}),
			Entry("without repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
				ContainerBackendMode:        "buildkit-docker",
				WithLocalRepo:               false,
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
			}}),
		)
	})

	Describe("lightweight", Label("digest"), func() {
		DescribeTable("should get SBOM by image digest",
			func(ctx SpecContext, testOpts simpleTestOptions) {
				By("initializing")
				setupEnv(testOpts.setupEnvOptions)

				repoDirname := "repo0-digest"
				buildReportPath := filepath.Join(SuiteData.TmpDir, "get-digest-build-report.json")

				By("preparing test repo")
				SuiteData.InitTestRepo(ctx, repoDirname, "state0")

				By("building images with report")
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
				reportProject := report.NewProjectWithReport(werfProject)
				_, buildReport := reportProject.BuildWithReport(ctx, buildReportPath, nil)

				record, ok := buildReport.Images["stapel"]
				Expect(ok).To(BeTrue(), "build report must contain 'stapel' image")

				imageDigest := record.DockerImageDigest
				if !strings.HasPrefix(imageDigest, "sha256:") {
					imageDigest = "sha256:" + imageDigest
				}
				Expect(imageDigest).NotTo(BeEmpty(), "DockerImageDigest must not be empty")

				By("getting SBOM by --digest")
				output := werfProject.SbomGet(ctx, &werf.SbomGetOptions{
					CommonOptions: werf.CommonOptions{
						ExtraArgs: []string{
							"--digest", imageDigest,
							"--repo", os.Getenv("WERF_REPO"),
						},
					},
				})

				Expect(output).To(ContainSubstring(`"$schema":"http://cyclonedx.org/schema/bom-1.6.schema.json"`))
			},
			Entry("without repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
				ContainerBackendMode:        "vanilla-docker",
				WithLocalRepo:               false,
				WithStagedDockerfileBuilder: false,
			}}),
			Entry("with local repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
				ContainerBackendMode:        "vanilla-docker",
				WithLocalRepo:               true,
				WithStagedDockerfileBuilder: false,
			}}),
			Entry("without repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
				ContainerBackendMode:        "buildkit-docker",
				WithLocalRepo:               false,
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
			}}),
		)
	})

	Describe("negative cases", Label("negative"), func() {
		DescribeTable("should fail with mutually exclusive flags",
			func(ctx SpecContext, extraArgs []string) {
				setupEnv(setupEnvOptions{
					ContainerBackendMode: "vanilla-docker",
					WithLocalRepo:        true,
				})

				repoDirname := "repo0-neg"
				SuiteData.InitTestRepo(ctx, repoDirname, "state0")
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

				werfProject.SbomGet(ctx, &werf.SbomGetOptions{
					CommonOptions: werf.CommonOptions{
						ShouldFail: true,
						ExtraArgs:  extraArgs,
					},
				})
			},
			Entry("--tag and --digest together",
				[]string{
					"--tag", "some-tag",
					"--digest", "sha256:abc123",
					"--repo", "localhost:5000/test",
				},
			),
		)
	})
})
