package tar_test

import (
	"archive/tar"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/v1/tarball"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	elfTar "github.com/werf/werf/v2/pkg/signature/elf/tar"
)

var _ = DescribeTable("Reader should work",
	func(tarPath string, useImageTarLoader bool, expectedFiles []fileExpectation) {
		var rcList []io.ReadCloser

		if useImageTarLoader {
			rcList = newReadCloserListFromImageTarPath(tarPath)
		} else {
			rcList = newReadCloserListFromTarPath(tarPath)
		}

		var actualFiles []fileExpectation

		for _, rc := range rcList {
			elfTarReader := elfTar.NewReader(tar.NewReader(rc))

			for {
				header, err := elfTarReader.Next()
				if errors.Is(err, io.EOF) {
					break
				} else if err != nil {
					Expect(err).To(Succeed())
				}

				actualFiles = append(actualFiles, fileExpectation{
					Name:  header.Name,
					Size:  header.Size,
					IsELF: header.IsELF,
				})
			}

			Expect(rc.Close()).To(Succeed())
		}

		Expect(actualFiles).To(Equal(expectedFiles))
	},
	Entry(
		"with unsigned curl from docker image",
		"testdata/docker-image-with-curl.tar",
		true,
		[]fileExpectation{
			{Name: "usr/", Size: 0, IsELF: false},
			{Name: "usr/bin/", Size: 0, IsELF: false},
			{Name: "usr/bin/curl", Size: 239848, IsELF: true},
		},
	),
	Entry(
		"with unsigned curl and text file from docker image",
		"testdata/docker-image-with-curl-and-text.tar",
		true,
		[]fileExpectation{
			{Name: "opt/", Size: 0, IsELF: false},
			{Name: "opt/some.txt", Size: 15, IsELF: false},
			{Name: "usr/", Size: 0, IsELF: false},
			{Name: "usr/.wh..wh..opq", Size: 0, IsELF: false},
			{Name: "usr/bin/", Size: 0, IsELF: false},
			{Name: "usr/bin/curl", Size: 239848, IsELF: true},
		},
	),
	Entry(
		"with unsigned curl from docker image",
		"testdata/plain-hello.tar",
		false,
		[]fileExpectation{
			{Name: "hello.elf", Size: 15960, IsELF: true},
			{Name: "hello.txt", Size: 6, IsELF: false},
		},
	),
)

func newReadCloserListFromImageTarPath(path string) []io.ReadCloser {
	img, err := tarball.ImageFromPath(fullPath(path), nil)
	Expect(err).To(Succeed())

	layers, err := img.Layers()
	Expect(err).To(Succeed())

	uncompressedLayers := make([]io.ReadCloser, len(layers))

	for i, layer := range layers {
		uncompressed, err := layer.Uncompressed()
		Expect(err).To(Succeed())
		uncompressedLayers[i] = uncompressed
	}

	return uncompressedLayers
}

func newReadCloserListFromTarPath(path string) []io.ReadCloser {
	file, err := os.Open(fullPath(path))
	Expect(err).To(Succeed())
	return []io.ReadCloser{file}
}

func fullPath(path string) string {
	dir, err := os.Getwd()
	Expect(err).To(Succeed())
	return filepath.Join(dir, path)
}

type fileExpectation struct {
	Name  string
	Size  int64
	IsELF bool
}
