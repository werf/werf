package config

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"gopkg.in/yaml.v2"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v3/pkg/image"
)

var _ = Describe("layer compression in werf.yaml", func() {
	prepare := func(metaYaml string, stapelYaml, dockerfileYaml string) *WerfConfig {
		parentStack = util.NewStack()
		giterminismManager := NewGiterminismManagerStub(NewLocalGitRepoStub("9d8059842b6fde712c58315ca0ab4713d90761c0"))

		meta := &rawMeta{doc: &doc{Content: []byte(metaYaml)}}
		Expect(yaml.UnmarshalStrict(meta.doc.Content, meta)).To(Succeed())

		stapelImage := &rawStapelImage{doc: &doc{Content: []byte(stapelYaml)}}
		Expect(yaml.UnmarshalStrict(stapelImage.doc.Content, stapelImage)).To(Succeed())

		dockerfileImage := &rawImageFromDockerfile{doc: &doc{Content: []byte(dockerfileYaml)}}
		Expect(yaml.UnmarshalStrict(dockerfileImage.doc.Content, dockerfileImage)).To(Succeed())

		werfConfig, err := prepareWerfConfig(giterminismManager, []*rawStapelImage{stapelImage}, []*rawImageFromDockerfile{dockerfileImage}, meta.toMeta())
		Expect(err).To(Succeed())

		return werfConfig
	}

	It("applies the build compression to images without their own and lets an image override it", func() {
		werfConfig := prepare(`
configVersion: 1
project: test
build:
  compression:
    format: gzip
    level: 9
`, `
image: stapel
from: alpine:3.20
compression:
  format: zstd
`, `
image: dockerfile
dockerfile: Dockerfile
`)

		Expect(werfConfig.GetImage("stapel").Compression()).To(Equal(image.LayerCompression{Format: image.LayerCompressionZstd}),
			"the image setting replaces the build setting as a whole, so the gzip level does not leak into zstd")
		Expect(werfConfig.GetImage("dockerfile").Compression()).To(Equal(image.LayerCompression{Format: image.LayerCompressionGzip, Level: lo.ToPtr(9)}))
	})

	It("applies the build compression to stapel images without their own", func() {
		werfConfig := prepare(`
configVersion: 1
project: test
build:
  compression:
    format: zstd
`, `
image: stapel
from: alpine:3.20
`, `
image: dockerfile
dockerfile: Dockerfile
compression:
  format: none
`)

		Expect(werfConfig.GetImage("stapel").Compression()).To(Equal(image.LayerCompression{Format: image.LayerCompressionZstd}))
		Expect(werfConfig.GetImage("dockerfile").Compression()).To(Equal(image.LayerCompression{Format: image.LayerCompressionNone}))
	})

	It("leaves images without compression settings to the command line default", func() {
		werfConfig := prepare(`
configVersion: 1
project: test
`, `
image: stapel
from: alpine:3.20
`, `
image: dockerfile
dockerfile: Dockerfile
compression:
  format: none
`)

		Expect(werfConfig.GetImage("stapel").Compression().IsSet()).To(BeFalse())
		Expect(werfConfig.GetImage("dockerfile").Compression()).To(Equal(image.LayerCompression{Format: image.LayerCompressionNone}))
	})
})
