package e2e_build_test

import (
	"context"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/buildah"
	"github.com/werf/werf/v3/test/pkg/contback"
	"github.com/werf/werf/v3/test/pkg/report"
	"github.com/werf/werf/v3/test/pkg/utils"
)

var _ = ginkgo.Describe("Test image cleanup", ginkgo.Label("e2e", "build", "extra"), func() {
	opts := setupEnvOptions{ContainerBackendMode: "native-rootless"}
	ginkgo.It("reports the busy image and its owner in the spec log before release", entryLabels(opts), func(ctx ginkgo.SpecContext) {
		setupEnv(opts)
		contback.SkipIfUnavailable(opts.ContainerBackendMode)
		commonArgs, err := buildah.GetBasicBuildahCliArgs(buildah.DefaultStorageDriver)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		cli := func(ctx context.Context, args ...string) string {
			return strings.TrimSpace(utils.SucceedCommandOutputString(ctx, "", "buildah", slices.Concat(commonArgs, args)...))
		}
		projectName := SuiteData.ProjectName
		ref := "localhost/" + projectName + ":held"
		builder := projectName + "-builder"
		cli(ctx, "from", "--name", builder, "scratch")
		ginkgo.DeferCleanup(func(ctx ginkgo.SpecContext) { cli(ctx, "rm", builder) }, ginkgo.NodeTimeout(time.Minute))
		cli(ctx, "config", "--label", "werf="+projectName, builder)
		cli(ctx, "commit", builder, ref)
		holder := projectName + "-holder"
		cli(ctx, "from", "--name", holder, ref)
		ginkgo.DeferCleanup(func(ctx ginkgo.SpecContext) {
			if holder != "" {
				cli(ctx, "rm", holder)
			}
		}, ginkgo.NodeTimeout(time.Minute))
		ownerID := cli(ctx, "inspect", "--type", "container", "--format", "{{.ContainerID}}", holder)
		gomega.Expect(ownerID).To(gomega.MatchRegexp(`^[a-f0-9]{64}$`))

		cleanupCtx, cancel := context.WithTimeout(ctx, time.Minute)
		var worker sync.WaitGroup
		var cleanupErr error
		worker.Go(func() {
			cleanupErr = contback.CleanupProject(cleanupCtx, projectName, contback.CleanupProjectOptions{BackendModes: []string{opts.ContainerBackendMode}})
		})
		defer func() {
			cancel()
			worker.Wait()
		}()
		gomega.Eventually(func() string {
			return ginkgo.CurrentSpecReport().CapturedGinkgoWriterOutput
		}).WithContext(cleanupCtx).WithTimeout(30 * time.Second).Should(gomega.MatchRegexp(`buildah cleanup: image "` + regexp.QuoteMeta(ref) + `"[^\n]*` + ownerID))
		cli(ctx, "rm", holder)
		holder = ""
		worker.Wait()
		gomega.Expect(cleanupErr).NotTo(gomega.HaveOccurred())
	})

	ginkgo.DescribeTable("removes project registry tags while retaining another project's alias", func(ctx ginkgo.SpecContext, opts setupEnvOptions) {
		setupEnv(opts)
		contback.SkipIfUnavailable(opts.ContainerBackendMode)
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
		cleanupOpts := contback.CleanupProjectOptions{BackendModes: []string{opts.ContainerBackendMode}}
		ginkgo.DeferCleanup(func(cleanupCtx ginkgo.SpecContext) {
			cleanupOpts.Repositories = []string{foreignRepo}
			gomega.Expect(contback.CleanupProject(cleanupCtx, SuiteData.ProjectName, cleanupOpts)).To(gomega.Succeed())
		}, ginkgo.NodeTimeout(time.Minute))
		utils.RunSucceedCommand(ctx, "", backend, append(slices.Clone(commonArgs), "tag", imageRef, registryRef)...)
		utils.RunSucceedCommand(ctx, "", backend, append(slices.Clone(commonArgs), "tag", imageRef, foreignRef)...)
		for _, ref := range []string{imageRef, registryRef, foreignRef} {
			utils.RunSucceedCommand(ctx, "", backend, append(slices.Concat(commonArgs, inspectArgs), ref)...)
		}

		cleanupCtx, cancelCleanup := context.WithTimeout(ctx, 2*time.Minute)
		defer cancelCleanup()
		gomega.Expect(contback.CleanupProject(cleanupCtx, SuiteData.ProjectName, cleanupOpts)).To(gomega.Succeed())
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
