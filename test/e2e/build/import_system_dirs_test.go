package e2e_build_test

import (
	"archive/tar"
	"bytes"
	"context"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/report"
	"github.com/werf/werf/v2/test/pkg/utils"
	"github.com/werf/werf/v2/test/pkg/werf"
)

var _ = Describe("Import system dirs", Label("e2e", "build", "import", "system-dirs"), func() {
	DescribeTable("should not import system directories when using wildcard includePaths",
		func(ctx SpecContext, testOpts setupEnvOptions) {
			By("initializing")
			setupEnv(testOpts)

			By("building")
			repoDirname := "repo0"
			fixtureRelPath := "import/no_system_dirs/state0"
			buildReportName := "report0.json"

			SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
			reportProject := report.NewProjectWithReport(werfProject)
			_, buildReport := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath(buildReportName), nil)

			By("checking system directories are not present in destination image")
			imageName := buildReport.Images["destination"].DockerImageName
			checkNoSystemDirsInImage(ctx, imageName)
		},
		Entry("Vanilla Docker", setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}),
	)
})

func checkNoSystemDirsInImage(ctx context.Context, imageName string) {
	containerName := "werf-test-system-dirs-check-" + utils.GetRandomString(8)

	createOut, err := utils.RunCommand(ctx, "/", "docker", "create", "--name", containerName, "--entrypoint", "", imageName)
	Expect(err).NotTo(HaveOccurred(), "docker create failed: %s", string(createOut))

	defer func() {
		utils.RunCommand(ctx, "/", "docker", "rm", "-f", containerName) //nolint:errcheck
	}()

	exportOut, err := utils.RunCommand(ctx, "/", "docker", "export", containerName)
	Expect(err).NotTo(HaveOccurred(), "docker export failed")

	systemDirs := []string{"proc", "sys", "dev", "run"}
	foundDirs := map[string]bool{}

	tr := tar.NewReader(bytes.NewReader(exportOut))
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		Expect(err).NotTo(HaveOccurred())

		for _, dir := range systemDirs {
			if hdr.Name == dir+"/" || hdr.Name == dir {
				foundDirs[dir] = true
			}
		}
	}

	Expect(foundDirs).To(BeEmpty(), "system directories must not be imported: found %v", foundDirs)
}
