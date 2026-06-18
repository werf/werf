package e2e_build_test

import (
	"archive/tar"
	"io"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/report"
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

			By("checking image filesystem contents via gocontainerregistry")
			imageName := buildReport.Images["destination"].DockerImageName
			checkImageFilesystem(imageName)
		},
		Entry("Vanilla Docker", setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}),
	)
})

func checkImageFilesystem(imageName string) {
	ref, err := name.ParseReference(imageName)
	Expect(err).NotTo(HaveOccurred())

	img, err := daemon.Image(ref)
	Expect(err).NotTo(HaveOccurred())

	rc := mutate.Extract(img)
	defer rc.Close()

	var (
		foundMarker bool
		foundDirs   []string
		systemDirs  = map[string]bool{"proc/": true, "sys/": true, "dev/": true, "run/": true}
	)

	tr := tar.NewReader(rc)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		Expect(err).NotTo(HaveOccurred())

		n := hdr.Name
		if n == "etc/werf-test-marker" || n == "./etc/werf-test-marker" {
			foundMarker = true
		}
		for dir := range systemDirs {
			if n == dir || n == "./"+dir {
				foundDirs = append(foundDirs, dir)
			}
		}
	}

	Expect(foundMarker).To(BeTrue(), "expected /etc/werf-test-marker to exist in image")
	Expect(foundDirs).To(BeEmpty(), "system dirs must not exist in image layers: %v", foundDirs)
}
