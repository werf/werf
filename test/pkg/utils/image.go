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
	img := loadLocalImage(ctx, backendMode, imageName)
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

func ExpectImageHasNonEmptyLabels(ctx context.Context, backendMode, imageName string, labelKeys ...string) {
	img := loadLocalImage(ctx, backendMode, imageName)
	config, err := img.ConfigFile()
	Expect(err).NotTo(HaveOccurred())

	labels := config.Config.Labels
	Expect(labels).NotTo(BeNil())

	for _, key := range labelKeys {
		Expect(labels).To(HaveKey(key))
		Expect(labels[key]).NotTo(BeEmpty())
	}
}

func loadLocalImage(ctx context.Context, backendMode, imageName string) v1.Image {
	tempFile, err := os.CreateTemp("", "werf-e2e-image-*.tar")
	Expect(err).NotTo(HaveOccurred())

	tempTarPath := tempFile.Name()
	Expect(tempFile.Close()).To(Succeed())
	defer os.Remove(tempTarPath)

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

	return img
}

func GetBuiltImageLastStageImageName(ctx context.Context, testDirPath, werfBinPath, imageName string) string {
	stageImageName := SucceedCommandOutputString(ctx, testDirPath, werfBinPath, "stage", "image", imageName)

	return strings.TrimSpace(stageImageName)
}
