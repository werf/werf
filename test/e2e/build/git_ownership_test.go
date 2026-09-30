package e2e_build_test

import (
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/test/pkg/contback"
	"github.com/werf/werf/v3/test/pkg/report"
	"github.com/werf/werf/v3/test/pkg/werf"
)

var _ = ginkgo.Describe("Git mapping ownership", ginkgo.Label("e2e", "build", "git-ownership"), func() {
	ginkgo.DescribeTable("should apply the declared owner to each image regardless of build order",
		func(ctx ginkgo.SpecContext, testOpts gitOwnershipTestOptions) {
			ginkgo.By("initializing")
			setupEnv(testOpts.setupEnvOptions)
			contback.SkipIfUnavailable(testOpts.ContainerBackendMode)
			contRuntime := contback.NewContainerBackend(testOpts.ContainerBackendMode)

			buildOrder := []string{gitOwnershipDefaultImage, gitOwnershipOwnedImage}
			if testOpts.OwnerFirst {
				buildOrder = []string{gitOwnershipOwnedImage, gitOwnershipDefaultImage}
			}

			repoDirname := "repo0"

			ginkgo.By("state0: preparing test repo")
			SuiteData.InitTestRepo(ctx, repoDirname, "git_ownership/state0")
			werfProject := newWerfProject(repoDirname)
			reportProject := report.NewProjectWithReport(werfProject)

			ginkgo.By("state0: building images one by one, so that the first one writes the shared git archive")
			for _, imageName := range buildOrder {
				buildOut := werfProject.Build(ctx, &werf.BuildOptions{
					CommonOptions: werf.CommonOptions{ExtraArgs: []string{imageName}},
				})
				gomega.Expect(buildOut).To(gomega.ContainSubstring(fmt.Sprintf("Building stage %s/gitArchive", imageName)))
			}

			ginkgo.By("state0: rebuilding both images reuses them by content-based tag")
			buildOut, buildReport := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath("report0.json"), nil)
			gomega.Expect(buildOut).To(gomega.ContainSubstring("Use previously built image"))
			gomega.Expect(buildOut).NotTo(gomega.ContainSubstring("Building stage"))

			defaultImage := buildReport.Images[gitOwnershipDefaultImage].DockerImageName
			ownedImage := buildReport.Images[gitOwnershipOwnedImage].DockerImageName
			if testOpts.WithLocalRepo {
				contRuntime.Pull(ctx, defaultImage)
				contRuntime.Pull(ctx, ownedImage)
			}

			ginkgo.By("state0: checking ownership of the image without owner/group")
			contRuntime.ExpectCmdsToSucceed(ctx, defaultImage, gitOwnershipChecks("0:0", "state0")...)

			ginkgo.By("state0: checking ownership of the image with owner/group")
			contRuntime.ExpectCmdsToSucceed(ctx, ownedImage, gitOwnershipChecks("1001:1002", "state0")...)

			ginkgo.By("state1: changing git files in test repo")
			SuiteData.UpdateTestRepo(ctx, repoDirname, "git_ownership/state1")

			ginkgo.By("state1: patching images one by one, without rebuilding the install stage")
			for _, imageName := range buildOrder {
				buildOut := werfProject.Build(ctx, &werf.BuildOptions{
					CommonOptions: werf.CommonOptions{ExtraArgs: []string{imageName}},
				})
				gomega.Expect(buildOut).To(gomega.ContainSubstring(fmt.Sprintf("Building stage %s/gitLatestPatch", imageName)))
				gomega.Expect(buildOut).NotTo(gomega.ContainSubstring(fmt.Sprintf("Building stage %s/install", imageName)))
			}

			ginkgo.By("state1: rebuilding both images reuses them by content-based tag")
			buildOut, buildReport = reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath("report1.json"), nil)
			gomega.Expect(buildOut).To(gomega.ContainSubstring("Use previously built image"))
			gomega.Expect(buildOut).NotTo(gomega.ContainSubstring("Building stage"))

			defaultImage = buildReport.Images[gitOwnershipDefaultImage].DockerImageName
			ownedImage = buildReport.Images[gitOwnershipOwnedImage].DockerImageName
			if testOpts.WithLocalRepo {
				contRuntime.Pull(ctx, defaultImage)
				contRuntime.Pull(ctx, ownedImage)
			}

			patchedPaths := []string{"/app/new-file", "/app/nested/new-file", "/app/new-link"}

			ginkgo.By("state1: checking ownership of the patched image without owner/group")
			contRuntime.ExpectCmdsToSucceed(ctx, defaultImage, gitOwnershipChecks("0:0", "state1", patchedPaths...)...)

			ginkgo.By("state1: checking ownership of the patched image with owner/group")
			contRuntime.ExpectCmdsToSucceed(ctx, ownedImage, gitOwnershipChecks("1001:1002", "state1", patchedPaths...)...)
		},
		backendEntry("Docker, image without owner built first", gitOwnershipTestOptions{
			setupEnvOptions: setupEnvOptions{ContainerBackendMode: "docker"},
		}),
		backendEntry("Docker, image with owner built first", gitOwnershipTestOptions{
			setupEnvOptions: setupEnvOptions{ContainerBackendMode: "docker"},
			OwnerFirst:      true,
		}),
		backendEntry("Native Buildah rootless, image without owner built first", gitOwnershipTestOptions{
			setupEnvOptions: setupEnvOptions{ContainerBackendMode: "native-rootless", WithLocalRepo: true},
		}),
		backendEntry("Native Buildah rootless, image with owner built first", gitOwnershipTestOptions{
			setupEnvOptions: setupEnvOptions{ContainerBackendMode: "native-rootless", WithLocalRepo: true},
			OwnerFirst:      true,
		}),
	)
})
