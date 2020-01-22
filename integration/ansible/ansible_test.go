package ansible_test

import (
	"os"
	"runtime"
	"strings"

	"github.com/flant/werf/pkg/testing/utils/liveexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Stapel builder with ansible", func() {
	Context("when building image based on alpine, ubuntu or centos", func() {
		AfterEach(func() {
			werfPurge("general", liveexec.ExecCommandOptions{})
		})

		It("should successfully build image using arbitrary ansible modules", func(done Done) {
			Expect(werfBuild("general", liveexec.ExecCommandOptions{})).To(Succeed())
			close(done)
		}, 500)
	})

	Context("when building stapel image based on centos 6 and 7", func() {
		AfterEach(func() {
			werfPurge("yum1", liveexec.ExecCommandOptions{})
		})

		It("successfully installs packages using yum module", func(done Done) {
			Expect(werfBuild("yum1", liveexec.ExecCommandOptions{})).To(Succeed())
			close(done)
		}, 500)
	})

	Context("when building stapel image based on centos 8", func() {
		AfterEach(func() {
			werfPurge("yum2", liveexec.ExecCommandOptions{})
		})

		It("successfully installs packages using yum module", func(done Done) {
			Skip("FIXME https://github.com/flant/werf/issues/1983")
			Expect(werfBuild("yum2", liveexec.ExecCommandOptions{})).To(Succeed())
			close(done)
		}, 500)
	})

	Context("when become_user task option used", func() {
		AfterEach(func() {
			werfPurge("become_user", liveexec.ExecCommandOptions{})
		})

		It("successfully installs packages using yum module", func(done Done) {
			Skip("FIXME https://github.com/flant/werf/issues/1806")
			Expect(werfBuild("become_user", liveexec.ExecCommandOptions{})).To(Succeed())
			close(done)
		}, 500)
	})

	Context("when using apt_key module used (1)", func() {
		AfterEach(func() {
			werfPurge("apt_key1-001", liveexec.ExecCommandOptions{})
		})

		It("should fail to install package without a key and succeed with the key", func(done Done) {
			Skip("https://github.com/flant/werf/issues/2000")

			gotNoPubkey := false
			Expect(werfBuild("apt_key1-001", liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					if strings.Index(line, `public key is not available: NO_PUBKEY`) != -1 {
						gotNoPubkey = true
					}
				},
			})).NotTo(Succeed())
			Expect(gotNoPubkey).To(BeTrue())

			gotPackageInstallDone := false
			Expect(werfBuild("apt_key1-002", liveexec.ExecCommandOptions{
				OutputLineHandler: func(line string) {
					if strings.Index(line, `apt 'Install package from new repository' [clickhouse-client]`) != -1 {
						gotPackageInstallDone = true
					}
					Expect(line).NotTo(MatchRegexp(`apt 'Install package from new repository' \[clickhouse-client\] \(".*" seconds\) FAILED`))
				},
			})).To(Succeed())
			Expect(gotPackageInstallDone).To(BeTrue())

			close(done)
		}, 500)
	})

	Context("when using apt_key module used (2)", func() {
		AfterEach(func() {
			werfPurge("apt_key2", liveexec.ExecCommandOptions{})
		})

		It("should fail to install package without a key and succeed with the key", func(done Done) {
			Skip("https://github.com/flant/werf/issues/2000")

			Expect(werfBuild("apt_key2", liveexec.ExecCommandOptions{})).To(Succeed())
			close(done)
		}, 500)
	})

	Context("when apt-mark from apt module used (https://github.com/flant/werf/issues/1820)", func() {
		AfterEach(func() {
			werfPurge("apt_mark_panic_1820", liveexec.ExecCommandOptions{})
		})

		It("should not panic in all supported ubuntu versions", func(done Done) {
			Expect(werfBuild("apt_mark_panic_1820", liveexec.ExecCommandOptions{})).To(Succeed())
			close(done)
		}, 500)
	})

	Context("when using yarn module to install nodejs packages", func() {
		AfterEach(func() {
			werfPurge("yarn", liveexec.ExecCommandOptions{})
			os.RemoveAll("yarn/.git")
			os.RemoveAll("yarn_repo")
		})

		It("should install packages successfully", func(done Done) {
			Expect(setGitRepoState("yarn", "yarn_repo", "initial commit")).To(Succeed())
			Expect(werfBuild("yarn", liveexec.ExecCommandOptions{})).To(Succeed())
			close(done)
		}, 500)
	})

	Context("when installing python requirements using ansible and python files contain utf-8 chars", func() {
		AfterEach(func() {
			werfPurge("python_encoding", liveexec.ExecCommandOptions{})
			os.RemoveAll("python_encoding/.git")
			os.RemoveAll("python_encoding_repo")
		})

		It("should install packages successfully without utf-8 related problems", func(done Done) {
			if runtime.GOOS == "windows" {
				Skip("Skipping on windows")
			}

			Expect(setGitRepoState("python_encoding", "python_encoding_repo", "initial commit")).To(Succeed())
			Expect(werfBuild("python_encoding", liveexec.ExecCommandOptions{})).To(Succeed())
			close(done)
		}, 900)
	})

	Context("Non standard PATH used in the base image (https://github.com/flant/werf/issues/1836) ", func() {
		AfterEach(func() {
			werfPurge("path_redefined_in_stapel_1836", liveexec.ExecCommandOptions{})
		})

		It("PATH should not be redefined in stapel build container", func(done Done) {
			Expect(werfBuild("path_redefined_in_stapel_1836", liveexec.ExecCommandOptions{})).To(Succeed())
			close(done)
		}, 500)
	})
})
