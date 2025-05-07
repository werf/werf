package ansible_test

import (
	"os"
	"runtime"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/utils"
	"github.com/werf/werf/v2/test/pkg/utils/liveexec"
)

var _ = Describe("Stapel builder with ansible", func() {
	Context("when building image based on alpine, ubuntu or centos", func() {
		AfterEach(func(ctx SpecContext) {
			werfHostPurge(ctx, "general", liveexec.ExecCommandOptions{}, "--force")
			os.RemoveAll("general/.git")
			os.RemoveAll("general_repo")
		})

		It("should successfully build image using arbitrary ansible modules", func(ctx SpecContext) {
			Expect(utils.SetGitRepoState(ctx, "general", "general_repo", "initial commit")).To(Succeed())
			Expect(werfBuild(ctx, "general", liveexec.ExecCommandOptions{})).To(Succeed())
		})
	})

	Context("when building stapel image based on centos 8", func() {
		AfterEach(func(ctx SpecContext) {
			werfHostPurge(ctx, "yum", liveexec.ExecCommandOptions{}, "--force")
			os.RemoveAll("yum/.git")
			os.RemoveAll("yum_repo")
		})

		It("successfully installs packages using yum module", func(ctx SpecContext) {
			Skip("FIXME https://github.com/werf/werf/issues/1983")
			Expect(utils.SetGitRepoState(ctx, "yum", "yum_repo", "initial commit")).To(Succeed())
			Expect(werfBuild(ctx, "yum", liveexec.ExecCommandOptions{})).To(Succeed())
		})
	})

	Context("when become_user task option used", func() {
		AfterEach(func(ctx SpecContext) {
			werfHostPurge(ctx, "become_user", liveexec.ExecCommandOptions{}, "--force")
			os.RemoveAll("become_user/.git")
			os.RemoveAll("become_user_repo")
		})

		It("successfully installs packages using yum module", func(ctx SpecContext) {
			Skip("FIXME https://github.com/werf/werf/issues/1806")
			Expect(utils.SetGitRepoState(ctx, "become_user", "become_user_repo", "initial commit")).To(Succeed())
			Expect(werfBuild(ctx, "become_user", liveexec.ExecCommandOptions{})).To(Succeed())
		})
	})

	Context("when using apt_key module used (1)", func() {
		AfterEach(func(ctx SpecContext) {
			werfHostPurge(ctx, "apt_key1-001", liveexec.ExecCommandOptions{}, "--force")
			os.RemoveAll("apt_key1-001/.git")
			os.RemoveAll("apt_key1_repo")
		})

		It("should fail to install package without a key and succeed with the key", func(ctx SpecContext) {
			Skip("https://github.com/werf/werf/issues/2000")

			Expect(utils.SetGitRepoState(ctx, "apt_key1-001", "apt_key1_repo", "initial commit")).To(Succeed())

			gotNoPubkey := false
			Expect(werfBuild(ctx, "apt_key1-001", liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					if strings.Contains(line, `public key is not available: NO_PUBKEY`) {
						gotNoPubkey = true
					}
				},
			})).NotTo(Succeed())
			Expect(gotNoPubkey).To(BeTrue())

			gotPackageInstallDone := false
			Expect(werfBuild(ctx, "apt_key1-002", liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					if strings.Contains(line, `apt 'Install package from new repository' [clickhouse-client]`) {
						gotPackageInstallDone = true
					}
					Expect(line).NotTo(MatchRegexp(`apt 'Install package from new repository' \[clickhouse-client\] \(".*" seconds\) FAILED`))
				},
			})).To(Succeed())
			Expect(gotPackageInstallDone).To(BeTrue())
		})
	})

	Context("when using apt_key module used (2)", func() {
		AfterEach(func(ctx SpecContext) {
			werfHostPurge(ctx, "apt_key2", liveexec.ExecCommandOptions{}, "--force")
			os.RemoveAll("apt_key2/.git")
			os.RemoveAll("apt_key2_repo")
		})

		It("should fail to install package without a key and succeed with the key", func(ctx SpecContext) {
			Skip("https://github.com/werf/werf/issues/2000")

			Expect(utils.SetGitRepoState(ctx, "apt_key2", "apt_key2_repo", "initial commit")).To(Succeed())
			Expect(werfBuild(ctx, "apt_key2", liveexec.ExecCommandOptions{})).To(Succeed())
		})
	})

	Context("when apt-mark from apt module used (https://github.com/werf/werf/issues/1820)", func() {
		AfterEach(func(ctx SpecContext) {
			werfHostPurge(ctx, "apt_mark_panic_1820", liveexec.ExecCommandOptions{}, "--force")
			os.RemoveAll("apt_mark_panic_1820/.git")
			os.RemoveAll("apt_mark_panic_1820_repo")
		})

		It("should not panic in all supported ubuntu versions", func(ctx SpecContext) {
			Expect(utils.SetGitRepoState(ctx, "apt_mark_panic_1820", "apt_mark_panic_1820_repo", "initial commit")).To(Succeed())
			Expect(werfBuild(ctx, "apt_mark_panic_1820", liveexec.ExecCommandOptions{})).To(Succeed())
		})
	})

	Context("when using yarn module to install nodejs packages", func() {
		BeforeEach(func() {
			os.RemoveAll("yarn/.git")
			os.RemoveAll("yarn_repo")
		})

		AfterEach(func(ctx SpecContext) {
			werfHostPurge(ctx, "yarn", liveexec.ExecCommandOptions{}, "--force")
			os.RemoveAll("yarn/.git")
			os.RemoveAll("yarn_repo")
		})

		It("should install packages successfully", func(ctx SpecContext) {
			Expect(utils.SetGitRepoState(ctx, "yarn", "yarn_repo", "initial commit")).To(Succeed())
			Expect(werfBuild(ctx, "yarn", liveexec.ExecCommandOptions{})).To(Succeed())
		})
	})

	Context("when installing python requirements using ansible and python files contain utf-8 chars", func() {
		BeforeEach(func() {
			os.RemoveAll("python_encoding/.git")
			os.RemoveAll("python_encoding_repo")
		})

		AfterEach(func(ctx SpecContext) {
			werfHostPurge(ctx, "python_encoding", liveexec.ExecCommandOptions{}, "--force")
			os.RemoveAll("python_encoding/.git")
			os.RemoveAll("python_encoding_repo")
		})

		It("should install packages successfully without utf-8 related problems", func(ctx SpecContext) {
			if runtime.GOOS == "windows" {
				Skip("Skipping on windows")
			}

			Expect(utils.SetGitRepoState(ctx, "python_encoding", "python_encoding_repo", "initial commit")).To(Succeed())
			Expect(werfBuild(ctx, "python_encoding", liveexec.ExecCommandOptions{})).To(Succeed())
		})
	})

	Context("Non standard PATH used in the base image (https://github.com/werf/werf/issues/1836) ", func() {
		AfterEach(func(ctx SpecContext) {
			werfHostPurge(ctx, "path_redefined_in_stapel_1836", liveexec.ExecCommandOptions{}, "--force")
			os.RemoveAll("path_redefined_in_stapel_1836/.git")
			os.RemoveAll("path_redefined_in_stapel_1836_repo")
		})

		It("PATH should not be redefined in stapel build container", func(ctx SpecContext) {
			Expect(utils.SetGitRepoState(ctx, "path_redefined_in_stapel_1836", "path_redefined_in_stapel_1836_repo", "initial commit")).To(Succeed())
			Expect(werfBuild(ctx, "path_redefined_in_stapel_1836", liveexec.ExecCommandOptions{})).To(Succeed())
		})
	})
})
