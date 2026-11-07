package e2e_helm_secret_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/test/pkg/utils"
)

var _ = ginkgo.Describe("Helm secret", ginkgo.Label("e2e", "helm-secret", "simple"), func() {
	var projectDir string
	var secretKey string

	ginkgo.BeforeEach(func(ctx ginkgo.SpecContext) {
		projectDir = SuiteData.GetProjectWorktree(SuiteData.ProjectName)
		gomega.Expect(os.MkdirAll(projectDir, os.ModePerm)).To(gomega.Succeed())

		secretKey = generateSecretKey(ctx)
		SuiteData.Stubs.SetEnv("WERF_SECRET_KEY", secretKey)
	})

	ginkgo.It("should encrypt stdin into a file and decrypt it back byte for byte", func(ctx ginkgo.SpecContext) {
		secret := "first line\nsecond line\n"
		encryptedPath := filepath.Join(projectDir, "encrypted")

		utils.RunCommandWithOptions(ctx, projectDir, SuiteData.WerfBinPath, []string{"helm", "secret", "encrypt", "-o", encryptedPath}, utils.RunCommandOptions{
			ToStdin:       secret,
			ShouldSucceed: true,
		})
		gomega.Expect(encryptedPath).To(gomega.BeAnExistingFile())

		encrypted := readFile(encryptedPath)
		gomega.Expect(encrypted).NotTo(gomega.ContainSubstring(secret))
		gomega.Expect(encrypted).To(gomega.HaveSuffix("\n"))

		decrypted, _ := utils.RunCommandWithOptions(ctx, projectDir, SuiteData.WerfBinPath, []string{"helm", "secret", "decrypt"}, utils.RunCommandOptions{
			ToStdin:       encrypted,
			ShouldSucceed: true,
		})
		gomega.Expect(string(decrypted)).To(gomega.Equal(secret))
	})

	ginkgo.It("should fail to decrypt with a wrong key and keep the output file untouched", func(ctx ginkgo.SpecContext) {
		encryptedPath := filepath.Join(projectDir, "encrypted")

		utils.RunCommandWithOptions(ctx, projectDir, SuiteData.WerfBinPath, []string{"helm", "secret", "encrypt", "-o", encryptedPath}, utils.RunCommandOptions{
			ToStdin:       "secret data\n",
			ShouldSucceed: true,
		})
		encrypted := readFile(encryptedPath)

		output, err := utils.RunCommandWithOptions(ctx, projectDir, SuiteData.WerfBinPath, []string{"helm", "secret", "decrypt", "-o", encryptedPath}, utils.RunCommandOptions{
			ExtraEnv: []string{fmt.Sprintf("WERF_SECRET_KEY=%s", generateSecretKey(ctx))},
			ToStdin:  encrypted,
		})
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(string(output)).To(gomega.ContainSubstring("check encryption key and data"))
		gomega.Expect(readFile(encryptedPath)).To(gomega.Equal(encrypted))
	})

	ginkgo.It("should encrypt values file scalars and decrypt them back", func(ctx ginkgo.SpecContext) {
		values := "app:\n  db:\n    password: secret-password\n  replicas: 2\n"
		valuesPath := filepath.Join(projectDir, "values.yaml")
		encryptedPath := filepath.Join(projectDir, "values.enc.yaml")
		gomega.Expect(os.WriteFile(valuesPath, []byte(values), 0o644)).To(gomega.Succeed())

		utils.RunSucceedCommand(ctx, projectDir, SuiteData.WerfBinPath, "helm", "secret", "values", "encrypt", valuesPath, "-o", encryptedPath)

		encrypted := readFile(encryptedPath)
		gomega.Expect(encrypted).NotTo(gomega.ContainSubstring("secret-password"))
		gomega.Expect(encrypted).To(gomega.ContainSubstring("    password: "))

		decrypted := utils.SucceedCommandOutputString(ctx, projectDir, SuiteData.WerfBinPath, "helm", "secret", "values", "decrypt", encryptedPath)
		gomega.Expect(decrypted).To(gomega.Equal(values))
	})
})
