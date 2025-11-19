package e2e_verify_test

import (
	_ "embed"
	"encoding/base64"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/report"
	"github.com/werf/werf/v2/test/pkg/utils/gpg"
	"github.com/werf/werf/v2/test/pkg/werf"
)

var (
	// ----- Inhouse -----
	//go:embed _fixtures/signature/inhouse/keys/delivery-kit_959497322.pem.key
	testKeyData []byte
	//go:embed _fixtures/signature/inhouse/keys/delivery-kit_1666162742.pem.crt
	testCertData []byte
	//go:embed _fixtures/signature/inhouse/keys/delivery-kit_chain_3247019714.pem.crt
	testChainData []byte
	//go:embed _fixtures/signature/inhouse/keys/delivery-kit_intermediates_1698313569.pem.crt
	testIntermediatesData []byte
	//go:embed _fixtures/signature/inhouse/keys/delivery-kit_root_1959509881.pem.crt
	testRootCertData []byte

	// ----- BSign -----
	//go:embed _fixtures/signature/bsign/keys/rsa_4096_private.pgp
	testBSignPrivateKeyData []byte
	//go:embed _fixtures/signature/bsign/keys/rsa_4096_public.pgp
	testBSignPublicKeyData []byte
)

var (
	// ----- Inhouse -----
	testKeyBase64           = base64.StdEncoding.EncodeToString(testKeyData)
	testCertBase64          = base64.StdEncoding.EncodeToString(testCertData)
	_                       = base64.StdEncoding.EncodeToString(testChainData)
	testIntermediatesBase64 = base64.StdEncoding.EncodeToString(testIntermediatesData)
	testRootCertBase64      = base64.StdEncoding.EncodeToString(testRootCertData)

	// ----- BSign -----
	testBSignPrivateKeyBase64 = base64.StdEncoding.EncodeToString(testBSignPrivateKeyData)
	_                         = base64.StdEncoding.EncodeToString(testBSignPublicKeyData)
)

const (
// ----- BSign -----
// testBSignPrivateKeyFingerprint = "817BB0449D2100E750C18DDBE2BA2D4D3FCC2C6B"
)

type integrityTestOptions struct {
	setupEnvOptions
}

var _ = Describe("Signature", Label("e2e", "signature", "simple"), func() {
	DescribeTable("should sign and verify image manifest and binaries (inhouse)",
		func(ctx SpecContext, testOpts integrityTestOptions) {
			setupEnv(testOpts.setupEnvOptions)

			By("inhouse: starting")
			repoDirname := "repo0"
			fixtureRelPath := "signature/inhouse/state"
			buildReportName := "report0.json"

			By("inhouse: preparing test repo")
			SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

			By("inhouse: building image")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

			extraBuildArgs := []string{
				"--sign-manifest",
				"--sign-key", testKeyBase64,
				"--sign-cert", testCertBase64,
				"--sign-intermediates", testIntermediatesBase64,
				"--sign-elf-files",
			}

			reportProject := report.NewProjectWithReport(werfProject)
			buildOut, buildReport := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath(buildReportName), &werf.WithReportOptions{CommonOptions: werf.CommonOptions{ExtraArgs: extraBuildArgs}})
			Expect(buildOut).To(ContainSubstring("Building stage dockerfile/sign"))
			Expect(buildOut).To(ContainSubstring("Signing ELF files"))

			extraVerifyArgs := []string{
				"--image-ref", buildReport.Images["dockerfile"].DockerImageName,
				"--verify-roots", testRootCertBase64,
				"--verify-manifest",
				"--verify-elf-files",
			}

			By("inhouse: verifying image")
			verifyOut := werfProject.Verify(ctx, &werf.VerifyOptions{CommonOptions: werf.CommonOptions{ExtraArgs: extraVerifyArgs}})
			Expect(verifyOut).To(ContainSubstring("Verifying image (1/1)"))
			Expect(verifyOut).To(ContainSubstring(fmt.Sprintf("Using reference: %s", buildReport.Images["dockerfile"].DockerImageName)))
			Expect(verifyOut).To(ContainSubstring("Manifest signature ... ok"))
			Expect(verifyOut).To(ContainSubstring("ELF files signatures"))
			Expect(verifyOut).To(ContainSubstring("usr/bin/curl ... ok"))
		},
		// TODO: enable when it will be supported
		// XEntry("without repo using Vanilla Docker", integrityTestOptions{setupEnvOptions: setupEnvOptions{
		//	 ContainerBackendMode:        "vanilla-docker",
		//	 WithLocalRepo:               false,
		//	 WithStagedDockerfileBuilder: false,
		// }}),
		Entry("with repo using Vanilla Docker", integrityTestOptions{setupEnvOptions: setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		// TODO: enable when it will be supported
		// XEntry("without repo using BuildKit Docker", integrityTestOptions{setupEnvOptions: setupEnvOptions{
		//	 ContainerBackendMode:        "buildkit-docker",
		//	 WithLocalRepo:               false,
		//	 WithStagedDockerfileBuilder: false,
		// }}),
		Entry("with local repo using BuildKit Docker", integrityTestOptions{setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		// TODO: enable when it will be supported
		// XEntry("without local repo using Native Buildah with rootless isolation", integrityTestOptions{setupEnvOptions{
		// 	ContainerBackendMode:        "native-rootless",
		// 	WithLocalRepo:               false,
		// 	WithStagedDockerfileBuilder: true,
		// }}),
		// XEntry("with local repo using Native Buildah with rootless isolation", integrityTestOptions{setupEnvOptions{
		// 	ContainerBackendMode:        "native-rootless",
		// 	WithLocalRepo:               true,
		// 	WithStagedDockerfileBuilder: true,
		// }}),
		// XEntry("without local repo using Native Buildah with chroot isolation", integrityTestOptions{setupEnvOptions{
		// 	ContainerBackendMode:        "native-chroot",
		// 	WithLocalRepo:               false,
		// 	WithStagedDockerfileBuilder: true,
		// }}),
		// XEntry("with local repo using Native Buildah with chroot isolation", integrityTestOptions{setupEnvOptions{
		// 	ContainerBackendMode:        "native-chroot",
		// 	WithLocalRepo:               true,
		// 	WithStagedDockerfileBuilder: true,
		// }}),
	)

	// TODO: enable when test environment will be ready
	XDescribe("gpg host configuration", func() {
		BeforeEach(func(ctx SpecContext) {
			// Expect(gpg.ImportKey(ctx, testBSignPrivateKeyData)).To(Succeed()) // the key will be imported while building
			Expect(gpg.ImportKey(ctx, testBSignPublicKeyData)).To(Succeed())
		})
		AfterEach(func(ctx SpecContext) {
			// Don't delete keys, because we run tests in parallel. Otherwise, we can use ORDERED for ginkgo.
			// Expect(gpg.SafeDeleteSecretKey(ctx, testBSignKeyFingerprint)).To(Succeed())
			// Expect(gpg.SafeDeleteKey(ctx, testBSignKeyFingerprint)).To(Succeed())
		})
		DescribeTable(
			"should sign and verify binaries (bsign)",
			func(ctx SpecContext, testOpts integrityTestOptions) {
				setupEnv(testOpts.setupEnvOptions)

				By("bsign: starting")
				repoDirname := "repo0"
				fixtureRelPath := "signature/bsign/state"
				buildReportName := "report0.json"

				By("bsign: preparing test repo")
				SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

				By("bsign: building image")
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

				extraBuildArgs := []string{
					"--bsign-elf-files",
					"--elf-pgp-private-key-base64", testBSignPrivateKeyBase64,
				}

				reportProject := report.NewProjectWithReport(werfProject)
				buildOut, buildReport := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath(buildReportName), &werf.WithReportOptions{CommonOptions: werf.CommonOptions{ExtraArgs: extraBuildArgs}})
				Expect(buildOut).To(ContainSubstring("Signing ELF files"))

				extraVerifyArgs := []string{
					"--image-ref", buildReport.Images["dockerfile"].DockerImageName,
					"--verify-roots", testRootCertBase64,
					"--verify-bsign-elf-files",
				}
				By("bsign: verifying image")
				verifyOut := werfProject.Verify(ctx, &werf.VerifyOptions{CommonOptions: werf.CommonOptions{ExtraArgs: extraVerifyArgs}})
				Expect(verifyOut).To(ContainSubstring("Verifying image (1/1)"))
				Expect(verifyOut).To(ContainSubstring(fmt.Sprintf("Using reference: %s", buildReport.Images["dockerfile"].DockerImageName)))
				Expect(verifyOut).To(ContainSubstring("ELF files signatures"))
				Expect(verifyOut).To(ContainSubstring("usr/bin/curl ... ok"))
			},
			// TODO: enable when it will be supported
			// XEntry("without repo using Vanilla Docker", integrityTestOptions{setupEnvOptions: setupEnvOptions{
			//	 ContainerBackendMode:        "vanilla-docker",
			//	 WithLocalRepo:               false,
			//	 WithStagedDockerfileBuilder: false,
			// }}),
			Entry("with repo using Vanilla Docker", integrityTestOptions{setupEnvOptions: setupEnvOptions{
				ContainerBackendMode:        "vanilla-docker",
				WithLocalRepo:               true,
				WithStagedDockerfileBuilder: false,
			}}),
			// TODO: enable when it will be supported
			// XEntry("without repo using BuildKit Docker", integrityTestOptions{setupEnvOptions: setupEnvOptions{
			//	 ContainerBackendMode:        "buildkit-docker",
			//	 WithLocalRepo:               false,
			//	 WithStagedDockerfileBuilder: false,
			// }}),
			Entry("with local repo using BuildKit Docker", integrityTestOptions{setupEnvOptions{
				ContainerBackendMode:        "buildkit-docker",
				WithLocalRepo:               true,
				WithStagedDockerfileBuilder: false,
			}}),
			// TODO: enable when it will be supported
			// XEntry("without local repo using Native Buildah with rootless isolation", integrityTestOptions{setupEnvOptions{
			// 	ContainerBackendMode:        "native-rootless",
			// 	WithLocalRepo:               false,
			// 	WithStagedDockerfileBuilder: true,
			// }}),
			// XEntry("with local repo using Native Buildah with rootless isolation", integrityTestOptions{setupEnvOptions{
			// 	ContainerBackendMode:        "native-rootless",
			// 	WithLocalRepo:               true,
			// 	WithStagedDockerfileBuilder: true,
			// }}),
			// XEntry("without local repo using Native Buildah with chroot isolation", integrityTestOptions{setupEnvOptions{
			// 	ContainerBackendMode:        "native-chroot",
			// 	WithLocalRepo:               false,
			// 	WithStagedDockerfileBuilder: true,
			// }}),
			// XEntry("with local repo using Native Buildah with chroot isolation", integrityTestOptions{setupEnvOptions{
			// 	ContainerBackendMode:        "native-chroot",
			// 	WithLocalRepo:               true,
			// 	WithStagedDockerfileBuilder: true,
			// }}),
		)
	})
})
