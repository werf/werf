package e2e_build_test

import (
	"archive/tar"
	"io"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/contback"
	"github.com/werf/werf/v2/test/pkg/report"
	"github.com/werf/werf/v2/test/pkg/werf"
)

var _ = Describe("Import system dirs", Label("e2e", "build", "import", "system-dirs"), func() {
	DescribeTable("should not import system directories when using wildcard includePaths",
		func(ctx SpecContext, testOpts setupEnvOptions) {
			By("initializing")
			setupEnv(testOpts)
			_, err := contback.NewContainerBackend(testOpts.ContainerBackendMode)
			if err == contback.ErrRuntimeUnavailable {
				Skip(err.Error())
			} else if err != nil {
				Fail(err.Error())
			}

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
		Entry("BuildKit Docker", setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}),
		Entry("Native Buildah rootless", setupEnvOptions{
			ContainerBackendMode:        "native-rootless",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}),
		Entry("Native Buildah chroot", setupEnvOptions{
			ContainerBackendMode:        "native-chroot",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: true,
		}),
	)
})

func checkImageFilesystem(imageName string) {
	ref, err := name.ParseReference(imageName, name.Insecure)
	Expect(err).NotTo(HaveOccurred())

	img, err := remote.Image(ref)
	Expect(err).NotTo(HaveOccurred())

	rc := mutate.Extract(img)
	defer rc.Close()

	var (
		foundMarker       bool
		foundNestedMarker bool
		foundDirs         []string
		systemDirs        = map[string]bool{"proc/": true, "sys/": true, "dev/": true, "run/": true}
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
		if n == "var/run/werf-test-nested-marker" || n == "./var/run/werf-test-nested-marker" {
			foundNestedMarker = true
		}
		for dir := range systemDirs {
			if n == dir || n == "./"+dir {
				foundDirs = append(foundDirs, dir)
			}
		}
	}

	Expect(foundMarker).To(BeTrue(), "expected /etc/werf-test-marker to exist in image")
	Expect(foundNestedMarker).To(BeTrue(), "expected nested /var/run/werf-test-nested-marker to survive import (anchored excludes must not drop nested same-name dirs)")
	Expect(foundDirs).To(BeEmpty(), "system dirs must not exist in image layers: %v", foundDirs)
}
