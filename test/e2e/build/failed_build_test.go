package e2e_build_test

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v3/test/pkg/contback"
	"github.com/werf/werf/v3/test/pkg/utils"
	"github.com/werf/werf/v3/test/pkg/werf"
)

var _ = Describe("Failed Dockerfile build", Label("e2e", "build", "failed-build"), func() {
	DescribeTable("leaves no working container in the Buildah storage",
		func(ctx SpecContext, opts setupEnvOptions) {
			By("initializing")
			contback.SkipIfUnavailable(opts.ContainerBackendMode)
			setupEnv(opts)
			// The background cleaner removes werf containers and would hide a leaked one.
			SuiteData.Stubs.SetEnv("WERF_DISABLE_AUTO_HOST_CLEANUP", "1")

			By("preparing test repo")
			const repoDirname = "repo0"
			SuiteData.InitTestRepo(ctx, repoDirname, "failed_build")
			repoPath := SuiteData.GetTestRepoPath(repoDirname)
			dockerfilePath := filepath.Join(repoPath, "Dockerfile")
			dockerfile, err := os.ReadFile(dockerfilePath)
			Expect(err).NotTo(HaveOccurred())
			dockerfile = []byte(strings.ReplaceAll(string(dockerfile), "PROJECT_NAME", SuiteData.ProjectName))
			Expect(os.WriteFile(dockerfilePath, dockerfile, 0o644)).To(Succeed())
			utils.RunSucceedCommand(ctx, repoPath, "git", "commit", "-am", "project marker")

			By("building images")
			buildOut := newWerfProject(repoDirname).Build(ctx, &werf.BuildOptions{
				CommonOptions: werf.CommonOptions{ShouldFail: true},
			})
			Expect(buildOut).To(ContainSubstring(`building at STEP "RUN false"`))

			By("checking the Buildah storage")
			snapshot, err := contback.SnapshotBuildahStorage(ctx, SuiteData.ProjectName)
			Expect(err).NotTo(HaveOccurred())
			Expect(snapshot.ProjectImageIDs).NotTo(BeEmpty(), "the layer committed before the failing step is expected in the storage")
			for _, container := range snapshot.Containers {
				Expect(snapshot.ProjectImageIDs).NotTo(ContainElement(container.ImageID), "working container left by the failed build: %s %v", container.ID, container.Names)
			}
		},
		backendEntry("with Native Buildah with rootless isolation", setupEnvOptions{ContainerBackendMode: "native-rootless"}),
		backendEntry("with Native Buildah with chroot isolation", setupEnvOptions{ContainerBackendMode: "native-chroot"}),
	)
})
