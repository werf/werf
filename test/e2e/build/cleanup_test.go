package e2e_build_test

import (
	"context"
	"slices"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/buildah"
	"github.com/werf/werf/v3/test/pkg/contback"
	"github.com/werf/werf/v3/test/pkg/report"
	"github.com/werf/werf/v3/test/pkg/utils"
)

var _ = ginkgo.Describe("Test image cleanup", ginkgo.Label("e2e", "build", "extra"), func() {
	ginkgo.DescribeTable("removes project registry tags while retaining another project's alias", func(ctx ginkgo.SpecContext, opts setupEnvOptions) {
		contback.SkipIfUnavailable(opts.ContainerBackendMode)
		setupEnv(opts)
		backend := "docker"
		var commonArgs []string
		inspectArgs := []string{"image", "inspect"}
		if opts.ContainerBackendMode != "docker" {
			backend = "buildah"
			var err error
			commonArgs, err = buildah.GetBasicBuildahCliArgs(buildah.DefaultStorageDriver)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			inspectArgs = []string{"inspect", "--type", "image"}
		}
		SuiteData.InitTestRepo(ctx, "repo0", "simple/state0")
		project := report.NewProjectWithReport(newWerfProject("repo0"))
		_, buildReport := project.BuildWithReport(ctx, SuiteData.GetBuildReportPath("cleanup.json"), nil)
		imageRef := buildReport.Images["dockerfile"].DockerImageName
		registryRef := "localhost:65534/" + SuiteData.ProjectName + ":local-copy"
		foreignRepo := "localhost:65534/werf-test-neighbor-" + utils.GetRandomString(10)
		foreignRef := foreignRepo + ":keep"
		ginkgo.DeferCleanup(func(cleanupCtx ginkgo.SpecContext) {
			gomega.Expect(contback.CleanupProject(cleanupCtx, SuiteData.ProjectName, contback.CleanupProjectOptions{Repositories: []string{foreignRepo}})).To(gomega.Succeed())
		}, ginkgo.NodeTimeout(time.Minute))
		utils.RunSucceedCommand(ctx, "", backend, append(slices.Clone(commonArgs), "tag", imageRef, registryRef)...)
		utils.RunSucceedCommand(ctx, "", backend, append(slices.Clone(commonArgs), "tag", imageRef, foreignRef)...)
		for _, ref := range []string{imageRef, registryRef, foreignRef} {
			utils.RunSucceedCommand(ctx, "", backend, append(slices.Concat(commonArgs, inspectArgs), ref)...)
		}

		cleanupCtx, cancelCleanup := context.WithTimeout(ctx, 2*time.Minute)
		defer cancelCleanup()
		gomega.Expect(contback.CleanupProject(cleanupCtx, SuiteData.ProjectName, contback.CleanupProjectOptions{})).To(gomega.Succeed())
		for _, ref := range []string{imageRef, registryRef} {
			_, err := utils.RunCommand(ctx, "", backend, append(slices.Concat(commonArgs, inspectArgs), ref)...)
			gomega.Expect(err).To(gomega.HaveOccurred())
		}
		utils.RunSucceedCommand(ctx, "", backend, append(slices.Concat(commonArgs, inspectArgs), foreignRef)...)
	},
		backendEntry("using Docker", setupEnvOptions{ContainerBackendMode: "docker"}),
		backendEntry("using native rootless Buildah", setupEnvOptions{ContainerBackendMode: "native-rootless"}),
	)
})
