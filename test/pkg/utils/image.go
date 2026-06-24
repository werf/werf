package utils

import (
	"archive/tar"
	"context"
	"io"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	. "github.com/onsi/gomega"
)

func ExpectFileContentInImage(imageName, filePath, expectedContent string) {
	ref, err := name.ParseReference(imageName, name.Insecure)
	Expect(err).NotTo(HaveOccurred())

	img, err := remote.Image(ref)
	Expect(err).NotTo(HaveOccurred())

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

func GetBuiltImageLastStageImageName(ctx context.Context, testDirPath, werfBinPath, imageName string) string {
	stageImageName := SucceedCommandOutputString(ctx, testDirPath, werfBinPath, "stage", "image", imageName)

	return strings.TrimSpace(stageImageName)
}
