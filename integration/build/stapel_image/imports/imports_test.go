package imports_test

import (
	"strings"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/werf/integration/utils"
	"github.com/werf/werf/integration/utils/liveexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func werfBuild(dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(dir, SuiteData.WerfBinPath, opts, utils.WerfBinArgs(append([]string{"build"}, extraArgs...)...)...)
}

func werfRunOutput(dir string, extraArgs ...string) string {
	output, _ := utils.RunCommandWithOptions(
		dir, SuiteData.WerfBinPath,
		append([]string{"run", "--"}, extraArgs...),
		utils.RunCommandOptions{ShouldSucceed: true},
	)
	return string(output)
}

func werfPurge(dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(dir, SuiteData.WerfBinPath, opts, utils.WerfBinArgs(append([]string{"purge"}, extraArgs...)...)...)
}

var _ = Describe("Stapel imports", func() {
	BeforeEach(func() {
		Expect(kube.Init(kube.InitOptions{})).To(Succeed())
	})

	Context("importing files and directories from artifact", func() {
		AfterEach(func() {
			werfPurge("imports_app1-003", liveexec.ExecCommandOptions{}, "--force")
		})

		It("should allow importing files and directories, optionally rename files and directories and merge directories", func() {
			gotNoSuchFileError := false
			Expect(werfBuild("imports_app1-001", liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					if strings.Index(line, "/myartifact/no-such-dir") != -1 && strings.Index(line, "No such file or directory") != -1 {
						gotNoSuchFileError = true
					}
				},
			})).NotTo(Succeed())
			Expect(gotNoSuchFileError).To(BeTrue())

			gotNoSuchFileError = false
			Expect(werfBuild("imports_app1-002", liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					if strings.Index(line, "/myartifact/file-no-such-file") != -1 && strings.Index(line, "No such file or directory") != -1 {
						gotNoSuchFileError = true
					}
				},
			})).NotTo(Succeed())
			Expect(gotNoSuchFileError).To(BeTrue())

			Expect(werfBuild("imports_app1-003", liveexec.ExecCommandOptions{})).To(Succeed())

			Expect(werfRunOutput("imports_app1-003", "cat", "/usr/local/FILE")).To(ContainSubstring("FILE\n"))
			Expect(werfRunOutput("imports_app1-003", "cat", "/usr/locallll")).To(ContainSubstring("FILE\n"))
			Expect(werfRunOutput("imports_app1-003", "cat", "/usr/newlocal/file")).To(ContainSubstring("GOGOGO\n"))
			Expect(werfRunOutput("imports_app1-003", "cat", "/usr/newlocal/a/b/FILE")).To(ContainSubstring("FILE\n"))
			Expect(werfRunOutput("imports_app1-003", "cat", "/usr/share/file")).To(ContainSubstring("GOGOGO\n"))
			Expect(werfRunOutput("imports_app1-003", "cat", "/usr/share/a/b/FILE")).To(ContainSubstring("FILE\n"))
			Expect(werfRunOutput("imports_app1-003", "ls", "/usr/share/apk")).To(ContainSubstring("keys\n"))
			Expect(werfRunOutput("imports_app1-003", "cat", "/file2")).To(ContainSubstring("GOGOGO\n"))
			Expect(werfRunOutput("imports_app1-003", "cat", "/file")).To(ContainSubstring("GOGOGO\n"))
		})
	})

	Context("caching by import source checksum", func() {
		AfterEach(func() {
			utils.RunSucceedCommand(
				"import_metadata",
				SuiteData.WerfBinPath,
				"purge",
			)
		})

		It("should cache image when import source checksum was not changed", func() {
			SuiteData.Stubs.SetEnv("WERF_CONFIG", "werf_1.yaml")

			utils.RunSucceedCommand(
				"import_metadata",
				SuiteData.WerfBinPath,
				"build",
			)

			lastStageImageNameAfterFirstBuild := utils.GetBuiltImageLastStageImageName("import_metadata", SuiteData.WerfBinPath, "image")

			SuiteData.Stubs.SetEnv("WERF_CONFIG", "werf_2.yaml")

			utils.RunSucceedCommand(
				"import_metadata",
				SuiteData.WerfBinPath,
				"build",
			)

			lastStageImageNameAfterSecondBuild := utils.GetBuiltImageLastStageImageName("import_metadata", SuiteData.WerfBinPath, "image")

			Ω(lastStageImageNameAfterFirstBuild).Should(Equal(lastStageImageNameAfterSecondBuild))
		})

		It("should rebuild image when import source checksum was changed", func() {
			SuiteData.Stubs.SetEnv("WERF_CONFIG", "werf_1.yaml")

			utils.RunSucceedCommand(
				"import_metadata",
				SuiteData.WerfBinPath,
				"build",
			)

			lastStageImageNameAfterFirstBuild := utils.GetBuiltImageLastStageImageName("import_metadata", SuiteData.WerfBinPath, "image")

			SuiteData.Stubs.SetEnv("WERF_CONFIG", "werf_3.yaml")

			utils.RunSucceedCommand(
				"import_metadata",
				SuiteData.WerfBinPath,
				"build",
			)

			lastStageImageNameAfterSecondBuild := utils.GetBuiltImageLastStageImageName("import_metadata", SuiteData.WerfBinPath, "image")

			Ω(lastStageImageNameAfterFirstBuild).ShouldNot(Equal(lastStageImageNameAfterSecondBuild))
		})
	})
})
