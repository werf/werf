package common_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/contback"
	"github.com/werf/werf/v2/test/pkg/utils"
	"github.com/werf/werf/v2/test/pkg/werf"
)

type testOptions struct {
	ContainerBackendMode        string
	State                       string
	SSH                         bool
	WithStagedDockerfileBuilder bool
}

var _ = Describe("build with secrets and ssh mounts", Label("integration", "build", "with secrets"), func() {
	DescribeTable("should succeed",
		func(ctx SpecContext, testOpts testOptions) {
			setupEnv(testOpts)
			_, err := contback.NewContainerBackend(testOpts.ContainerBackendMode)
			if err == contback.ErrRuntimeUnavailable {
				Skip(err.Error())
			} else if err != nil {
				Fail(err.Error())
			}

			runOpts := &werf.BuildOptions{}
			repoDirname := "repo0"
			fixtureRelPath := "build_with_secrets"
			var keyFile string
			if testOpts.SSH {
				By(fmt.Sprintf("%s: generating sekret key for ssh agent", testOpts.State))
				fixtureRelPath = "build_with_ssh"
				keyFile, err = generateSSHKey(fmt.Sprintf("id_rsa_werf_test_%s", utils.GetRandomString(5)), 2048)
				Expect(err).NotTo(HaveOccurred())
				runOpts.ExtraArgs = append(runOpts.ExtraArgs, "--ssh-key", keyFile)
			}

			defer func() {
				if len(keyFile) > 0 {
					utils.DeleteFile(keyFile)
				}
			}()

			Expect(err).NotTo(HaveOccurred())

			By(fmt.Sprintf("%s: preparing test repo", testOpts.State))
			SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

			By(fmt.Sprintf("%s: building images", testOpts.State))
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
			if testOpts.ContainerBackendMode == "vanilla-docker" {
				runOpts.ExtraArgs = append([]string{"stapel-shell"}, runOpts.ExtraArgs...)
			}
			buildOut := werfProject.Build(ctx, runOpts)
			Expect(buildOut).To(ContainSubstring("Building stage"))
			Expect(buildOut).NotTo(ContainSubstring("Use previously built image"))
		},
		Entry("with Vanilla Docker", testOptions{
			ContainerBackendMode: "vanilla-docker",
		}),
		Entry("with BuildKit Docker", testOptions{
			ContainerBackendMode: "buildkit-docker",
			SSH:                  false,
		}),
		// Entry("with Native Buildah with rootless isolation", testOptions{
		//	ContainerBackendMode: "native-rootless",
		// }),
		Entry("with Native Buildah with chroot isolation", testOptions{
			ContainerBackendMode: "native-chroot",
			SSH:                  false,
		}),
		Entry("with Native Buildah with chroot isolation STAGED", testOptions{
			ContainerBackendMode:        "native-chroot",
			WithStagedDockerfileBuilder: true,
			SSH:                         false,
		}),
		Entry("with Vanilla Docker with SSH", testOptions{
			ContainerBackendMode: "vanilla-docker",
			SSH:                  true,
		}),
		Entry("with BuildKit Docker with SSH", testOptions{
			ContainerBackendMode: "buildkit-docker",
			SSH:                  true,
		}),
		Entry("with Native Buildah with rootless isolation with SSH", testOptions{
			SSH:                  true,
			ContainerBackendMode: "native-rootless",
		}),
		Entry("with Native Buildah with chroot isolation with SSH", testOptions{
			SSH:                  true,
			ContainerBackendMode: "native-chroot",
		}),
		Entry("with Native Buildah with chroot isolation with SSH STAGED", testOptions{
			SSH:                         true,
			ContainerBackendMode:        "native-chroot",
			WithStagedDockerfileBuilder: true,
		}),
	)
})

func generateSSHKey(filename string, bits int) (string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", fmt.Errorf("error generating RSA key: %w", err)
	}

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	file, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	if err := pem.Encode(file, privateKeyPEM); err != nil {
		return "", fmt.Errorf("error encoding PEM: %w", err)
	}

	return filepath.Abs(file.Name())
}
