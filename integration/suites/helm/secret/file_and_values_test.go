package secret_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/utils"
)

var _ = Describe("helm secret file/values encrypt/decrypt", func() {
	decryptAndCheckFileOrValues := func(secretType, fileToProcess string, withPipe bool) {
		if withPipe {
			runSucceedCommandWithFileDataOnStdin([]string{"helm", "secret", secretType, "decrypt", "-o", "result"}, fileToProcess)
		} else {
			utils.RunSucceedCommand(
				SuiteData.GetProjectWorktree(SuiteData.ProjectName),
				SuiteData.WerfBinPath,
				"helm", "secret", secretType, "decrypt", fileToProcess, "-o", "result",
			)
		}

		fileContentsShouldBeEqual("result", "secret")
	}

	decryptItBody := func(secretType string, withPipe bool) {
		SuiteData.CommitProjectWorktree(SuiteData.ProjectName, utils.FixturePath(secretType), "initial commit")
		decryptAndCheckFileOrValues(secretType, "encrypted_secret", withPipe)
	}

	encryptItBody := func(secretType string, withPipe bool) {
		SuiteData.CommitProjectWorktree(SuiteData.ProjectName, utils.FixturePath(secretType), "initial commit")

		if withPipe {
			runSucceedCommandWithFileDataOnStdin([]string{"helm", "secret", secretType, "encrypt", "-o", "result"}, "secret")
		} else {
			utils.RunSucceedCommand(
				SuiteData.GetProjectWorktree(SuiteData.ProjectName),
				SuiteData.WerfBinPath,
				"helm", "secret", secretType, "encrypt", "secret", "-o", "result",
			)
		}

		decryptAndCheckFileOrValues(secretType, "result", withPipe)
	}

	editItBody := func(secretType string) {
		if runtime.GOOS == "windows" {
			Skip("skip on windows")
		}

		SuiteData.CommitProjectWorktree(SuiteData.ProjectName, utils.FixturePath(secretType), "initial commit")

		_, _ = utils.RunCommandWithOptions(
			SuiteData.GetProjectWorktree(SuiteData.ProjectName),
			SuiteData.WerfBinPath,
			[]string{"helm", "secret", secretType, "edit", "result"},
			utils.RunCommandOptions{
				ExtraEnv:      []string{"EDITOR=./editor.sh"},
				ShouldSucceed: true,
			},
		)

		decryptAndCheckFileOrValues(secretType, "result", false)
	}

	_ = DescribeTable("edit", editItBody,
		Entry("secret file", "file"),
		Entry("secret file", "values"))

	_ = DescribeTable("encryption", encryptItBody,
		Entry("secret file", "file", false),
		Entry("secret file (pipe)", "file", true),
		Entry("secret values", "values", false),
		Entry("secret values (pipe)", "values", true))

	_ = DescribeTable("decryption", decryptItBody,
		Entry("secret file", "file", false),
		Entry("secret file (pipe)", "file", true),
		Entry("secret values", "values", false),
		Entry("secret values (pipe)", "values", true))
})

func fileContentsShouldBeEqual(path1, path2 string) {
	data1, err := ioutil.ReadFile(filepath.Join(SuiteData.GetProjectWorktree(SuiteData.ProjectName), path1))
	Ω(err).ShouldNot(HaveOccurred())

	data2, err := ioutil.ReadFile(filepath.Join(SuiteData.GetProjectWorktree(SuiteData.ProjectName), path2))
	Ω(err).ShouldNot(HaveOccurred())

	data1 = bytes.ReplaceAll(data1, []byte(utils.LineBreak), []byte("\n"))
	data2 = bytes.ReplaceAll(data2, []byte(utils.LineBreak), []byte("\n"))

	_, _ = fmt.Fprintf(GinkgoWriter, "=== %s ===\n", filepath.Join(SuiteData.GetProjectWorktree(SuiteData.ProjectName), path1))
	_, _ = fmt.Fprintf(GinkgoWriter, string(data1))
	_, _ = fmt.Fprintf(GinkgoWriter, "=== %s ===\n", filepath.Join(SuiteData.GetProjectWorktree(SuiteData.ProjectName), path1))

	_, _ = fmt.Fprintf(GinkgoWriter, "=== %s ===\n", filepath.Join(SuiteData.GetProjectWorktree(SuiteData.ProjectName), path2))
	_, _ = fmt.Fprintf(GinkgoWriter, string(data2))
	_, _ = fmt.Fprintf(GinkgoWriter, "=== %s ===\n", filepath.Join(SuiteData.GetProjectWorktree(SuiteData.ProjectName), path2))

	Ω(bytes.Equal(data1, data2)).Should(BeTrue())
}

func runSucceedCommandWithFileDataOnStdin(werfArgs []string, secretFileName string) {
	data, err := ioutil.ReadFile(filepath.Join(SuiteData.GetProjectWorktree(SuiteData.ProjectName), secretFileName))

	Ω(err).ShouldNot(HaveOccurred())

	_, _ = utils.RunCommandWithOptions(
		SuiteData.GetProjectWorktree(SuiteData.ProjectName),
		SuiteData.WerfBinPath,
		werfArgs,
		utils.RunCommandOptions{
			ToStdin:       string(data),
			ShouldSucceed: true,
		},
	)
}
