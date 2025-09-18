package e2e_build_test

import (
	_ "embed"
	"encoding/base64"

	"github.com/deckhouse/delivery-kit-sdk/pkg/signature/elf/inhouse"
	"github.com/deckhouse/delivery-kit-sdk/pkg/signature/image"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/utils"
	"github.com/werf/werf/v2/test/pkg/utils/gpg"
	"github.com/werf/werf/v2/test/pkg/werf"
)

var (
	// ----- Inhouse -----
	//go:embed _fixtures/integrity_control/inhouse/keys/sigstore_959497322.pem.key
	testKeyData []byte
	//go:embed _fixtures/integrity_control/inhouse/keys/sigstore_1666162742.pem.crt
	testCertData []byte
	//go:embed _fixtures/integrity_control/inhouse/keys/sigstore_chain_3247019714.pem.crt
	testChainData []byte
	//go:embed _fixtures/integrity_control/inhouse/keys/sigstore_intermediates_1698313569.pem.crt
	testIntermediatesData []byte
	//go:embed _fixtures/integrity_control/inhouse/keys/sigstore_root_1959509881.pem.crt
	testRootCertData []byte

	// ----- BSign -----
	//go:embed _fixtures/integrity_control/bsign/keys/rsa_4096_private.pgp
	testBSignPrivateKeyData []byte
	//go:embed _fixtures/integrity_control/bsign/keys/rsa_4096_public.pgp
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

var _ = Describe("Integrity control", Label("e2e", "integrity", "simple"), func() {
	DescribeTable("should sign and verify image manifest and binaries (inhouse), and annotate image with verity",
		func(ctx SpecContext, testOpts integrityTestOptions) {
			setupEnv(testOpts.setupEnvOptions)

			By("inhouse: starting")
			repoDirname := "repo0"
			fixtureRelPath := "integrity_control/inhouse/state"
			buildReportName := "report0.json"

			By("inhouse: preparing test repo")
			SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

			By("inhouse: building image")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

			extraArgs := []string{
				"--sign-manifest",
				"--sign-key", testKeyBase64,
				"--sign-cert", testCertBase64,
				"--sign-intermediates", testIntermediatesBase64,
				"--sign-elf-files",
				"--annotate-layers-with-dm-verity-root-hash",
			}

			buildOut, buildReport := werfProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath(buildReportName), &werf.BuildWithReportOptions{CommonOptions: werf.CommonOptions{ExtraArgs: extraArgs}})
			Expect(buildOut).To(ContainSubstring("Building stage dockerfile/sign"))
			Expect(buildOut).To(ContainSubstring("Signing ELF files"))

			By("inhouse: loading image and manifest from registry")
			img := loadImageFromRegistry(buildReport.Images["dockerfile"].DockerImageName)
			manifest, err := img.Manifest()
			Expect(err).To(Succeed())

			// ----- Verification -----
			// Debugging using docker:
			// docker manifest inspect --insecure localhost:38903/werf-test-none-31906--22cc85d6:b0eacfdc7b53f2db71f9e76ff7c08395f0c9d993ecdfd763a9159e7a-1757065424241

			// ----- Manifest verification -----

			By("inhouse: verify image manifest signature")
			Expect(manifest.Annotations).To(HaveKeyWithValue("io.deckhouse.delivery-kit.signature", Not(BeEmpty())))
			Expect(manifest.Annotations).To(HaveKeyWithValue("io.deckhouse.delivery-kit.cert", testCertBase64))
			Expect(manifest.Annotations).To(HaveKeyWithValue("io.deckhouse.delivery-kit.chain", testIntermediatesBase64))

			err = image.VerifyImageManifestSignature(ctx, []string{testRootCertBase64}, manifest)
			Expect(err).To(Succeed())

			By("inhouse: verify image manifest dm-verity annotation")
			// Expect(manifest.Annotations).To(HaveKeyWithValue("io.deckhouse.delivery-kit.build-timestamp", Not(BeEmpty())))
			// Expect(manifest.Annotations).To(HaveKeyWithValue("io.deckhouse.delivery-kit.dm-verity-root-hash", Not(BeEmpty())))

			// ----- ELF files verification -----

			By("inhouse: verify ELF's signatures")
			containerFilePaths := []string{
				"/usr/bin/curl",
			}

			tmpLocalPaths := utils.ExtractFilesFromImage(SuiteData.TmpDir, img, containerFilePaths)
			Expect(tmpLocalPaths).To(HaveLen(len(containerFilePaths)))

			for _, tmpLocalPath := range tmpLocalPaths {
				err = inhouse.Verify(ctx, []string{testRootCertBase64}, tmpLocalPath)
				Expect(err).To(Succeed())
			}
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

	Describe("gpg host configuration", func() {
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
				fixtureRelPath := "integrity_control/bsign/state"
				buildReportName := "report0.json"

				By("bsign: preparing test repo")
				SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

				By("bsign: building image")
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

				extraArgs := []string{
					"--bsign-elf-files",
					"--elf-pgp-private-key-base64", testBSignPrivateKeyBase64,
				}

				buildOut, _ := werfProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath(buildReportName), &werf.BuildWithReportOptions{CommonOptions: werf.CommonOptions{ExtraArgs: extraArgs}})
				Expect(buildOut).To(ContainSubstring("Signing ELF files"))

				// By("bsign: loading image and manifest from registry")
				// img := loadImageFromRegistry(buildReport.Images["dockerfile"].DockerImageName)

				// By("bsign: verify ELF's signatures")
				// containerFilePaths := []string{
				// 	"/usr/bin/curl",
				// }

				// tmpLocalPaths := utils.ExtractFilesFromImage(SuiteData.TmpDir, img, containerFilePaths)
				// Expect(tmpLocalPaths).To(HaveLen(len(containerFilePaths)))

				// for _, tmpLocalPath := range tmpLocalPaths {
				// 	// https://manpages.debian.org/testing/bsign/bsign.1.en.html

				// 	cmd := exec.CommandContextCancellation(ctx, "bsign", "--verify", tmpLocalPath)
				// 	output, err := cmd.CombinedOutput()
				// 	Expect(err).To(Succeed(), "'bsign --verify %s' output: %s", tmpLocalPath, output)
				// }
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

func loadImageFromRegistry(imageName string) v1.Image {
	ref, err := name.ParseReference(imageName)
	Expect(err).To(Succeed())
	desc, err := remote.Get(ref)
	Expect(err).To(Succeed())
	img, err := desc.Image()
	Expect(err).To(Succeed())
	return img
}
