package utils

import (
	"archive/tar"
	"context"
	"io"
	"os"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	. "github.com/onsi/gomega"
)

func ExpectFileContentInImage(ctx context.Context, backendMode, imageName, filePath, expectedContent string) {
	img, cleanup := loadLocalImage(ctx, backendMode, imageName)
	defer cleanup()
	rc := mutate.Extract(img)
	defer rc.Close()

	var fileContent string

	tr := tar.NewReader(rc)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		Expect(err).NotTo(HaveOccurred())

		if hdr.Name != filePath && hdr.Name != "./"+filePath {
			continue
		}

		data, err := io.ReadAll(tr)
		Expect(err).NotTo(HaveOccurred())
		fileContent = string(data)
	}

	Expect(fileContent).To(Equal(expectedContent), "expected image file content to match for %s", filePath)
}

// ExpectImageIsReadable asserts that the image can be saved and fully parsed by
// go-containerregistry: its manifest, config and every layer must be readable
// and the whole rootfs must extract without EOF/malformed-tar errors. This
// guards against malformed images (e.g. a broken `from: scratch` base layer)
// that Docker can still list/tag but that tools like `dive` fail to read.
func ExpectImageIsReadable(ctx context.Context, backendMode, imageName string) {
	img, cleanup := loadLocalImage(ctx, backendMode, imageName)
	defer cleanup()

	_, err := img.ConfigFile()
	Expect(err).NotTo(HaveOccurred(), "expected image config to be readable for %s", imageName)

	_, err = img.Manifest()
	Expect(err).NotTo(HaveOccurred(), "expected image manifest to be readable for %s", imageName)

	layers, err := img.Layers()
	Expect(err).NotTo(HaveOccurred(), "expected image layers to be readable for %s", imageName)

	for i, layer := range layers {
		rc, err := layer.Uncompressed()
		Expect(err).NotTo(HaveOccurred(), "expected layer %d to be readable for %s", i, imageName)
		_, err = io.Copy(io.Discard, rc)
		Expect(err).NotTo(HaveOccurred(), "expected layer %d to be fully readable for %s", i, imageName)
		Expect(rc.Close()).To(Succeed())
	}

	rc := mutate.Extract(img)
	defer rc.Close()
	_, err = io.Copy(io.Discard, rc)
	Expect(err).NotTo(HaveOccurred(), "expected image rootfs to fully extract for %s", imageName)
}

func ExpectImageHasNonEmptyLabels(ctx context.Context, backendMode, imageName string, labelKeys ...string) {
	img, cleanup := loadLocalImage(ctx, backendMode, imageName)
	defer cleanup()
	config, err := img.ConfigFile()
	Expect(err).NotTo(HaveOccurred())

	labels := config.Config.Labels
	Expect(labels).NotTo(BeNil())

	for _, key := range labelKeys {
		Expect(labels).To(HaveKey(key))
		Expect(labels[key]).NotTo(BeEmpty())
	}
}

func loadLocalImage(ctx context.Context, backendMode, imageName string) (v1.Image, func()) {
	tempFile, err := os.CreateTemp("", "werf-e2e-image-*.tar")
	Expect(err).NotTo(HaveOccurred())

	tempTarPath := tempFile.Name()
	Expect(tempFile.Close()).To(Succeed())

	switch backendMode {
	case "docker", "vanilla-docker", "buildkit-docker":
		RunSucceedCommand(ctx, "/", "docker", "save", "-o", tempTarPath, imageName)
	case "native-rootless", "native-chroot":
		RunSucceedCommand(ctx, "/", "buildah", "push", "--tls-verify=false", "--format", "docker", imageName, "docker-archive:"+tempTarPath)
	default:
		Expect(false).To(BeTrue(), "unsupported backend mode: %s", backendMode)
	}

	img, err := tarball.ImageFromPath(tempTarPath, nil)
	Expect(err).NotTo(HaveOccurred())

	return img, func() { os.Remove(tempTarPath) }
}

func GetBuiltImageLastStageImageName(ctx context.Context, testDirPath, werfBinPath, imageName string) string {
	stageImageName := SucceedCommandOutputString(ctx, testDirPath, werfBinPath, "stage", "image", imageName)

	return strings.TrimSpace(stageImageName)
}
