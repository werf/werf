package stream_reader

import (
	"archive/tar"
	"bytes"
	_ "embed"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Generation command: "docker save -o /tmp/docker-native-uncompressed-layers.tar <image>"
//
//go:embed testdata/docker-native-uncompressed-layers.tar
var dockerNativeUncompressedLayersTarball []byte

// Generation command: "buildah push -D <image> oci-archive:/tmp/buildah-oci-uncompressed-layers.tar"
//
//go:embed testdata/buildah-oci-uncompressed-layers.tar
var buildahOciUncompressedLayersTarball []byte

// Generation command: "buildah push -D <image> docker-archive:/tmp/buildah-docker-uncompressed-layers.tar"
//
//go:embed testdata/buildah-docker-uncompressed-layers.tar
var buildahDockerUncompressedLayersTarball []byte

// Generation command: "buildah push <image> oci-archive:/tmp/buildah-oci-compressed-layers.tar"
//
//go:embed testdata/buildah-oci-compressed-layers.tar
var buildahOciCompressedLayersTarball []byte

var _ = Describe("file system stream reader", func() {
	DescribeTable("read tarball via reader.Nex()",
		func(tarball []byte) {
			tarReader := tar.NewReader(bytes.NewBuffer(tarball))

			fsStreamReader, err := NewFileSystemStreamReader(tarReader)
			Expect(err).To(Succeed())

			f1, err := fsStreamReader.Next()
			Expect(err).To(Succeed())
			Expect(f1.Path()).To(Equal("sbom/"))

			f2, err := fsStreamReader.Next()
			Expect(err).To(Succeed())
			Expect(f2.Path()).To(Equal("sbom/cyclonedx@1.6/"))

			f3, err := fsStreamReader.Next()
			Expect(err).To(Succeed())
			Expect(f3.Path()).To(Equal("sbom/cyclonedx@1.6/70ee6b0600f471718988bc123475a625ecd4a5763059c62802ae6280e65f5623.json"))

			f4, err := fsStreamReader.Next()
			Expect(err).To(Succeed())
			Expect(f4).To(BeNil())
		},
		Entry(
			"should work for Docker native which produces uncompressed layers",
			dockerNativeUncompressedLayersTarball,
		),
		Entry(
			"should work for Buildah oci-archive transport which produces uncompressed layers",
			buildahOciUncompressedLayersTarball,
		),
		Entry(
			"should work for Buildah docker-archive transport which produces uncompressed layers",
			buildahDockerUncompressedLayersTarball,
		),
		Entry(
			"should work for Buildah docker-archive transport which produces compressed layers",
			buildahOciCompressedLayersTarball,
		),
	)

	DescribeTable("find tarball file via reader.Find()",
		func(tarball []byte, expectedLen int) {
			tarReader := tar.NewReader(bytes.NewBuffer(tarball))

			fsStreamReader, err := NewFileSystemStreamReader(tarReader)
			Expect(err).To(Succeed())

			found, ok, err := fsStreamReader.Find(func(file *File) bool {
				return file.Path() == "sbom/cyclonedx@1.6/70ee6b0600f471718988bc123475a625ecd4a5763059c62802ae6280e65f5623.json"
			})
			Expect(err).To(Succeed())
			Expect(ok).To(BeTrue())

			b, err := io.ReadAll(found)
			Expect(err).To(Succeed())
			Expect(b).To(HaveLen(expectedLen))
		},
		Entry(
			"should work for Docker native which produces uncompressed layers",
			dockerNativeUncompressedLayersTarball,
			53704,
		),
		Entry(
			"should work for Buildah oci-archive transport which produces uncompressed layers",
			buildahOciUncompressedLayersTarball,
			57364,
		),
		Entry(
			"should work for Buildah docker-archive transport which produces uncompressed layers",
			buildahDockerUncompressedLayersTarball,
			57364,
		),
		Entry(
			"should work for Buildah docker-archive transport which produces compressed layers",
			buildahOciCompressedLayersTarball,
			52082,
		),
	)
})
