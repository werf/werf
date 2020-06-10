package imports_test

import (
	"strings"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/werf/pkg/testing/utils"
	"github.com/werf/werf/pkg/testing/utils/liveexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func werfBuild(dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(dir, werfBinPath, opts, utils.WerfBinArgs(append([]string{"build", "-s", ":local"}, extraArgs...)...)...)
}

func werfRunOutput(dir string, extraArgs ...string) string {
	output, _ := utils.RunCommandWithOptions(
		dir, werfBinPath,
		append([]string{"run", "-s", ":local", "--"}, extraArgs...),
		utils.RunCommandOptions{ShouldSucceed: true},
	)
	return string(output)
}

func werfPurge(dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(dir, werfBinPath, opts, utils.WerfBinArgs(append([]string{"stages", "purge", "-s", ":local"}, extraArgs...)...)...)
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
})
