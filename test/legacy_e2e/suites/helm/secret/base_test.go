package secret_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/utils"
)

var _ = It("should generate secret key", func(ctx SpecContext) {
	utils.RunSucceedCommand(ctx, "", SuiteData.WerfBinPath, "helm", "secret", "generate-secret-key")
})

var _ = It("should rotate secret key", func(ctx SpecContext) {
	SuiteData.CommitProjectWorktree(ctx, SuiteData.ProjectName, utils.FixturePath("rotate_secret_key"), "initial commit")

	res, err := ioutil.ReadFile(filepath.Join(SuiteData.GetProjectWorktree(SuiteData.ProjectName), ".werf_secret_key"))
	Expect(err).ShouldNot(HaveOccurred())

	oldSecretKey := strings.TrimSpace(string(res))
	Expect(os.Remove(filepath.Join(SuiteData.GetProjectWorktree(SuiteData.ProjectName), ".werf_secret_key"))).Should(Succeed())

	output := utils.SucceedCommandOutputString(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "helm", "secret", "generate-secret-key")

	newSecretKey := strings.TrimSpace(output)

	cmd := exec.CommandContext(ctx, SuiteData.WerfBinPath, "helm", "secret", "rotate-secret-key")
	cmd.Dir = SuiteData.GetProjectWorktree(SuiteData.ProjectName)
	cmd.Env = append([]string{
		fmt.Sprintf("WERF_SECRET_KEY=%s", newSecretKey),
		fmt.Sprintf("WERF_OLD_SECRET_KEY=%s", oldSecretKey),
	}, os.Environ()...)

	res, err = cmd.Output()
	_, _ = fmt.Fprintf(GinkgoWriter, string(res))
	Expect(err).ShouldNot(HaveOccurred())

	filesShouldBeRegenerated := []string{".helm/secret/test", ".helm/secret/subdir/test", ".helm/secret-values.yaml"}
	for _, path := range filesShouldBeRegenerated {
		Expect(string(res)).Should(ContainSubstring(fmt.Sprintf("Regenerating file %q", filepath.FromSlash(path))))
	}
})

var _ = Describe("helm secret encrypt/decrypt", func() {
	secret := "test"
	encryptedSecret := "1000ceeb30457f57eb67a2dfecd65c563417f4ae06167fb21be60549d247bf388165"

	BeforeEach(func(ctx SpecContext) {
		SuiteData.CommitProjectWorktree(ctx, SuiteData.ProjectName, utils.FixturePath("default"), "initial commit")
	})

	It("should be encrypted", func(ctx SpecContext) {
		resultData, _ := utils.RunCommandWithOptions(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, []string{"helm", "secret", "encrypt"}, utils.RunCommandOptions{
			ToStdin:       secret,
			ShouldSucceed: true,
		})

		result := string(bytes.TrimSpace(resultData))

		resultData, _ = utils.RunCommandWithOptions(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, []string{"helm", "secret", "decrypt"}, utils.RunCommandOptions{
			ToStdin:       result,
			ShouldSucceed: true,
		})

		result = string(bytes.TrimSpace(resultData))

		Expect(result).Should(BeEquivalentTo(secret))
	})

	It("should be decrypted", func(ctx SpecContext) {
		resultData, _ := utils.RunCommandWithOptions(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, []string{"helm", "secret", "decrypt"}, utils.RunCommandOptions{
			ToStdin:       encryptedSecret,
			ShouldSucceed: true,
		})

		result := string(bytes.TrimSpace(resultData))

		Expect(result).Should(BeEquivalentTo(secret))
	})
})
