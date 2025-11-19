package utils

import (
	"archive/tar"
	"context"
	"errors"
	"io"
	"os"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	. "github.com/onsi/gomega"

	elfTar "github.com/werf/werf/v2/pkg/signature/elf/tar"
)

func GetBuiltImageLastStageImageName(ctx context.Context, testDirPath, werfBinPath, imageName string) string {
	stageImageName := SucceedCommandOutputString(ctx, testDirPath, werfBinPath, "stage", "image", imageName)

	return strings.TrimSpace(stageImageName)
}

type ForEachImageFileCallback func(header *elfTar.Header, tmpPath string)

func ForEachImageFile(tmpDir string, img v1.Image, callback ForEachImageFileCallback) {
	rc := mutate.Extract(img)
	defer rc.Close()

	elfTarReader := elfTar.NewReader(tar.NewReader(rc))

	for {
		header, err := elfTarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			Expect(err).To(Succeed())
		}

		if header.Typeflag != tar.TypeReg {
			callback(header, "")
			continue
		}

		tmpFile, err := os.CreateTemp(tmpDir, "file-*.tmp")
		Expect(err).To(Succeed())

		_, err = io.Copy(tmpFile, elfTarReader)
		Expect(err).To(Succeed())

		callback(header, tmpFile.Name())
	}
}
