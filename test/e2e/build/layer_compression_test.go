package e2e_build_test

import (
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v3/test/pkg/contback"
	"github.com/werf/werf/v3/test/pkg/report"
)

var _ = Describe("Layer compression", Label("e2e", "build", "compression", "simple"), func() {
	DescribeTable("should publish images with the layer compression from werf.yaml",
		func(ctx SpecContext, testOpts setupEnvOptions) {
			By("initializing")
			setupEnv(testOpts)
			contback.SkipIfUnavailable(testOpts.ContainerBackendMode)

			By("preparing test repo")
			SuiteData.InitTestRepo(ctx, "repo0", "layer_compression/state0")

			By("building images")
			reportProject := report.NewProjectWithReport(newWerfProject("repo0"))
			_, buildReport := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath("report0.json"), nil)

			for imageName, expectedLayerMediaType := range map[string]types.MediaType{
				"none":        types.OCIUncompressedLayer,
				"zstd":        types.OCILayerZStd,
				"stapel-none": types.OCIUncompressedLayer,
				"stapel-zstd": types.OCILayerZStd,
			} {
				By("checking the published manifest of image " + imageName)
				ref, err := name.ParseReference(buildReport.Images[imageName].DockerImageName, name.Insecure)
				Expect(err).NotTo(HaveOccurred())

				img, err := remote.Image(ref, remote.WithContext(ctx))
				Expect(err).NotTo(HaveOccurred())

				manifest, err := img.Manifest()
				Expect(err).NotTo(HaveOccurred())
				Expect(manifest.MediaType).To(Equal(types.OCIManifestSchema1), "image %s", imageName)
				Expect(manifest.Layers).NotTo(BeEmpty(), "image %s", imageName)
				for _, layer := range manifest.Layers {
					Expect(layer.MediaType).To(Equal(expectedLayerMediaType), "image %s layer %s", imageName, layer.Digest)
				}
			}
		},
		backendEntry("using Native Buildah with rootless isolation", setupEnvOptions{
			ContainerBackendMode: "native-rootless",
			WithLocalRepo:        true,
		}),
	)
})
