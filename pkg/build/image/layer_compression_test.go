package image

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/config"
	"github.com/werf/werf/v3/pkg/image"
)

var _ = Describe("Image.LayerCompression", func() {
	It("falls back to the command line compression for images without compression settings in werf.yaml", func() {
		defaultCompression := image.LayerCompression{Format: image.LayerCompressionZstd}

		dockerfileImage := &Image{
			CommonImageOptions:    CommonImageOptions{DefaultLayerCompression: defaultCompression},
			DockerfileImageConfig: &config.ImageFromDockerfile{Name: "app"},
		}
		Expect(dockerfileImage.LayerCompression()).To(Equal(defaultCompression))

		stapelImage := &Image{
			CommonImageOptions: CommonImageOptions{DefaultLayerCompression: defaultCompression},
			StapelImageConfig:  &config.StapelImage{StapelImageBase: &config.StapelImageBase{Name: "app"}},
		}
		Expect(stapelImage.LayerCompression()).To(Equal(defaultCompression))
	})
})
