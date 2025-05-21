package stream_reader

import (
	"archive/tar"
	"bytes"
	_ "embed"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

//go:embed testdata/docker-img.tar
var dockerImgTarball []byte

//go:embed testdata/buildah-img.tar
var buildahImgTarball []byte

var _ = Describe("file system stream reader", func() {
	DescribeTable("read tarball via reader.Nex()",
		func(tarball []byte) {
			tarReader := tar.NewReader(bytes.NewBuffer(tarball))

			fsReader, err := NewFileSystemStreamReader(tarReader)
			Expect(err).To(Succeed())

			f1, err := fsReader.Next()
			Expect(err).To(Succeed())
			Expect(f1.Path()).To(Equal("sbom/"))

			f2, err := fsReader.Next()
			Expect(err).To(Succeed())
			Expect(f2.Path()).To(Equal("sbom/cyclonedx@1.6/"))

			f3, err := fsReader.Next()
			Expect(err).To(Succeed())
			Expect(f3.Path()).To(Equal("sbom/cyclonedx@1.6/70ee6b0600f471718988bc123475a625ecd4a5763059c62802ae6280e65f5623.json"))

			f4, err := fsReader.Next()
			Expect(err).To(Succeed())
			Expect(f4).To(BeNil())
		},
		Entry(
			"should work for Docker",
			dockerImgTarball,
		),
		Entry(
			"should work for Buildah",
			buildahImgTarball,
		),
	)
})
