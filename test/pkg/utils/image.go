package utils

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	. "github.com/onsi/gomega"
)

func GetBuiltImageLastStageImageName(ctx context.Context, testDirPath, werfBinPath, imageName string) string {
	stageImageName := SucceedCommandOutputString(ctx, testDirPath, werfBinPath, "stage", "image", imageName)

	return strings.TrimSpace(stageImageName)
}

// ExtractFilesFromImage extracts specified files from a container image and saves them to the given destination directory.
// It returns a map where the keys are the original paths of the files inside the container image,
// and the values are the corresponding file paths on the host system.
//
// Example output:
//
//	map["/usr/bin/curl"] = "/tmp/img_digest[0:8]--usr_bin_curl"
func ExtractFilesFromImage(dstDir string, img v1.Image, srcContainerFilePaths []string) map[string]string {
	rc := mutate.Extract(img)
	defer rc.Close()

	imgDigest, err := img.Digest()
	Expect(err).To(Succeed())

	dstPaths := make(map[string]string, len(srcContainerFilePaths))

	tr := tar.NewReader(rc)

	for {
		hdr, err := tr.Next()

		if errors.Is(err, io.EOF) {
			break // End of archive
		} else if err != nil {
			Expect(err).To(Succeed())
		}

		if hdr.Typeflag != tar.TypeReg {
			continue // Skip non-regular files
		}

		idx := slices.IndexFunc(srcContainerFilePaths, func(containerFilePath string) bool {
			return strings.TrimPrefix(containerFilePath, "/") == hdr.Name
		})
		if idx == -1 {
			continue
		}

		tmpPath := filepath.Join(dstDir, fmt.Sprintf("%s--%s", imgDigest.Hex[0:8], strings.ReplaceAll(hdr.Name, "/", "_")))
		tmpFile, err := os.Create(tmpPath)
		Expect(err).To(Succeed())

		_, err = io.Copy(tmpFile, tr)
		Expect(err).To(Succeed())

		dstPaths[srcContainerFilePaths[idx]] = tmpFile.Name()
		Expect(tmpFile.Close()).To(Succeed())
	}

	return dstPaths
}
