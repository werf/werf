package e2e_build_test

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v3/test/pkg/contback"
	"github.com/werf/werf/v3/test/pkg/suite_init"
)

var _ = Describe("Staged Dockerfile build with RUN --mount from stage", Label("e2e", "build", "staged_dockerfile_run_mount"), func() {
	It("removes only tags from the requested repository", func(ctx SpecContext) {
		tmpDir := GinkgoT().TempDir()
		callsPath := filepath.Join(tmpDir, "calls")
		buildahPath := filepath.Join(tmpDir, "buildah")
		Expect(os.WriteFile(buildahPath, []byte("#!/bin/sh\nif [ \"$1\" = images ]; then\n  printf 'registry/project:one\\nregistry/other:keep\\n'\nelse\n  printf '%s\\n' \"$3\" >> \"$CALLS_PATH\"\nfi\n"), 0o755)).To(Succeed())

		oldPath := os.Getenv("PATH")
		oldCallsPath, hadCallsPath := os.LookupEnv("CALLS_PATH")
		Expect(os.Setenv("PATH", tmpDir+string(os.PathListSeparator)+oldPath)).To(Succeed())
		Expect(os.Setenv("CALLS_PATH", callsPath)).To(Succeed())
		DeferCleanup(func() {
			Expect(os.Setenv("PATH", oldPath)).To(Succeed())
			if hadCallsPath {
				Expect(os.Setenv("CALLS_PATH", oldCallsPath)).To(Succeed())
			} else {
				Expect(os.Unsetenv("CALLS_PATH")).To(Succeed())
			}
		})

		backend := &contback.NativeBuildahBackend{}
		backend.RmiByRepoRef(ctx, "registry/project")

		calls, err := os.ReadFile(callsPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(strings.Fields(string(calls))).To(Equal([]string{"registry/project:one"}))
	})

	DescribeTable("should pull the mount source stage missing in the local containers storage",
		func(ctx SpecContext, testOpts setupEnvOptions) {
			By("initializing")
			setupEnv(testOpts)
			contRuntime := contback.NewContainerBackend(testOpts.ContainerBackendMode)
			buildahRuntime, ok := contRuntime.(*contback.NativeBuildahBackend)
			Expect(ok).To(BeTrue(), "test requires the native buildah backend")

			repoDirname := "repo0"

			By("state0: preparing test repo")
			SuiteData.InitTestRepo(ctx, repoDirname, "staged_dockerfile_run_mount/state0")
			werfProject := newWerfProject(repoDirname)

			By("state0: building images")
			Expect(werfProject.Build(ctx, nil)).To(ContainSubstring("Building stage"))

			By("state0: removing project images from the local containers storage")
			buildahRuntime.RmiByRepoRef(ctx, suite_init.TestRepo(SuiteData.ProjectName))

			By("state1: changing the final stage only")
			SuiteData.UpdateTestRepo(ctx, repoDirname, "staged_dockerfile_run_mount/state1")

			By("state1: rebuilding with the mount source stage present only in repo")
			buildOut := werfProject.Build(ctx, nil)
			Expect(buildOut).To(ContainSubstring("Use previously built image"))
			Expect(buildOut).To(ContainSubstring("Building stage"))
		},
		backendEntry("with local repo using Native Buildah with rootless isolation", setupEnvOptions{
			ContainerBackendMode: "native-rootless",
			WithLocalRepo:        true,
		}),
		backendEntry("with local repo using Native Buildah with chroot isolation", setupEnvOptions{
			ContainerBackendMode: "native-chroot",
			WithLocalRepo:        true,
		}),
	)
})
