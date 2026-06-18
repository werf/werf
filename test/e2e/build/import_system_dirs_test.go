package e2e_build_test

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"strings"

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

			By("checking system directories are not present in destination image layers")
			imageName := buildReport.Images["destination"].DockerImageName
			checkNoSystemDirsInImageLayers(ctx, imageName)
		},
		Entry("Vanilla Docker", setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}),
	)
})

// checkNoSystemDirsInImageLayers inspects the raw image layers via docker save.
// Unlike docker run, this avoids the Docker runtime mounting /proc, /sys, /dev
// into the container — we check only what is actually stored in the image layers.
func checkNoSystemDirsInImageLayers(ctx context.Context, imageName string) {
	saveOut, err := utils.RunCommand(ctx, "/", "docker", "save", imageName)
	Expect(err).NotTo(HaveOccurred(), "docker save failed")

	systemDirs := []string{"proc/", "sys/", "dev/", "run/"}
	foundDirs := map[string]bool{}

	outerTar := tar.NewReader(bytes.NewReader(saveOut))
	for {
		hdr, err := outerTar.Next()
		if err == io.EOF {
			break
		}
		Expect(err).NotTo(HaveOccurred())

		// Each layer is a tar inside the outer tar named like "<hash>/layer.tar"
		if !strings.HasSuffix(hdr.Name, "/layer.tar") {
			continue
		}

		layerData, err := io.ReadAll(outerTar)
		Expect(err).NotTo(HaveOccurred())

		innerTar := tar.NewReader(bytes.NewReader(layerData))
		for {
			innerHdr, err := innerTar.Next()
			if err == io.EOF {
				break
			}
			Expect(err).NotTo(HaveOccurred())

			for _, dir := range systemDirs {
				if innerHdr.Name == dir || innerHdr.Name == "./"+dir {
					foundDirs[strings.TrimSuffix(dir, "/")] = true
				}
			}
		}
	}

	Expect(foundDirs).To(BeEmpty(),
		"system directories must not exist in image layers: found %v", foundDirs)
}
