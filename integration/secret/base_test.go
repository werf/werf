// +build integration

package secret_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/flant/werf/integration/utils"
)

var _ = It("should generate secret key", func() {
	utils.RunSucceedCommand(
		testDirPath,
		werfBinPath,
		"helm", "secret", "generate-secret-key",
	)
})

var _ = It("should rotate secret key", func() {
	utils.CopyIn(fixturePath("rotate_secret_key"), testDirPath)

	res, err := ioutil.ReadFile(filepath.Join(testDirPath, ".werf_secret_key"))
	Ω(err).ShouldNot(HaveOccurred())

	oldSecretKey := strings.TrimSpace(string(res))
	Ω(os.Remove(filepath.Join(testDirPath, ".werf_secret_key"))).Should(Succeed())

	output := utils.SucceedCommandOutput(
		testDirPath,
		werfBinPath,
		"helm", "secret", "generate-secret-key",
	)

	newSecretKey := strings.TrimSpace(output)

	cmd := exec.Command(werfBinPath, "helm", "secret", "rotate-secret-key")
	cmd.Dir = testDirPath
	cmd.Env = append([]string{
		fmt.Sprintf("WERF_SECRET_KEY=%s", newSecretKey),
		fmt.Sprintf("WERF_OLD_SECRET_KEY=%s", oldSecretKey),
	}, os.Environ()...)

	res, err = cmd.Output()
	_, _ = fmt.Fprintf(GinkgoWriter, string(res))
	Ω(err).ShouldNot(HaveOccurred())

	filesShouldBeRegenerated := []string{".helm/secret/test", ".helm/secret/subdir/test", ".helm/secret-values.yaml"}
	for _, path := range filesShouldBeRegenerated {
		Ω(string(res)).Should(ContainSubstring(fmt.Sprintf("Regenerating file '%s'", filepath.FromSlash(path))))
	}
})

var _ = Describe("helm secret encrypt/decrypt", func() {
	var secret = "test"
	var encryptedSecret = "1000ceeb30457f57eb67a2dfecd65c563417f4ae06167fb21be60549d247bf388165"

	BeforeEach(func() {
		if runtime.GOOS == "windows" {
			Skip("skip on windows")
		}

		utils.CopyIn(fixturePath("default"), testDirPath)
	})

	It("should be encrypted", func() {
		output := utils.SucceedCommandOutput(
			testDirPath,
			"bash",
			"-c", "echo "+secret+" | "+werfBinPath+" helm secret encrypt",
		)

		result := strings.TrimSpace(output)

		output = utils.SucceedCommandOutput(
			testDirPath,
			"bash",
			"-c", "echo "+result+" | "+werfBinPath+" helm secret decrypt",
		)

		result = strings.TrimSpace(output)

		Ω(result).Should(BeEquivalentTo(secret))
	})

	It("should be decrypted", func() {
		output := utils.SucceedCommandOutput(
			testDirPath,
			"bash",
			"-c", "echo "+encryptedSecret+" | "+werfBinPath+" helm secret decrypt",
		)

		result := strings.TrimSpace(output)

		Ω(result).Should(BeEquivalentTo(secret))
	})
})
