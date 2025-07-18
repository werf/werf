package e2e_build_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/werf"
)

var _ = Describe("Sbom get", Label("e2e", "sbom", "get", "simple"), func() {
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
						ExtraArgs: []string{"dockerfile"},
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

				Expect(output).To(HavePrefix(`{"$schema":"http://cyclonedx.org/schema/bom-1.6.schema.json"`))
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
		Entry("with local repo using Native Buildah with chroot isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-chroot",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
	)
})
